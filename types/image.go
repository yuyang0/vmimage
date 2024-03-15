package types

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/yuyang0/vmimage/utils"
)

type PullPolicy string

const (
	PullPolicyAlways       = "Always"
	PullPolicyIfNotPresent = "IfNotPresent"
	PullPolicyNever        = "Never"
)

type Image struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Tag      string `json:"tag" description:"image tag, default:latest"`
	Private  bool   `json:"private"`
	Size     int64  `json:"size"`
	Digest   string `json:"digest" description:"image digest"`
	Snapshot string `json:"snapshot" description:"image rbd snapshot"`

	ActualSize  int64
	VirtualSize int64
	LocalPath   string
}

func NewImage(fullname string) (*Image, error) {
	user, name, tag, err := utils.NormalizeImageName(fullname)
	if err != nil {
		return nil, err
	}
	return &Image{
		Username: user,
		Name:     name,
		Tag:      tag,
	}, nil
}

func (img *Image) Fullname() string {
	if img.Username == "" {
		return fmt.Sprintf("%s:%s", img.Name, img.Tag)
	} else { //nolint
		return fmt.Sprintf("%s/%s:%s", img.Username, img.Name, img.Tag)
	}
}

func (img *Image) RBDName() string {
	name := strings.ReplaceAll(img.Fullname(), "/", ".")
	return strings.ReplaceAll(name, ":", "-")
}

func (img *Image) Filepath() string {
	return img.LocalPath
}

func (img *Image) GetDigest() string {
	if img.Digest == "" {
		img.Digest, _ = utils.CalcDigestOfFile(img.LocalPath)
	}
	return img.Digest
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
