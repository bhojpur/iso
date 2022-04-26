package config

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
	"path/filepath"
	"strings"
)

type ConfigProtectConfFile struct {
	Filename string

	Name        string   `mapstructure:"name" yaml:"name" json:"name"`
	Directories []string `mapstructure:"dirs" yaml:"dirs" json:"dirs"`
}

func NewConfigProtectConfFile(filename string) *ConfigProtectConfFile {
	return &ConfigProtectConfFile{
		Filename:    filename,
		Name:        "",
		Directories: []string{},
	}
}

func (c *ConfigProtectConfFile) String() string {
	return fmt.Sprintf("[%s] filename: %s, dirs: %s", c.Name, c.Filename,
		c.Directories)
}

type ConfigProtect struct {
	AnnotationDir string
	MapProtected  map[string]bool
}

func NewConfigProtect(annotationDir string) *ConfigProtect {
	if len(annotationDir) > 0 && annotationDir[0:1] != "/" {
		annotationDir = "/" + annotationDir
	}
	return &ConfigProtect{
		AnnotationDir: annotationDir,
		MapProtected:  make(map[string]bool),
	}
}

func (c *ConfigProtect) Map(files []string, protected []ConfigProtectConfFile) {

	for _, file := range files {

		if file[0:1] != "/" {
			file = "/" + file
		}

		if len(protected) > 0 {
			for _, conf := range protected {
				for _, dir := range conf.Directories {
					// Note file is without / at begin (on unpack)
					if strings.HasPrefix(file, filepath.Clean(dir)) {
						// docker archive modifier works with path without / at begin.
						c.MapProtected[file] = true
						goto nextFile
					}
				}
			}
		}

		if c.AnnotationDir != "" && strings.HasPrefix(file, filepath.Clean(c.AnnotationDir)) {
			c.MapProtected[file] = true
		}
	nextFile:
	}

}

func (c *ConfigProtect) Protected(file string) bool {
	if file[0:1] != "/" {
		file = "/" + file
	}
	_, ans := c.MapProtected[file]
	return ans
}

func (c *ConfigProtect) GetProtectFiles(withSlash bool) []string {
	ans := []string{}

	for key := range c.MapProtected {
		if withSlash {
			ans = append(ans, key)
		} else {
			ans = append(ans, key[1:])
		}
	}
	return ans
}
