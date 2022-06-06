package local

import (
	"bufio"
	"comixifier/internal/vanceai/filesystem"
	"fmt"
	"io"
	"os"
)

type File struct {
	file *os.File
}

func OpenFile(filePath string) (filesystem.File, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	return &File{file: f}, nil
}

func WrapFile(f *os.File) (filesystem.File, error) {
	return &File{file: f}, nil
}

func (f *File) Name() string {
	return f.file.Name()
}

func (f *File) Content() io.Reader {
	return bufio.NewReader(f.file)
}
