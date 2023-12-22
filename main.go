package main

import (
	"fmt"
	"github.com/deoxykev/ftp2printer/m/v2/ftp"
	"github.com/spf13/afero"
	"goftp.io/server/v2"
	"log"
	"os"
)

func StartServer(fs afero.Fs) {
	fmt.Println("hello world")
	driver, err := ftp.NewDriver(fs)
	if err != nil {
		log.Fatal(err)
	}

	s, err := server.NewServer(&server.Options{
		Driver:         driver,
		WelcomeMessage: "Welcome to ftp2printer.",
		Auth: &server.SimpleAuth{
			Name:     "admin",
			Password: "admin",
		},
		Perm: server.NewSimplePerm("root", "root"),
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type CustomFile struct {
	afero.File
	onWrite func(p []byte) (int, error)
	onClose func() error
}

func (cf *CustomFile) Write(p []byte) (int, error) {
	// Call the callback function
	if cf.onWrite != nil {
		return cf.onWrite(p)
	}
	// Default write operation
	return cf.File.Write(p)
}

func (cf *CustomFile) Close() error {
	// Call the callback function
	if cf.onClose != nil {
		cf.onClose()
	}
	// Default close operation
	return cf.File.Close()
}

type CustomFs struct {
	afero.Fs
	onFileWrite func(p []byte) (int, error)
	onFileClose func() error
}

func (cfs *CustomFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	file, err := cfs.Fs.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return &CustomFile{
		File:    file,
		onWrite: cfs.onFileWrite,
		onClose: cfs.onFileClose,
	}, nil
}

func main() {
	fs := &CustomFs{
		Fs: afero.NewMemMapFs(),
		onFileWrite: func(p []byte) (int, error) {
			// Your custom logic for write callback
			println("Write operation performed")
			return len(p), nil
		},
		onFileClose: func() error {
			// Your custom logic for close callback
			println("File closed")
			return nil
		},
	}

	StartServer(fs)

}
