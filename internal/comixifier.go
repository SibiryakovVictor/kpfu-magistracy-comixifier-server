package internal

import "io"

type Comixifier interface {
	Do(imgData io.Reader) (io.Reader, error)
}
