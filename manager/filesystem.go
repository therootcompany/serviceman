package manager

import (
	"io"
	"os"
)

// "A little copying is better than a little dependency"
// These are here so that we don't need a dependency on http.FileSystem and http.File

// FileSystem is the same as http.FileSystem
type FileSystem interface {
	Open(name string) (File, error)
}

// File is the same as http.File
type File interface {
	io.Closer
	io.Reader
	io.Seeker
	Readdir(count int) ([]os.FileInfo, error)
	Stat() (os.FileInfo, error)
}
