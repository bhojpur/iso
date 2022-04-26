package schema

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
	"time"

	"github.com/pkg/errors"
	"github.com/twpayne/go-vfs"
	"gopkg.in/yaml.v2"
)

type SystemSpec struct {
	Initramfs       Initramfs       `yaml:"initramfs"`
	Label           string          `yaml:"label"`
	Packages        Packages        `yaml:"packages"`
	Bhojpur         Bhojpur         `yaml:"bhojpur"`
	Repository      Repository      `yaml:"repository"`
	Overlay         Overlay         `yaml:"overlay"`
	ImagePrefix     string          `yaml:"image_prefix"`
	Date            bool            `yaml:"image_date"`
	ImageName       string          `yaml:"image_name"`
	Arch            string          `yaml:"arch"`
	UEFIImage       string          `yaml:"uefi_img"`
	RootfsImage     string          `yaml:"rootfs_image"`
	SquashfsOptions SquashfsOptions `yaml:"squashfs_options"`

	BootFile     string `yaml:"boot_file"`
	BootCatalog  string `yaml:"boot_catalog"`
	IsoHybridMBR string `yaml:"isohybrid_mbr"`

	EnsureCommonDirs bool `yaml:"ensure_common_dirs"`
}

type Bhojpur struct {
	Repositories Repositories `yaml:"repositories"`
}

type BhojpurRepository struct {
	Name     string   `yaml:"name"`
	Enable   bool     `yaml:"enable"`
	Urls     []string `yaml:"urls"`
	Type     string   `yaml:"type"`
	Priority int      `yaml:"priority"`
}

type Repositories []*BhojpurRepository

func (r Repositories) Marshal() (string, error) {
	b, err := yaml.Marshal(&Bhojpur{Repositories: r})

	return string(b), err
}

func genRepo(name, url, t string) *BhojpurRepository {
	return &BhojpurRepository{Name: name, Enable: true, Urls: []string{url}, Type: t}
}

func NewDockerRepo(name, url string) *BhojpurRepository {
	return genRepo(name, url, "docker")
}

func NewHTTPRepo(name, url string) *BhojpurRepository {
	return genRepo(name, url, "http")
}

func NewLocalRepo(name, path string) *BhojpurRepository {
	return genRepo(name, path, "disk")
}

type Repository struct {
	Initramfs []string `yaml:"initramfs"`
	Packages  []string `yaml:"packages"`
}

type Packages struct {
	KeepBhojpurDB bool     `yaml:"keep_bhojpur_db"`
	Rootfs        []string `yaml:"rootfs"`
	Initramfs     []string `yaml:"initramfs"`
	IsoImage      []string `yaml:"isoimage"`
	UEFI          []string `yaml:"uefi"`
}

// Overlay represent additional folders to overlay on top of the rootfs, isoimage, or UEFI partition
type Overlay struct {
	Rootfs   string `yaml:"rootfs"`
	IsoImage string `yaml:"isoimage"`
	UEFI     string `yaml:"uefi"`
}

type Initramfs struct {
	KernelFile string `yaml:"kernel_file"`
	RootfsFile string `yaml:"rootfs_file"`
}

type SquashfsOptions struct {
	Compression        string `yaml:"compression"`
	CompressionOptions string `yaml:"compression_options"`
	Label              string `yaml:"label"`
}

func (s *SystemSpec) ISOName() (imageName string) {
	if s.ImageName != "" {
		imageName = s.ImageName
	}
	if s.ImagePrefix != "" {
		imageName = s.ImagePrefix + imageName
	}
	if s.Date {
		currentTime := time.Now()
		imageName = imageName + currentTime.Format("20060102")
	}
	if imageName == "" {
		imageName = "dev"
	}
	imageName = imageName + ".iso"
	return
}

// LoadFromFile loads a yip config from a YAML file
func LoadFromFile(s string, fs vfs.FS) (*SystemSpec, error) {
	yamlFile, err := fs.ReadFile(s)
	if err != nil {
		return nil, err
	}

	spec, err := LoadFromYaml(yamlFile)
	if err != nil {
		return nil, errors.Wrap(err, "while reading system spec")
	}

	return setDefaults(spec), nil
}

func setDefaults(s *SystemSpec) *SystemSpec {
	if s.BootCatalog == "" {
		s.BootCatalog = "boot/syslinux/boot.cat" // defaults to syslinux
	}
	if s.BootFile == "" {
		s.BootFile = "boot/syslinux/isolinux.bin"
	}
	if s.IsoHybridMBR == "" {
		s.IsoHybridMBR = "boot/syslinux/isohdpfx.bin"
	}
	if s.SquashfsOptions.Label == "" {
		s.SquashfsOptions.Label = "squashfs"
	}
	if s.SquashfsOptions.Compression == "" {
		s.SquashfsOptions.Compression = "xz"
		if s.SquashfsOptions.CompressionOptions == "" {
			s.SquashfsOptions.CompressionOptions = "-Xbcj x86"
		}
	}
	return s
}

// LoadFromYaml loads a config from bytes
func LoadFromYaml(b []byte) (*SystemSpec, error) {

	var yamlConfig SystemSpec
	err := yaml.Unmarshal(b, &yamlConfig)
	if err != nil {
		return nil, err
	}

	return &yamlConfig, nil
}
