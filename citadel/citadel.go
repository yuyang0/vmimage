package citadel

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/yuyang0/vmimage/types"
	"github.com/yuyang0/vmimage/utils"
	imageAPI "jihulab.com/wanjie/iaas/citadel/client/image"
	apitypes "jihulab.com/wanjie/iaas/citadel/client/types"
)

type Manager struct {
	api imageAPI.API
	cfg *types.Config
}

func NewManager(cfg *types.Config) (*Manager, error) {
	cred := &apitypes.Credential{
		Username: cfg.Citadel.Username,
		Password: cfg.Citadel.Password,
	}
	api, err := imageAPI.NewAPI(cfg.Citadel.Addr, cfg.Citadel.BaseDir, cred)
	if err != nil {
		return nil, err
	}
	return &Manager{
		api: api,
		cfg: cfg,
	}, nil
}

func (mgr *Manager) ListLocalImages(ctx context.Context, user string) ([]*types.Image, error) {
	apiImages, err := mgr.api.ListLocalImages()
	if err != nil {
		return nil, err
	}
	ans := make([]*types.Image, 0, len(apiImages))
	for _, img := range apiImages {
		ans = append(ans, &types.Image{
			Username:  img.Username,
			Name:      img.Name,
			Tag:       img.Tag,
			Private:   img.Private,
			Size:      img.Size,
			Digest:    img.Digest,
			LocalPath: img.Filepath(),
		})
	}
	return ans, nil
}

func (mgr *Manager) LoadImage(ctx context.Context, imgName string) (*types.Image, error) {
	apiImage, err := mgr.api.GetInfo(ctx, imgName)
	if err != nil {
		return nil, err
	}
	img := &types.Image{
		Username: apiImage.Username,
		Name:     apiImage.Name,
		Tag:      apiImage.Tag,
		Private:  apiImage.Private,
		Size:     apiImage.Size,
		Digest:   apiImage.Digest,
		OS: types.OSInfo{
			Type:    apiImage.OS.Type,
			Distrib: apiImage.OS.Distrib,
			Version: apiImage.OS.Version,
			Arch:    apiImage.OS.Arch,
		},
		Snapshot: apiImage.Snapshot,
	}
	return img, nil
}

func (mgr *Manager) Prepare(fname string, img *types.Image) (io.ReadCloser, error) {
	apiImage, err := mgr.api.NewImage(img.Fullname())
	if err != nil {
		return nil, err
	}
	err = apiImage.CopyFrom(fname)
	return &nullReadCloser{}, err
}

func (mgr *Manager) Pull(ctx context.Context, img *types.Image, policy types.PullPolicy) (io.ReadCloser, error) {
	newImg, err := mgr.api.Pull(ctx, img.Fullname(), imageAPI.PullPolicy(policy))
	img.Tag = newImg.Tag
	img.Snapshot = newImg.Snapshot
	img.OS = types.OSInfo{
		Type:    newImg.OS.Type,
		Distrib: newImg.OS.Distrib,
		Version: newImg.OS.Version,
		Arch:    newImg.OS.Arch,
	}

	return &nullReadCloser{}, err
}

func (mgr *Manager) Push(ctx context.Context, img *types.Image, force bool) (io.ReadCloser, error) {
	apiImage := toAPIImage(img)
	err := mgr.api.Push(ctx, apiImage, force)
	return &nullReadCloser{}, err
}

func (mgr *Manager) RemoveLocal(ctx context.Context, img *types.Image) error {
	return mgr.api.RemoveLocalImage(ctx, toAPIImage(img))
}

func (mgr *Manager) CheckHealth(ctx context.Context) error {
	u, err := url.Parse(mgr.cfg.Citadel.Addr)
	if err != nil {
		return err
	}
	hn := u.Hostname()
	if len(hn) > 0 {
		if err := utils.IPReachable(hn, time.Second); err != nil {
			return errors.Wrapf(err, "failed to ping image hub %s", hn)
		}
	}
	return nil
}

func toAPIImage(img *types.Image) *apitypes.Image {
	apiImage := &apitypes.Image{}
	apiImage.Username = img.Username
	apiImage.Name = img.Name
	apiImage.Tag = img.Tag
	apiImage.Size = img.Size
	apiImage.Digest = img.Digest
	apiImage.OS = apitypes.OSInfo{
		Type:    img.OS.Type,
		Distrib: img.OS.Distrib,
		Version: img.OS.Version,
		Arch:    img.OS.Arch,
	}
	return apiImage
}

type nullReadCloser struct{}

func (rc *nullReadCloser) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (rc *nullReadCloser) Close() error {
	return nil
}
