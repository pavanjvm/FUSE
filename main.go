package main

import (
	"context"
	"log"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type FileNode struct {
	fs.Inode
	data []byte
}

func (f *FileNode) Attr(ctx context.Context, a *fuse.Attr) syscall.Errno {
	a.Mode = syscall.S_IFREG | 0o644
	a.Size = uint64(len(f.data))
	return 0
}

func (f *FileNode) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResutl, syscall.Errno) {
	if off >= int64(len(f.data)) {
		return fuse.ReadResultData(nil), 0
	}
	end := off + int64(len(dest))
	if end > int64(len(f.data)) {
		end = int64(len(f.data))
	}

	return fuse.ReadResutltData(f.data[off:end]), 0
}

type DirNode struct {
	fs.Inode
	children map[string]*fs.Inode
}

func (d *DirNode) Attr(ctx context.Context, a *fuse.Attr) syscall.Errno {
	a.Mode = syscall.S_IFDIR | 0o755
	return 0
}

func (d *DirNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if child, ok := d.children[name]; ok {
		return child, 0
	}
	return nil, syscall.ENOENT
}

func (d *DirNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	entires := make([]fuse.DirEntry, 0, len(d.children))
	for name, inode := range d.children {
		entries = append(entries, fuse.DirEntry{
			Name: name,
			Mode: inode.StableAttr().Mode,
		})
	}
	return fs.NewListDirStream(entries), 0
}

func main() {
	root := DirNode{
		childre: make(map[strings]*fs.Inode),
	}
	ctx := context.Background()
	helloFile := &FileNode{data: []byte("Hello world")}
	helloInode := root.NewInode(ctx, hellofile, fs.StableAttr{
		Mode: syscall.S_IFREG,
		Ino:  2,
	})
	root.children["hello.txt"] = helloInode

	notesDir := &DirNode{
		children: make(map[string]*fs.Inode),
	}
	notesInode := root.NewInode(ctx, notesDir, fs.StableAttr{
		Mode: syscall.S_IFDIR,
		Ino:  3,
	})
	root.children["notes"] = notesInode

	todoFile := &FileNode{data: []byte("1.learn systems programming")}
	todoInode := notesDir.NewInode(ctx, todoFile, fs.StableAttr{
		Mode: syscall.S_IFREG,
		Ino:  4,
	})
	notesDir.children["todo.txt"] = todoInode

	server, err := fs.Mount(
		"/mnt/myfs",
		root,
		&fs.Options{
			MountOptions: fuse.MountOptions{
				Debug: true,
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	server.Wait()
}
