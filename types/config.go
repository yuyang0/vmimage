package types

type DockerConfig struct {
	Endpoint string `toml:"endpoint" default:"unix:///var/run/docker.sock"`
	Auth     string `toml:"-"` // in base64
}

type CitadelConfig struct {
	BaseDir string `toml:"base_dir"`
	Addr    string `toml:"addr"`
}

type Config struct {
	Type    string        `toml:"type" default:"docker"`
	Docker  DockerConfig  `toml:"docker"`
	Citadel CitadelConfig `toml:"citadel"`

	// config for registry
	Prefix     string `toml:"prefix"`
	Username   string `toml:"username"`
	Password   string `toml:"password"`
	PullPolicy string `toml:"pull_policy"`
}
