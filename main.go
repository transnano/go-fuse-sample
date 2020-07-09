// Copyright 2016 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This program is the analogon of libfuse's hello.c, a a program that
// exposes a single file "file.txt" in the root directory.
package main

import (
	"context"
	"flag"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"
)

type HelloRoot struct {
	fs.Inode
}

func (r *HelloRoot) OnAdd(ctx context.Context) {
	ch := r.NewPersistentInode(
		ctx, &fs.MemRegularFile{
			Data: []byte("foobar"),
			Attr: fuse.Attr{
				Size:      100,
				Mode:      0644,
				Atime:     uint64(time.Now().Unix()),
				Atimensec: uint32(time.Now().Nanosecond()),
				Ctime:     uint64(time.Now().Unix()),
				Ctimensec: uint32(time.Now().Nanosecond()),
				Mtime:     uint64(time.Now().Unix()),
				Mtimensec: uint32(time.Now().Nanosecond()),
			},
		}, fs.StableAttr{Ino: 2})
	r.AddChild("read-only-file.txt", ch, false)

	node := r.NewPersistentInode(
		ctx, &bytesNode{},
		fs.StableAttr{
			Mode: syscall.S_IFREG,
			// Make debug output readable.
			Ino: 2,
		})
	r.AddChild("writable-file.txt", node, true)

}

func (r *HelloRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755
	logrus.Debugf("Getattr: %v - %v\n", fh, r)
	return 0
}

func (r *HelloRoot) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	logrus.Debugf("Lookup: %s - %v\n", name, r)
	child := &fs.MemRegularFile{}
	childNode := r.NewInode(ctx, child, fs.StableAttr{Mode: syscall.S_IFREG})
	return childNode, fs.OK
}

var _ = (fs.NodeGetattrer)((*HelloRoot)(nil))
var _ = (fs.NodeOnAdder)((*HelloRoot)(nil))
var _ = (fs.NodeLookuper)((*HelloRoot)(nil))

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
	bn.mu.Lock()
	defer bn.mu.Unlock()
	logrus.Debugf("Getattr: %v - %v\n", fh, bn)
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
	logrus.Debugf("Setattr: %v - %v\n", fh, bn)

	if sz, ok := in.GetSize(); ok {
		bn.resize(sz)
	}
	bn.getattr(out)
	return 0
}

// Implement handleless read.
var _ = (fs.NodeReader)((*bytesNode)(nil))

func (bn *bytesNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	bn.mu.Lock()
	defer bn.mu.Unlock()
	logrus.Debugf("Read: %v - %v\n", fh, bn)

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
	sz := int64(len(buf))

	logrus.Infof("Write: %s :: %v\n", string(buf), off)
	return uint32(sz), 0
}

// Implement (handleless) Open
var _ = (fs.NodeOpener)((*bytesNode)(nil))

func (f *bytesNode) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
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
	opts := &fs.Options{}
	opts.Debug = *debug
	server, err := fs.Mount(flag.Arg(0), &HelloRoot{}, opts)
	if err != nil {
		logrus.Fatalf("Mount fail: %v\n", err)
	}
	server.Wait()
}
