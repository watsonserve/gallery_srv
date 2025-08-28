package fileSys

import (
	"context"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/net/webdav"
)

type FileSystem struct {
	Root string
}

func (fs *FileSystem) AbsPathName(name string) string {
	return filepath.Join(fs.Root, path.Clean("/"+name))
}

func (fs *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return os.Mkdir(fs.AbsPathName(name), perm&0770)
}

func (fs *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return os.OpenFile(fs.AbsPathName(name), flag, perm&0660)
}

func (fs *FileSystem) RemoveAll(ctx context.Context, name string) error {
	return os.RemoveAll(fs.AbsPathName(name))
}

func (fs *FileSystem) Rename(ctx context.Context, oldName, newName string) error {
	return os.Rename(fs.AbsPathName(oldName), fs.AbsPathName(newName))
}

func (fs *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return os.Stat(fs.AbsPathName(name))
}
