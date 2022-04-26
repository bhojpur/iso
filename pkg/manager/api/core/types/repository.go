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
	"runtime"

	"gopkg.in/yaml.v2"
)

type BhojpurRepository struct {
	Name           string            `json:"name" yaml:"name" mapstructure:"name"`
	Description    string            `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description"`
	Urls           []string          `json:"urls" yaml:"urls" mapstructure:"urls"`
	Type           string            `json:"type" yaml:"type" mapstructure:"type"`
	Mode           string            `json:"mode,omitempty" yaml:"mode,omitempty" mapstructure:"mode,omitempty"`
	Priority       int               `json:"priority,omitempty" yaml:"priority,omitempty" mapstructure:"priority"`
	Enable         bool              `json:"enable" yaml:"enable" mapstructure:"enable"`
	Cached         bool              `json:"cached,omitempty" yaml:"cached,omitempty" mapstructure:"cached,omitempty"`
	Authentication map[string]string `json:"auth,omitempty" yaml:"auth,omitempty" mapstructure:"auth,omitempty"`
	TreePath       string            `json:"treepath,omitempty" yaml:"treepath,omitempty" mapstructure:"treepath"`
	MetaPath       string            `json:"metapath,omitempty" yaml:"metapath,omitempty" mapstructure:"metapath"`
	Verify         bool              `json:"verify,omitempty" yaml:"verify,omitempty" mapstructure:"verify"`
	Arch           string            `json:"arch,omitempty" yaml:"arch,omitempty" mapstructure:"arch"`

	ReferenceID string `json:"reference,omitempty" yaml:"reference,omitempty" mapstructure:"reference"`

	// Incremented value that identify revision of the repository in a user-friendly way.
	Revision int `json:"revision,omitempty" yaml:"-" mapstructure:"-"`
	// Epoch time in seconds
	LastUpdate string `json:"last_update,omitempty" yaml:"-" mapstructure:"-"`
}

func (r *BhojpurRepository) String() string {
	return fmt.Sprintf("[%s] prio: %d, type: %s, enable: %t, cached: %t",
		r.Name, r.Priority, r.Type, r.Enable, r.Cached)
}

// Enabled returns a boolean indicating if the repository should be considered enabled or not
func (r *BhojpurRepository) Enabled() bool {
	return r.Arch != "" && r.Arch == runtime.GOARCH && !r.Enable || r.Enable
}

type BhojpurRepositories []BhojpurRepository

func (l BhojpurRepositories) Enabled() (res BhojpurRepositories) {
	for _, r := range l {
		if r.Enabled() {
			res = append(res, r)
		}
	}
	return
}

func NewBhojpurRepository(name, t, descr string, urls []string, priority int, enable, cached bool) *BhojpurRepository {
	return &BhojpurRepository{
		Name:           name,
		Description:    descr,
		Urls:           urls,
		Type:           t,
		Priority:       priority,
		Enable:         enable,
		Cached:         cached,
		Authentication: make(map[string]string),
	}
}

func NewEmptyBhojpurRepository() *BhojpurRepository {
	return &BhojpurRepository{
		Priority:       9999,
		Authentication: make(map[string]string),
	}
}

func LoadRepository(data []byte) (*BhojpurRepository, error) {
	ans := NewEmptyBhojpurRepository()
	err := yaml.Unmarshal(data, &ans)
	if err != nil {
		return nil, err
	}
	return ans, nil
}
