package ftp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"goftp.io/server/v2"
)

// MemoryDriver implements MemoryDriver (in-memory only)
type MemoryDriver struct {
	RootPath     string
	fs           afero.Fs
	onFileUpload func(file io.Reader) (int64, error)
	opts         *Options
}

// NewMemoryDriver implements Driver
func NewMemoryDriver(opts *Options, fs afero.Fs, onFileUpload func(file io.Reader) (int64, error)) (server.Driver, error) {
	rootPath := "/"
	var err error
	rootPath, err = filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	return &MemoryDriver{rootPath, fs, onFileUpload, opts}, nil
}

func (driver *MemoryDriver) realPath(path string) string {
	paths := strings.Split(path, "/")
	return filepath.Join(append([]string{driver.RootPath}, paths...)...)
}

// Stat implements Driver
func (driver *MemoryDriver) Stat(ctx *server.Context, path string) (os.FileInfo, error) {
	basepath := driver.realPath(path)
	rPath, err := filepath.Abs(basepath)
	if err != nil {
		return nil, err
	}

	return driver.fs.Stat(rPath)
	//return os.Lstat(rPath)
}

// ListDir implements the Driver interface to list directories
func (driver *MemoryDriver) ListDir(ctx *server.Context, path string, callback func(os.FileInfo) error) error {
	basepath := driver.realPath(path)

	return afero.Walk(driver.fs, basepath, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// The relative path calculation needs to be adjusted to prevent incorrect skips
		rPath, _ := filepath.Rel(basepath, f)
		if rPath == "." || rPath == info.Name() {
			return callback(info)
		}
		return nil
	})
}

// DeleteDir implements Driver
func (driver *MemoryDriver) DeleteDir(ctx *server.Context, path string) error {
	rPath := driver.realPath(path)
	f, err := driver.fs.Stat(rPath)
	if err != nil {
		return err
	}
	if f.IsDir() {
		return driver.fs.RemoveAll(rPath)
		//return os.RemoveAll(rPath)
	}
	return errors.New("Not a directory")
}

// DeleteFile implements Driver
func (driver *MemoryDriver) DeleteFile(ctx *server.Context, path string) error {
	rPath := driver.realPath(path)
	f, err := driver.fs.Stat(rPath)
	if err != nil {
		return err
	}
	if !f.IsDir() {
		//return os.Remove(rPath)
		return driver.fs.Remove(rPath)
	}
	return errors.New("Not a file")
}

// Rename implements Driver
func (driver *MemoryDriver) Rename(ctx *server.Context, fromPath string, toPath string) error {
	oldPath := driver.realPath(fromPath)
	newPath := driver.realPath(toPath)
	//return os.Rename(oldPath, newPath)
	return driver.fs.Rename(oldPath, newPath)
}

// MakeDir implements Driver
func (driver *MemoryDriver) MakeDir(ctx *server.Context, path string) error {
	rPath := driver.realPath(path)
	//return os.MkdirAll(rPath, os.ModePerm)
	return driver.fs.MkdirAll(rPath, os.ModePerm)
}

// GetFile implements Driver
func (driver *MemoryDriver) GetFile(ctx *server.Context, path string, offset int64) (int64, io.ReadCloser, error) {
	rPath := driver.realPath(path)
	//f, err := os.Open(rPath)
	f, err := driver.fs.Open(rPath)
	if err != nil {
		return 0, nil, err
	}
	defer func() {
		if err != nil && f != nil {
			f.Close()
		}
	}()

	info, err := f.Stat()
	if err != nil {
		return 0, nil, err
	}

	_, err = f.Seek(offset, io.SeekStart)
	if err != nil {
		return 0, nil, err
	}

	return info.Size() - offset, f, nil
}

/*
func (driver *Driver) PutFile(ctx *server.Context, destPath string, data io.Reader, offset int64) (int64, error) {
	return driver.onFileUpload(data)
}
*/

// PutFile implements Driver
func (driver *MemoryDriver) PutFile(ctx *server.Context, destPath string, reader io.Reader, offset int64) (int64, error) {
	data, pw := io.Pipe()
	tee := io.TeeReader(reader, pw)
	go func() {
		defer pw.Close()
		n, err := driver.onFileUpload(tee)
		if err != nil {
			log.Printf("onFileUpload callback error: %s\n", err)
		}
		log.Printf("onFileUpload callback wrote %d bytes\n", n)
	}()

	if !driver.opts.Keepfiles {
		return 0, nil
	}

	rPath := driver.realPath(destPath)
	var isExist bool
	//f, err := os.Lstat(rPath)
	f, err := driver.fs.Stat(rPath)
	if err == nil {
		isExist = true
		if f.IsDir() {
			return 0, errors.New("A dir has the same name")
		}
	}

	if offset > -1 && !isExist {
		offset = -1
	}

	if offset == -1 {
		if isExist {
			err = driver.fs.Remove(rPath)
			if err != nil {
				return 0, err
			}
		}
		f, err := driver.fs.Create(rPath)
		if err != nil {
			return 0, err
		}
		defer f.Close()

		n, err := io.Copy(f, data)
		if err != nil {
			return 0, err
		}
		return n, nil
	}

	of, err := driver.fs.OpenFile(rPath, os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		return 0, err
	}
	defer of.Close()

	info, err := of.Stat()
	if err != nil {
		return 0, err
	}
	if offset > info.Size() {
		return 0, fmt.Errorf("Offset %d is beyond file size %d", offset, info.Size())
	}

	_, err = of.Seek(offset, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	bytes, err := io.Copy(of, data)
	if err != nil {
		return 0, err
	}

	return bytes, nil
}
