package fileSys

import (
	"errors"
	"io/fs"
	"os"
)

type FS interface {
	Open(fn string) (*os.File, error)
	Stat(fn string) (fs.FileInfo, error)
	IsDir(fn string) (bool, error)
	Read(fn string, offset, length int64) ([]byte, error)
}

type fileSys struct {
	fs.StatFS
}

func NewFileSys(root string) FS {
	return &fileSys{
		StatFS: os.DirFS(root).(fs.StatFS),
	}
}

func (d *fileSys) Open(fn string) (*os.File, error) {
	fp, err := d.StatFS.Open(fn)
	return fp.(*os.File), err
}

func (d *fileSys) IsDir(fn string) (bool, error) {
	info, err := d.Stat(fn)
	if nil != err {
		return false, err
	}
	return info.IsDir(), nil
}

func (d *fileSys) Read(fileName string, offset, length int64) ([]byte, error) {
	fp, err := d.Open(fileName)
	if nil != err {
		return nil, err
	}
	defer fp.Close()
	stat, err := fp.Stat()
	if nil != err {
		return nil, err
	}
	if stat.IsDir() {
		return nil, errors.New("Can not read content of a dir")
	}
	siz := stat.Size()
	if offset < 0 {
		offset = 0
	}
	if length < 0 {
		length = 0
	}
	if siz < offset {
		offset = siz
	}
	siz -= offset
	if siz < length || length < 1 {
		length = siz
	}
	buf := make([]byte, length)
	_, err = fp.ReadAt(buf, offset)
	return buf, err
}
