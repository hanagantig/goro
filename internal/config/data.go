package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
)

type DependencyName string

type Config struct {
	App      App      `yaml:"app"`
	UseCase  UseCase  `yaml:"use_case"`
	Storages Storages `yaml:"storages"`
	Services Services `yaml:"services"`
	Adapters Adapters `yaml:"adapters"`
	Chunks   []Chunk

	// Path is the file NewConfig read the config from, as it was passed; empty for a config
	// that wasn't read from a file. A relative work_dir counts from its directory, and init
	// copies goro.yaml from it into the project root (ADR 0002).
	Path string `yaml:"-"`
}

type Chunk struct {
	Name              string
	Scope             string
	ArgName           string
	ReturnType        string
	DefinitionImports string
	BuildImports      string
	InitFunc          string
	Build             string
	Configs           string
	InitConfig        string
	InitHasErr        bool
}

type App struct {
	Name    string `yaml:"name"`
	Module  string `yaml:"module"`
	WorkDir string `yaml:"work_dir"`
}

type Dependency struct {
	Pkg       string   `yaml:"pkg"`
	Type      string   `yaml:"type"`
	BuildFunc string   `yaml:"build_func"`
	Methods   []string `yaml:"methods"`
	Deps      []string `yaml:"deps"`
}

func (d Dependency) GetPackageName() string {
	path := strings.Split(d.Pkg, "/")
	if len(path) > 0 {
		return path[len(path)-1]
	}

	return ""
}

// NewConfig reads the goro.yaml at pathToFile. An empty path gives an empty config.
func NewConfig(pathToFile string) (Config, error) {
	if pathToFile == "" {
		return Config{}, nil
	}

	data, err := os.ReadFile(pathToFile)
	if err != nil {
		return Config{}, err
	}

	return ParseConfig(data, pathToFile)
}

// ParseConfig parses data, the content of the goro.yaml at path. The config records path in Path,
// and a relative work_dir in it counts from the directory that holds the file (ADR 0002).
func ParseConfig(data []byte, path string) (Config, error) {
	var cfg Config

	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", path, err)
	}

	cfg.Path = path
	if wd := cfg.App.WorkDir; wd != "" && !filepath.IsAbs(wd) {
		cfg.App.WorkDir = filepath.Join(filepath.Dir(path), wd)
	}

	return cfg, nil
}

func (c *Config) AskAndSetName() error {
	if c.App.Name != "" {
		return nil
	}

	name, err := c.askName()
	if err != nil {
		return err
	}

	c.App.Name = name
	return nil
}

func (c *Config) askName() (string, error) {
	return (&promptui.Prompt{
		Label: "Enter an app name",
		Validate: func(s string) error {
			if s == "" {
				return fmt.Errorf("config name can't be empty")
			}
			return nil
		},
	}).Run()
}

func (c *Config) AskAndSetWorkDir() error {
	if c.App.WorkDir != "" {
		return nil
	}

	wd, err := c.askWorkDir()
	if err != nil {
		return err
	}

	if wd == "" {
		wd, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	c.App.WorkDir = wd
	return nil
}

func (c *Config) askWorkDir() (string, error) {
	return (&promptui.Prompt{
		Label: "Enter an app work dir or leave it empty to use current directory",
		// TODO: validate path
	}).Run()
}

func (c *Config) GetChunksByScope(scope string) []Chunk {
	chunks := make([]Chunk, 0, len(c.Chunks))
	for _, ch := range c.Chunks {
		if strings.HasPrefix(ch.Scope, scope) {
			chunks = append(chunks, ch)
		}
	}

	return chunks
}

func (c *Config) GetChunksByScopeAndInitHasErr(scope string) []Chunk {
	chunks := make([]Chunk, 0, len(c.Chunks))
	for _, ch := range c.Chunks {
		if strings.HasPrefix(ch.Scope, scope) && ch.InitHasErr {
			chunks = append(chunks, ch)
		}
	}

	return chunks
}

func (c *Config) GetChunksByScopeAndInitHasNotErr(scope string) []Chunk {
	chunks := make([]Chunk, 0, len(c.Chunks))
	for _, ch := range c.Chunks {
		if strings.HasPrefix(ch.Scope, scope) && !ch.InitHasErr {
			chunks = append(chunks, ch)
		}
	}

	return chunks
}
