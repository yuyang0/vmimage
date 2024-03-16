package types

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

type DockerConfig struct {
	Endpoint string `toml:"endpoint" default:"unix:///var/run/docker.sock"`
	Auth     string `toml:"-"` // in base64
	Prefix   string `toml:"prefix"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type CitadelConfig struct {
	BaseDir  string `toml:"base_dir"`
	Addr     string `toml:"addr"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

type Config struct {
	Type    string        `toml:"type" default:"docker"`
	Docker  DockerConfig  `toml:"docker"`
	Citadel CitadelConfig `toml:"citadel"`
}

func (cfg *Config) CheckAndRefine() error {
	switch cfg.Type {
	case "docker":
		if cfg.Docker.Username == "" || cfg.Docker.Password == "" {
			return errors.New("docker's username or password should not be empty")
		}
		auth := map[string]string{
			"username": cfg.Docker.Username,
			"password": cfg.Docker.Password,
		}
		authBytes, _ := json.Marshal(auth)
		cfg.Docker.Auth = base64.StdEncoding.EncodeToString(authBytes)
	case "citadel":
		if cfg.Citadel.Username == "" || cfg.Citadel.Password == "" {
			return errors.New("ImageHub's username or password should not be empty")
		}
		if cfg.Citadel.Addr == "" {
			return errors.New("ImageHub's address shouldn't be empty")
		}
	case "mock":
		return nil
	default:
		return errors.New("unknown image hub type")
	}
	return nil
}
