package types

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/bhojpur/iso/pkg/manager/api/core/config"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// BhojpurLoggingConfig is the config
// relative to logging of Bhojpur ISO
type BhojpurLoggingConfig struct {
	// Path of the logfile
	Path string `yaml:"path" mapstructure:"path"`
	// Enable/Disable logging to file
	EnableLogFile bool `yaml:"enable_logfile" mapstructure:"enable_logfile"`
	// Enable JSON format logging in file
	JSONFormat bool `yaml:"json_format" mapstructure:"json_format"`

	// Log level
	Level string `yaml:"level" mapstructure:"level"`

	// Enable emoji
	EnableEmoji bool `yaml:"enable_emoji" mapstructure:"enable_emoji"`
	// Enable/Disable color in logging
	Color bool `yaml:"color" mapstructure:"color"`

	// NoSpinner disable spinner
	NoSpinner bool `yaml:"no_spinner" mapstructure:"no_spinner"`
}

// BhojpurGeneralConfig is the general configuration structure
// which applies to all the Bhojpur ISO actions
type BhojpurGeneralConfig struct {
	SameOwner       bool `yaml:"same_owner,omitempty" mapstructure:"same_owner"`
	Concurrency     int  `yaml:"concurrency,omitempty" mapstructure:"concurrency"`
	Debug           bool `yaml:"debug,omitempty" mapstructure:"debug"`
	ShowBuildOutput bool `yaml:"show_build_output,omitempty" mapstructure:"show_build_output"`
	FatalWarns      bool `yaml:"fatal_warnings,omitempty" mapstructure:"fatal_warnings"`
	HTTPTimeout     int  `yaml:"http_timeout,omitempty" mapstructure:"http_timeout"`
	Quiet           bool `yaml:"quiet" mapstructure:"quiet"`
}

// BhojpurSolverOptions this is the option struct for the Bhojpur ISO solver
type BhojpurSolverOptions struct {
	SolverOptions  `yaml:"options,omitempty"`
	Type           string     `yaml:"type,omitempty" mapstructure:"type"`
	LearnRate      float32    `yaml:"rate,omitempty" mapstructure:"rate"`
	Discount       float32    `yaml:"discount,omitempty" mapstructure:"discount"`
	MaxAttempts    int        `yaml:"max_attempts,omitempty" mapstructure:"max_attempts"`
	Implementation SolverType `yaml:"implementation,omitempty" mapstructure:"implementation"`
}

// CompactString returns a compact string to display solver options over CLI
func (opts *BhojpurSolverOptions) CompactString() string {
	return fmt.Sprintf("type: %s rate: %f, discount: %f, attempts: %d, initialobserved: %d",
		opts.Type, opts.LearnRate, opts.Discount, opts.MaxAttempts, 999999)
}

// BhojpurSystemConfig is the system configuration.
// Typically this represent a host system that is about to perform
// operations on a Rootfs. Note all the fields needs to be in absolute form.
type BhojpurSystemConfig struct {
	DatabaseEngine string `yaml:"database_engine" mapstructure:"database_engine"`
	DatabasePath   string `yaml:"database_path" mapstructure:"database_path"`
	Rootfs         string `yaml:"rootfs" mapstructure:"rootfs"`
	PkgsCachePath  string `yaml:"pkgs_cache_path" mapstructure:"pkgs_cache_path"`
	TmpDirBase     string `yaml:"tmpdir_base" mapstructure:"tmpdir_base"`
}

// Init reads the config and replace user-defined paths with
// absolute paths where necessary, and construct the paths for the cache
// and database on the real system
func (c *BhojpurConfig) Init() error {
	if err := c.System.init(); err != nil {
		return err
	}

	if err := c.loadConfigProtect(); err != nil {
		return err
	}

	// Load repositories
	if err := c.loadRepositories(); err != nil {
		return err
	}

	return nil
}

func (s *BhojpurSystemConfig) init() error {
	if err := s.setRootfs(); err != nil {
		return err
	}

	if err := s.setDBPath(); err != nil {
		return err
	}

	s.setCachePath()

	return nil
}

func (s *BhojpurSystemConfig) setRootfs() error {
	p, err := fileHelper.Rel2Abs(s.Rootfs)
	if err != nil {
		return err
	}

	s.Rootfs = p
	return nil
}

// GetRepoDatabaseDirPath is synatx sugar to return the repository path given
// a repository name in the system target
func (s BhojpurSystemConfig) GetRepoDatabaseDirPath(name string) string {
	dbpath := filepath.Join(s.DatabasePath, "repos/"+name)
	err := os.MkdirAll(dbpath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return dbpath
}

func (s *BhojpurSystemConfig) setDBPath() error {
	dbpath := filepath.Join(
		s.Rootfs,
		s.DatabasePath,
	)
	err := os.MkdirAll(dbpath, os.ModePerm)
	if err != nil {
		return err
	}
	s.DatabasePath = dbpath
	return nil
}

func (s *BhojpurSystemConfig) setCachePath() {
	var cachepath string
	if s.PkgsCachePath != "" {
		if !filepath.IsAbs(cachepath) {
			cachepath = filepath.Join(s.DatabasePath, s.PkgsCachePath)
			os.MkdirAll(cachepath, os.ModePerm)
		} else {
			cachepath = s.PkgsCachePath
		}
	} else {
		// Create dynamic cache for test suites
		cachepath, _ = ioutil.TempDir(os.TempDir(), "cachepkgs")
	}

	s.PkgsCachePath = cachepath // Be consistent with the path we set
}

// FinalizerEnv represent a Key/Value environment to be set
// while running a package finalizer
type FinalizerEnv struct {
	Key   string `json:"key" yaml:"key" mapstructure:"key"`
	Value string `json:"value" yaml:"value" mapstructure:"value"`
}

// Finalizers are a slice of K/V environments to set
// while running package finalizers
type Finalizers []FinalizerEnv

// Slice returns the finalizers as a string slice in
// k=v form.
func (f Finalizers) Slice() (sl []string) {
	for _, kv := range f {
		sl = append(sl, fmt.Sprintf("%s=%s", kv.Key, kv.Value))
	}
	return
}

// BhojpurConfig is the general structure which holds
// all the configuration fields.
// It includes, Logging, General, System and Solver sub configurations.
type BhojpurConfig struct {
	Logging BhojpurLoggingConfig `yaml:"logging,omitempty" mapstructure:"logging"`
	General BhojpurGeneralConfig `yaml:"general,omitempty" mapstructure:"general"`
	System  BhojpurSystemConfig  `yaml:"system" mapstructure:"system"`
	Solver  BhojpurSolverOptions `yaml:"solver,omitempty" mapstructure:"solver"`

	RepositoriesConfDir  []string            `yaml:"repos_confdir,omitempty" mapstructure:"repos_confdir"`
	ConfigProtectConfDir []string            `yaml:"config_protect_confdir,omitempty" mapstructure:"config_protect_confdir"`
	ConfigProtectSkip    bool                `yaml:"config_protect_skip,omitempty" mapstructure:"config_protect_skip"`
	ConfigFromHost       bool                `yaml:"config_from_host,omitempty" mapstructure:"config_from_host"`
	SystemRepositories   BhojpurRepositories `yaml:"repositories,omitempty" mapstructure:"repositories"`

	FinalizerEnvs Finalizers `json:"finalizer_envs,omitempty" yaml:"finalizer_envs,omitempty" mapstructure:"finalizer_envs,omitempty"`

	ConfigProtectConfFiles []config.ConfigProtectConfFile `yaml:"-" mapstructure:"-"`
}

// AddSystemRepository is just syntax sugar to add a repository in the system set
func (c *BhojpurConfig) AddSystemRepository(r BhojpurRepository) {
	c.SystemRepositories = append(c.SystemRepositories, r)
}

// SetFinalizerEnv sets a k,v couple among the finalizers
// It ensures that the key is unique, and if set again it gets updated
func (c *BhojpurConfig) SetFinalizerEnv(k, v string) {
	keyPresent := false
	envs := []FinalizerEnv{}

	for _, kv := range c.FinalizerEnvs {
		if kv.Key == k {
			keyPresent = true
			envs = append(envs, FinalizerEnv{Key: kv.Key, Value: v})
		} else {
			envs = append(envs, kv)
		}
	}
	if !keyPresent {
		envs = append(envs, FinalizerEnv{Key: k, Value: v})
	}

	c.FinalizerEnvs = envs
}

// YAML returns the config in yaml format
func (c *BhojpurConfig) YAML() ([]byte, error) {
	return yaml.Marshal(c)
}

func (c *BhojpurConfig) addProtectFile(file *config.ConfigProtectConfFile) {
	if c.ConfigProtectConfFiles == nil {
		c.ConfigProtectConfFiles = []config.ConfigProtectConfFile{*file}
	} else {
		c.ConfigProtectConfFiles = append(c.ConfigProtectConfFiles, *file)
	}
}

func (c *BhojpurConfig) loadRepositories() error {
	var regexRepo = regexp.MustCompile(`.yml$|.yaml$`)
	rootfs := ""

	// Respect the rootfs param on read repositories
	if !c.ConfigFromHost {
		rootfs = c.System.Rootfs
	}

	for _, rdir := range c.RepositoriesConfDir {

		rdir = filepath.Join(rootfs, rdir)

		files, err := ioutil.ReadDir(rdir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			if !regexRepo.MatchString(file.Name()) {
				continue
			}

			content, err := ioutil.ReadFile(path.Join(rdir, file.Name()))
			if err != nil {
				continue
			}

			r, err := LoadRepository(content)
			if err != nil {
				continue
			}

			if r.Name == "" || len(r.Urls) == 0 || r.Type == "" {
				continue
			}

			c.AddSystemRepository(*r)
		}
	}
	return nil
}

// GetSystemRepository retrieve the system repository inside the configuration
// Note, the configuration needs to be loaded first.
func (c *BhojpurConfig) GetSystemRepository(name string) (*BhojpurRepository, error) {
	var ans *BhojpurRepository

	for idx, repo := range c.SystemRepositories {
		if repo.Name == name {
			ans = &c.SystemRepositories[idx]
			break
		}
	}
	if ans == nil {
		return nil, errors.New("Repository " + name + " not found")
	}

	return ans, nil
}

func (c *BhojpurConfig) loadConfigProtect() error {
	var regexConfs = regexp.MustCompile(`.yml$`)
	rootfs := ""

	// Respect the rootfs param on read repositories
	if !c.ConfigFromHost {
		rootfs = c.System.Rootfs
	}

	for _, cdir := range c.ConfigProtectConfDir {
		cdir = filepath.Join(rootfs, cdir)

		files, err := ioutil.ReadDir(cdir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			if !regexConfs.MatchString(file.Name()) {
				continue
			}

			content, err := ioutil.ReadFile(path.Join(cdir, file.Name()))
			if err != nil {
				continue
			}

			r, err := loadConfigProtectConfFile(file.Name(), content)
			if err != nil {
				continue
			}

			if r.Name == "" || len(r.Directories) == 0 {
				continue
			}

			c.addProtectFile(r)
		}
	}
	return nil

}

func loadConfigProtectConfFile(filename string, data []byte) (*config.ConfigProtectConfFile, error) {
	ans := config.NewConfigProtectConfFile(filename)
	err := yaml.Unmarshal(data, &ans)
	if err != nil {
		return nil, err
	}
	return ans, nil
}
