package filesystem

//go:generate mockgen --build_flags=--mod=mod -destination filesystem_mock.go -package filesystem integration-vanceai/internal/filesystem File

import "io"

type File interface {
	Name() string
	Content() io.Reader
}
