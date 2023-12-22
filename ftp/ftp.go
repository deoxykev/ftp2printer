package ftp

import (
	"github.com/spf13/afero"
	"goftp.io/server/v2"
	"io"
)

type Options struct {
	Username  string
	Password  string
	Keepfiles bool
}

func StartServer(opts *Options, fs afero.Fs, onFileUpload func(file io.Reader) (int64, error)) error {
	driver, err := NewMemoryDriver(opts, fs, onFileUpload)
	if err != nil {
		return err
	}

	s, err := server.NewServer(&server.Options{
		Driver:         driver,
		WelcomeMessage: "Welcome to ftp2printer.",
		Auth: &server.SimpleAuth{
			Name:     opts.Username,
			Password: opts.Password,
		},
		Perm: server.NewSimplePerm("root", "root"),
	})
	if err != nil {
		return err
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
