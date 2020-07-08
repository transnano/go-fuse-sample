// Copyright 2016 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This program is the analogon of libfuse's hello.c, a a program that
// exposes a single file "file.txt" in the root directory.
package main

import (
	"context"
	"flag"
	"sync"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"
)

type HelloRoot struct {
	fs.Inode

	// When file systems are mutable, all access must use
	// synchronization.
	mu      sync.Mutex
	content []byte
	mtime   time.Time
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
	r.AddChild("file.txt", ch, false)
}

func (r *HelloRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755
	return 0
}

var _ = (fs.NodeGetattrer)((*HelloRoot)(nil))
var _ = (fs.NodeOnAdder)((*HelloRoot)(nil))

// Implement handleless read.
var _ = (fs.NodeReader)((*HelloRoot)(nil))

func (bn *HelloRoot) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
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
var _ = (fs.NodeWriter)((*HelloRoot)(nil))

func (bn *HelloRoot) Write(ctx context.Context, fh fs.FileHandle, buf []byte, off int64) (uint32, syscall.Errno) {
	sz := int64(len(buf))

	logrus.Infof("%s :: %v ¥n", string(buf), off)
	return uint32(sz), 0
}

// Implement (handleless) Open
var _ = (fs.NodeOpener)((*HelloRoot)(nil))

func (f *HelloRoot) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	return nil, 0, 0
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
