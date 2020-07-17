// Copyright 2019 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"
)

// bytesNode is a file that can be read and written
type bytesNode struct {
	fs.Inode

	// When file systems are mutable, all access must use
	// synchronization.
	mu      sync.Mutex
	content []byte
	mtime   time.Time
}

// Implement GetAttr to provide size and mtime
var _ = (fs.NodeGetattrer)((*bytesNode)(nil))

func (bn *bytesNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	logrus.Infof("Getattr: %v", fh)
	bn.mu.Lock()
	defer bn.mu.Unlock()
	bn.getattr(out)
	return 0
}

func (bn *bytesNode) getattr(out *fuse.AttrOut) {
	out.Size = uint64(len(bn.content))
	out.SetTimes(nil, &bn.mtime, nil)
}

func (bn *bytesNode) resize(sz uint64) {
	if sz > uint64(cap(bn.content)) {
		n := make([]byte, sz)
		copy(n, bn.content)
		bn.content = n
	} else {
		bn.content = bn.content[:sz]
	}
	bn.mtime = time.Now()
}

// Implement Setattr to support truncation
var _ = (fs.NodeSetattrer)((*bytesNode)(nil))

func (bn *bytesNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	bn.mu.Lock()
	defer bn.mu.Unlock()

	if sz, ok := in.GetSize(); ok {
		bn.resize(sz)
	}
	bn.getattr(out)
	return 0
}

// Implement handleless read.
var _ = (fs.NodeReader)((*bytesNode)(nil))

func (bn *bytesNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	logrus.Infof("Read: %s, %d", string(dest), off)
	bn.mu.Lock()
	defer bn.mu.Unlock()

	end := off + int64(len(dest))
	if end > int64(len(bn.content)) {
		end = int64(len(bn.content))
	}

	// We could copy to the `dest` buffer, but since we have a
	// []byte already, return that.
	return fuse.ReadResultData(bn.content[off:end]), 0
}

// Implement handleless write.
var _ = (fs.NodeWriter)((*bytesNode)(nil))

func (bn *bytesNode) Write(ctx context.Context, fh fs.FileHandle, buf []byte, off int64) (uint32, syscall.Errno) {
	logrus.Infof("Write: %s, %d", string(buf), off)
	bn.mu.Lock()
	defer bn.mu.Unlock()

	sz := int64(len(buf))
	if off+sz > int64(len(bn.content)) {
		bn.resize(uint64(off + sz))
	}
	copy(bn.content[off:], buf)
	bn.mtime = time.Now()
	return uint32(sz), 0
}

// Implement (handleless) Open
var _ = (fs.NodeOpener)((*bytesNode)(nil))

func (f *bytesNode) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	logrus.Infof("Open: %v", openFlags)
	return nil, 0, 0
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	// logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	debug := flag.Bool("debug", false, "print debug data")
	flag.Parse()
	if len(flag.Args()) < 1 {
		logrus.Fatal("Usage:\n  hello MOUNTPOINT")
	}
	root := &fs.Inode{}

	// Mount the file system
	server, err := fs.Mount(flag.Arg(0), root, &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug:      *debug,
			AllowOther: true,
		},

		// Setup the file.
		OnAdd: func(ctx context.Context) {
			ch := root.NewPersistentInode(
				ctx,
				&bytesNode{},
				fs.StableAttr{
					Mode: syscall.S_IFREG,
					// Make debug output readable.
					Ino: 2,
				})
			root.AddChild("bytes", ch, true)
		},
	})
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Printf(`Try:

  sudo ls -l %s/bytes
  sudo sh -c "echo 'hello' > %s/bytes"
  sudo ls -l %s/bytes
  sudo cat %s/bytes

`, flag.Arg(0))
	fmt.Printf("Unmount by calling 'sudo fusermount -u %s'\n", flag.Arg(0))

	// Serve the file system, until unmounted by calling fusermount -u
	server.Wait()
}
