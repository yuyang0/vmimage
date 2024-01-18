package vmimage

import (
	"context"
	"io"
)

type Image interface { //nolint:interfacebloat
	Prepare(fname string) (io.ReadCloser, error)
	Pull(ctx context.Context) (io.ReadCloser, error)
	Push(ctx context.Context, force bool) (io.ReadCloser, error)
	RemoveLocal(ctx context.Context) error
	LoadMetadata(ctx context.Context) (err error)

	Fullname() string
	Filepath() string
	RBDName() string
	VirtualSize() int64
	Distro() string
	Digest() string
}

type Manager interface {
	ListLocalImages(ctx context.Context, user string) ([]Image, error)
	NewImage(fullname string) (Image, error) // create image object, but don't pull
	LoadImage(imgName string) (Image, error) // create image object and pull the image to local
}
