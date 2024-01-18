package factory

import (
	"context"
	"fmt"

	"github.com/alphadose/haxmap"
	"github.com/yuyang0/vmimage"
	"github.com/yuyang0/vmimage/docker"
	"github.com/yuyang0/vmimage/mocks"
	"github.com/yuyang0/vmimage/types"
)

const (
	dockerType = "docker"
	mockType   = "mock"
)

var (
	gF *Factory
)

func Setup(config *types.Config) (err error) {
	gF, err = NewFactory(config)
	return err
}

type Factory struct {
	cfg    *types.Config
	mgrMap *haxmap.Map[string, vmimage.Manager]
}

func NewFactory(cfg *types.Config) (f *Factory, err error) {
	f = &Factory{
		cfg:    cfg,
		mgrMap: haxmap.New[string, vmimage.Manager](),
	}
	f.mgrMap.Set(mockType, &mocks.Manager{})

	var mgr vmimage.Manager
	switch cfg.Type {
	case dockerType:
		mgr, err = docker.NewManager(cfg)
	default:
		err = fmt.Errorf("invalid type: %s", cfg.Type)
	}
	if err != nil {
		return nil, err
	}
	f.mgrMap.Set(cfg.Type, mgr)
	return f, nil
}

func (f *Factory) GetManager(ty string) (mgr vmimage.Manager, err error) {
	if ty == "" {
		ty = f.cfg.Type
	}
	if mgr, _ = f.mgrMap.Get(ty); mgr != nil {
		return mgr, nil
	}
	switch ty {
	case dockerType:
		mgr, err = docker.NewManager(f.cfg)
	case mockType:
		mgr = &mocks.Manager{}
	default:
		return nil, fmt.Errorf("invalid image manager type: %s", ty)
	}
	f.mgrMap.Set(ty, mgr)
	return mgr, err
}

func GetManager(tys ...string) (vmimage.Manager, error) {
	ty := ""
	if len(tys) > 0 {
		ty = tys[0]
	}
	return gF.GetManager(ty)
}

func LoadImage(imgName string) (img vmimage.Image, err error) {
	mgr, err := GetManager()
	if err != nil {
		return nil, err
	}
	return mgr.LoadImage(imgName)
}

func ListLocalImages(ctx context.Context, user string) ([]vmimage.Image, error) {
	mgr, err := GetManager()
	if err != nil {
		return nil, err
	}
	return mgr.ListLocalImages(ctx, user)
}

func NewImage(imgName string) (vmimage.Image, error) {
	mgr, err := GetManager()
	if err != nil {
		return nil, err
	}
	return mgr.NewImage(imgName)
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
