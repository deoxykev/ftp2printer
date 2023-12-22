package main

import (
	"github.com/deoxykev/ftp2printer/m/v2/ftp"
	"github.com/spf13/afero"
	"io"
	"os"
)

func main() {
	fs := afero.NewMemMapFs()

	opts := &ftp.Options{
		Username:  "admin",
		Password:  "admin",
		Keepfiles: false,
	}

	ftp.StartServer(opts, fs, func(file io.Reader) (int64, error) {
		return io.Copy(os.Stdout, file)
	})
}
