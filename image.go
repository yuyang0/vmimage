package vmimage

import (
	"context"
	"fmt"
	"io"

	"github.com/yuyang0/vmimage/docker"
	"github.com/yuyang0/vmimage/mocks"
	"github.com/yuyang0/vmimage/types"
	"github.com/yuyang0/vmimage/utils"
)

const (
	dockerType = "docker"
)

var (
	m *Manager
)

func Setup(config *types.ImageHubConfig) (err error) {
	m = &Manager{
		cfg: config,
	}
	switch config.Type {
	case dockerType:
		err = docker.Setup(config)
	default:
		err = fmt.Errorf("invalid image hub type: %s", config.Type)
	}
	return err
}

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

type Manager struct {
	cfg *types.ImageHubConfig
}

// Load creates object, it don't pull image
func (mgr *Manager) Load(imgName string) (img Image, err error) {
	cfg := mgr.cfg
	switch cfg.Type {
	case dockerType:
		img, err = docker.New(imgName)
	default:
		err = fmt.Errorf("invalid image hub type: %s", cfg.Type)
	}
	if err != nil {
		return nil, err
	}
	rc, err := img.Pull(context.TODO())
	if err != nil {
		return nil, err
	}
	utils.EnsureReaderClosed(rc)
	if err := img.LoadMetadata(context.TODO()); err != nil {
		return nil, err
	}
	return img, nil
}

// ListLocalImages lists all local images
func (mgr *Manager) ListLocalImages(ctx context.Context, user string) ([]Image, error) {
	cfg := mgr.cfg
	switch cfg.Type {
	case dockerType:
		imgs, err := docker.ListLocalImages(ctx, user)
		if err != nil {
			return nil, err
		}
		ans := make([]Image, 0, len(imgs))
		for _, img := range imgs {
			ans = append(ans, img)
		}
		return ans, nil
	default:
		return nil, fmt.Errorf("invalid image hub type: %s", cfg.Type)
	}
}

func (mgr *Manager) NewImage(imgName string) (Image, error) {
	cfg := mgr.cfg
	switch cfg.Type {
	case dockerType:
		return docker.New(imgName)
	default:
		return nil, fmt.Errorf("invalid image hub type: %s", cfg.Type)
	}
}

func Load(imgName string) (img Image, err error) {
	return m.Load(imgName)
}

func ListLocalImages(ctx context.Context, user string) ([]Image, error) {
	return m.ListLocalImages(ctx, user)
}

func NewImage(imgName string) (Image, error) {
	return m.NewImage(imgName)
}

func NewImageName(user, name string) string {
	imgName := fmt.Sprintf("%s/%s", user, name)
	if user == "" {
		imgName = name
	}
	return imgName
}

func NewMockImage() *mocks.Image {
	return &mocks.Image{}
}
