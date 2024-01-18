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
	"github.com/yuyang0/vmimage"
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

func (m *Manager) ListLocalImages(ctx context.Context, user string) ([]vmimage.Image, error) {
	images, err := m.cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	var ans []vmimage.Image
	prefix := path.Join(m.cfg.Prefix, user)
	for _, dockerImg := range images {
		for _, repoTag := range dockerImg.RepoTags {
			if strings.HasPrefix(repoTag, prefix) {
				fullname := strings.TrimPrefix(repoTag, prefix)
				fullname = strings.TrimPrefix(fullname, "/")
				fullname = strings.TrimPrefix(fullname, "library/")
				img, _ := m.NewImage(fullname)
				ans = append(ans, img)
			}
		}
	}
	return ans, nil
}

func (m *Manager) NewImage(fullname string) (vmimage.Image, error) {
	user, name, tag, err := utils.NormalizeImageName(fullname)
	if err != nil {
		return nil, err
	}
	return &Image{
		user: user,
		name: name,
		tag:  tag,

		cfg: m.cfg,
		cli: m.cli,
	}, nil
}

func (m *Manager) LoadImage(imgName string) (img vmimage.Image, err error) {
	if img, err = m.NewImage(imgName); err != nil {
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

type Image struct {
	user        string
	name        string
	tag         string
	distro      string
	digest      string
	actualSize  int64
	virtualSize int64
	localPath   string

	cfg *pkgtypes.Config
	cli *engineapi.Client
}

// Prepare prepares the image for use by creating a Dockerfile and building a Docker image.
//
// Parameters:
//   - fname: a local filename or an url
//
// Returns:
//   - io.ReadCloser: a ReadCloser to read the prepared image.
//   - error: an error if any occurred during the preparation process.
func (img *Image) Prepare(fname string) (io.ReadCloser, error) {
	cli := img.cli
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
		Tags:       []string{img.dockerImageName()},
	}

	resp, err := cli.ImageBuild(context.Background(), buildContext, buildOptions)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (img *Image) Pull(ctx context.Context) (io.ReadCloser, error) {
	cli, cfg := img.cli, img.cfg
	return cli.ImagePull(ctx, img.dockerImageName(), types.ImagePullOptions{
		RegistryAuth: cfg.Docker.Auth,
	})
}

func (img *Image) Push(ctx context.Context, force bool) (io.ReadCloser, error) {
	cli, cfg := img.cli, img.cfg
	return cli.ImagePush(ctx, img.dockerImageName(), types.ImagePushOptions{
		RegistryAuth: cfg.Docker.Auth,
		All:          force,
	})
}

func (img *Image) LoadMetadata(ctx context.Context) (err error) {
	cli := img.cli
	resp, _, err := cli.ImageInspectWithRaw(ctx, img.dockerImageName())
	if err != nil {
		return err
	}
	upperDir := resp.GraphDriver.Data["UpperDir"]
	img.localPath = filepath.Join(upperDir, destImgName)
	img.actualSize, img.virtualSize, err = utils.ImageSize(ctx, img.localPath)

	img.digest = resp.Config.Labels["SHA256"]
	return err
}

func (img *Image) RemoveLocal(ctx context.Context) error {
	cli := img.cli
	_, err := cli.ImageRemove(ctx, img.dockerImageName(), types.ImageRemoveOptions{
		Force:         true, // Remove even if the image is in use
		PruneChildren: true, // Prune dependent child images
	})
	return err
}

func (img *Image) Fullname() string {
	if img.user == "" {
		return fmt.Sprintf("%s:%s", img.name, img.tag)
	} else { //nolint
		return fmt.Sprintf("%s/%s:%s", img.user, img.name, img.tag)
	}
}

func (img *Image) RBDName() string {
	name := strings.ReplaceAll(img.Fullname(), "/", ".")
	return strings.ReplaceAll(name, ":", "-")
}

func (img *Image) Filepath() string {
	return img.localPath
}

func (img *Image) VirtualSize() int64 {
	return img.virtualSize
}

func (img *Image) Distro() string {
	return img.distro
}

func (img *Image) Digest() string {
	if img.digest == "" {
		img.digest, _ = utils.CalcDigestOfFile(img.localPath)
	}
	return img.digest
}

func (img *Image) dockerImageName() string {
	cfg := img.cfg
	if img.user == "" {
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
