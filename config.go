package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func ConfigFromYamlFile(filePath string) (*Config, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := &Config{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return nil, err
	}
	for k, v := range cfg.Projects {
		configProjectDefaults(cfg, &v)
		cfg.Projects[k] = v
	}
	return cfg, nil
}

type ConfigDefaults struct {
	Mode       string `yaml:"mode"`
	Constraint string `yaml:"constraint"`
	Integrate  struct {
		Composer     *bool  `yaml:"composer"`
		Npm          *bool  `yaml:"npm"`
		NpmScriptKey string `yaml:"npm_script_key"`
	} `yaml:"integrate"`
	Pre  [][]string `yaml:"pre"`
	Post [][]string `yaml:"post"`
}

type ConfigProject struct {
	Remote string `yaml:"remote"`
	Local  string `yaml:"local"`
	ConfigDefaults
}

type Config struct {
	Listen string `yaml:"listen"`
	Token  struct {
		Key   string `yaml:"key"`
		Value string `yaml:"value"`
	} `yaml:"token"`
	Defaults ConfigDefaults           `yaml:"defaults"`
	Projects map[string]ConfigProject `yaml:"projects"`
}

func configProjectDefaults(c *Config, p *ConfigProject) {
	if p.Mode == "" {
		p.Mode = c.Defaults.Mode
	}
	if p.Constraint == "" {
		p.Constraint = c.Defaults.Constraint
	}
	if p.Integrate.Composer == nil {
		p.Integrate.Composer = c.Defaults.Integrate.Composer
	}
	if p.Integrate.Npm == nil {
		p.Integrate.Npm = c.Defaults.Integrate.Npm
	}
	if p.Integrate.NpmScriptKey == "" {
		p.Integrate.NpmScriptKey = c.Defaults.Integrate.NpmScriptKey
	}
}

func (r *Config) GetProject(name string) (*ConfigProject, error) {
	log.Println("GetProject: " + name)
	if p, ok := r.Projects[name]; ok {
		return &p, nil
	}
	return nil, fmt.Errorf("no such project: " + name)
}

func (r *Config) GetProjectByRemote(remote string) (*ConfigProject, error) {
	log.Println("GetProjectByRemote: " + remote)
	fmt.Println(r.Projects)
	for _, p := range r.Projects {
		log.Println("comparing [" + p.Remote + "] with [" + remote + "]")
		if p.Remote == remote {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("no such project: " + remote)
}
