package vmimage

import (
	"context"
	"io"

	"github.com/yuyang0/vmimage/types"
)

type Manager interface {
	ListLocalImages(ctx context.Context, user string) ([]*types.Image, error)
	LoadImage(imgName string) (*types.Image, error) // create image object and pull the image to local

	Prepare(fname string, img *types.Image) (io.ReadCloser, error)
	Pull(ctx context.Context, img *types.Image) (io.ReadCloser, error)
	Push(ctx context.Context, img *types.Image, force bool) (io.ReadCloser, error)
	RemoveLocal(ctx context.Context, img *types.Image) error
}
