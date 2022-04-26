package installer

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
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/bhojpur/iso/pkg/manager/compiler"
)

type RepositoryOption func(cfg *RepositoryConfig) error

type RepositoryConfig struct {
	Name, Description, Type string
	Urls                    []string
	Priority                int
	Src                     string
	Tree                    []string
	DB                      types.PackageDatabase
	CompilerBackend         compiler.CompilerBackend
	ImagePrefix             string

	context                                         types.Context
	PushImages, Force, FromRepository, FromMetadata bool
}

// Apply applies the given options to the config, returning the first error
// encountered (if any).
func (cfg *RepositoryConfig) Apply(opts ...RepositoryOption) error {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}

func WithContext(c types.Context) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.context = c
		return nil
	}
}

func WithDatabase(b types.PackageDatabase) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.DB = b
		return nil
	}
}

func WithCompilerBackend(b compiler.CompilerBackend) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.CompilerBackend = b
		return nil
	}
}

func WithTree(s ...string) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.Tree = append(cfg.Tree, s...)
		return nil
	}
}

func WithUrls(s ...string) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.Urls = append(cfg.Urls, s...)
		return nil
	}
}

func WithSource(s string) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.Src = s
		return nil
	}
}

func WithName(s string) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.Name = s
		return nil
	}
}

func WithDescription(s string) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.Description = s
		return nil
	}
}

func WithType(s string) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.Type = s
		return nil
	}
}

func WithImagePrefix(s string) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.ImagePrefix = s
		return nil
	}
}

func WithPushImages(b bool) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.PushImages = b
		return nil
	}
}

func WithForce(b bool) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.Force = b
		return nil
	}
}

// FromRepository when enabled
// considers packages metadata
// from remote repositories when building
// the new Bhojpur ISO repository index
func FromRepository(b bool) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.FromRepository = b
		return nil
	}
}

// FromMetadata when enabled
// considers packages metadata
// when building Bhojpur ISO repository indexes
func FromMetadata(b bool) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.FromMetadata = b
		return nil
	}
}

func WithPriority(b int) func(cfg *RepositoryConfig) error {
	return func(cfg *RepositoryConfig) error {
		cfg.Priority = b
		return nil
	}
}
