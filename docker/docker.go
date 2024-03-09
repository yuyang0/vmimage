package docker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	engineapi "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	pkgtypes "github.com/yuyang0/vmimage/types"
	"github.com/yuyang0/vmimage/utils"
)

const (
	destImgName      = "vm.img"
	dockerCliVersion = "1.35"
)

type Manager struct {
	cfg *pkgtypes.Config
	cli *engineapi.Client
}

func NewManager(config *pkgtypes.Config) (m *Manager, err error) {
	cli, err := makeDockerClient(config.Docker.Endpoint)
	if err != nil {
		return nil, err
	}
	m = &Manager{
		cfg: config,
		cli: cli,
	}
	return m, nil
}

func (m *Manager) ListLocalImages(ctx context.Context, user string) ([]*pkgtypes.Image, error) {
	images, err := m.cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	var ans []*pkgtypes.Image
	prefix := path.Join(m.cfg.Prefix, user)
	for _, dockerImg := range images {
		for _, repoTag := range dockerImg.RepoTags {
			if strings.HasPrefix(repoTag, prefix) {
				fullname := strings.TrimPrefix(repoTag, prefix)
				fullname = strings.TrimPrefix(fullname, "/")
				fullname = strings.TrimPrefix(fullname, "library/")
				img, _ := pkgtypes.NewImage(fullname)
				ans = append(ans, img)
			}
		}
	}
	return ans, nil
}

func (m *Manager) LoadImage(imgName string) (img *pkgtypes.Image, err error) {
	if img, err = pkgtypes.NewImage(imgName); err != nil {
		return nil, err
	}
	rc, err := m.Pull(context.TODO(), img)
	if err != nil {
		return nil, err
	}
	utils.EnsureReaderClosed(rc)
	if err := m.loadMetadata(context.TODO(), img); err != nil {
		return nil, err
	}
	return img, nil
}

// Prepare prepares the image for use by creating a Dockerfile and building a Docker image.
//
// Parameters:
//   - fname: a local filename or an url
//
// Returns:
//   - io.ReadCloser: a ReadCloser to read the prepared image.
//   - error: an error if any occurred during the preparation process.
func (mgr *Manager) Prepare(fname string, img *pkgtypes.Image) (io.ReadCloser, error) {
	cli := mgr.cli
	baseDir := filepath.Dir(fname)
	baseName := filepath.Base(fname)
	digest := ""
	tarOpts := &archive.TarOptions{
		IncludeFiles: []string{baseName, "Dockerfile.yavirt"},
		Compression:  archive.Uncompressed,
		NoLchown:     true,
	}
	if u, err := url.Parse(fname); err == nil && u.Scheme != "" && u.Host != "" {
		tmpDir, err := os.MkdirTemp(os.TempDir(), "image-prepare-")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(tmpDir)
		baseDir = tmpDir
		baseName = fname
		tarOpts.IncludeFiles = []string{"Dockerfile.yavirt"}
		if digest, err = httpGetSHA256(fname); err != nil {
			return nil, err
		}
	} else {
		if digest, err = utils.CalcDigestOfFile(fname); err != nil {
			return nil, err
		}
	}
	dockerfile := fmt.Sprintf("FROM scratch\nLABEL SHA256=%s\nADD %s /%s", digest, baseName, destImgName)
	if err := os.WriteFile(filepath.Join(baseDir, "Dockerfile.yavirt"), []byte(dockerfile), 0600); err != nil {
		return nil, err
	}
	defer os.Remove(filepath.Join(baseDir, "Dockerfile.yavirt"))

	// Create a build context from the specified directory
	buildContext, err := archive.TarWithOptions(baseDir, tarOpts)
	if err != nil {
		return nil, err
	}

	// Build the Docker image using the build context
	buildOptions := types.ImageBuildOptions{
		Context:    buildContext,
		Dockerfile: "Dockerfile.yavirt", // Use the default Dockerfile name
		Tags:       []string{mgr.dockerImageName(img)},
	}

	resp, err := cli.ImageBuild(context.Background(), buildContext, buildOptions)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (mgr *Manager) Pull(ctx context.Context, img *pkgtypes.Image) (io.ReadCloser, error) {
	cli, cfg := mgr.cli, mgr.cfg
	return cli.ImagePull(ctx, mgr.dockerImageName(img), types.ImagePullOptions{
		RegistryAuth: cfg.Docker.Auth,
	})
}

func (mgr *Manager) Push(ctx context.Context, img *pkgtypes.Image, force bool) (io.ReadCloser, error) {
	cli, cfg := mgr.cli, mgr.cfg
	return cli.ImagePush(ctx, mgr.dockerImageName(img), types.ImagePushOptions{
		RegistryAuth: cfg.Docker.Auth,
		All:          force,
	})
}

func (mgr *Manager) RemoveLocal(ctx context.Context, img *pkgtypes.Image) error {
	cli := mgr.cli
	_, err := cli.ImageRemove(ctx, mgr.dockerImageName(img), types.ImageRemoveOptions{
		Force:         true, // Remove even if the image is in use
		PruneChildren: true, // Prune dependent child images
	})
	return err
}

func (mgr *Manager) loadMetadata(ctx context.Context, img *pkgtypes.Image) (err error) {
	cli := mgr.cli
	resp, _, err := cli.ImageInspectWithRaw(ctx, mgr.dockerImageName(img))
	if err != nil {
		return err
	}
	upperDir := resp.GraphDriver.Data["UpperDir"]
	img.LocalPath = filepath.Join(upperDir, destImgName)
	img.ActualSize, img.VirtualSize, err = utils.ImageSize(ctx, img.LocalPath)

	img.Digest = resp.Config.Labels["SHA256"]
	return err
}

func (m *Manager) dockerImageName(img *pkgtypes.Image) string {
	cfg := m.cfg
	if img.User == "" {
		return path.Join(cfg.Prefix, "library", img.Fullname())
	} else { //nolint
		return path.Join(cfg.Prefix, img.Fullname())
	}
}

func makeDockerClient(endpoint string) (*engineapi.Client, error) {
	defaultHeaders := map[string]string{"User-Agent": "eru-yavirt"}
	return engineapi.NewClient(endpoint, dockerCliVersion, nil, defaultHeaders)
}

func httpGetSHA256(u string) (string, error) {
	if !strings.HasSuffix(u, ".img") {
		return "", fmt.Errorf("invalid url: %s", u)
	}
	url := strings.TrimSuffix(u, ".img")
	url += ".sha256sum"
	// Perform GET request
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
