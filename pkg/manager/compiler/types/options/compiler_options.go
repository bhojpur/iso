package options

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
	"runtime"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/bhojpur/iso/pkg/manager/compiler/types/compression"
)

type Compiler struct {
	PushImageRepository      string
	PullImageRepository      []string
	PullFirst, KeepImg, Push bool
	Concurrency              int
	CompressionType          compression.Implementation

	Wait            bool
	OnlyDeps        bool
	NoDeps          bool
	SolverOptions   types.BhojpurSolverOptions
	BuildValuesFile []string
	BuildValues     []map[string]interface{}

	PackageTargetOnly bool
	Rebuild           bool

	BackendArgs []string

	BackendType string

	// TemplatesFolder. should default to tree/templates
	TemplatesFolder []string

	// Tells wether to push final container images after building
	PushFinalImages      bool
	PushFinalImagesForce bool

	GenerateFinalImages bool

	// Image repository to push to
	PushFinalImagesRepository string
	RuntimeDatabase           types.PackageDatabase

	Context types.Context
}

func NewDefaultCompiler() *Compiler {
	return &Compiler{
		PushImageRepository: "bhojpur/cache",
		PullFirst:           false,
		Push:                false,
		CompressionType:     compression.None,
		KeepImg:             true,
		Concurrency:         runtime.NumCPU(),
		OnlyDeps:            false,
		NoDeps:              false,
		SolverOptions:       types.BhojpurSolverOptions{SolverOptions: types.SolverOptions{Concurrency: 1, Type: types.SolverSingleCoreSimple}},
	}
}

type Option func(cfg *Compiler) error

func (cfg *Compiler) Apply(opts ...Option) error {
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

func WithOptions(opt *Compiler) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg = opt
		return nil
	}
}

func WithRuntimeDatabase(db types.PackageDatabase) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.RuntimeDatabase = db
		return nil
	}
}

// WithFinalRepository Sets the final repository where to push
// images of built artifacts
func WithFinalRepository(r string) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.PushFinalImagesRepository = r
		return nil
	}
}

func EnableGenerateFinalImages(cfg *Compiler) error {
	cfg.GenerateFinalImages = true
	return nil
}

func EnablePushFinalImages(cfg *Compiler) error {
	cfg.PushFinalImages = true
	return nil
}

func ForcePushFinalImages(cfg *Compiler) error {
	cfg.PushFinalImagesForce = true
	return nil
}

func WithBackendType(r string) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.BackendType = r
		return nil
	}
}

func WithTemplateFolder(r []string) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.TemplatesFolder = r
		return nil
	}
}

func WithBuildValues(r []string) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.BuildValuesFile = r
		return nil
	}
}

func WithPullRepositories(r []string) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.PullImageRepository = r
		return nil
	}
}

// WithPushRepository Sets the image reference where to push
// cache images
func WithPushRepository(r string) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		if len(cfg.PullImageRepository) == 0 {
			cfg.PullImageRepository = []string{cfg.PushImageRepository}
		}
		cfg.PushImageRepository = r
		return nil
	}
}

func BackendArgs(r []string) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.BackendArgs = r
		return nil
	}
}

func PullFirst(b bool) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.PullFirst = b
		return nil
	}
}

func KeepImg(b bool) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.KeepImg = b
		return nil
	}
}

func Rebuild(b bool) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.Rebuild = b
		return nil
	}
}

func PushImages(b bool) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.Push = b
		return nil
	}
}

func Wait(b bool) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.Wait = b
		return nil
	}
}

func OnlyDeps(b bool) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.OnlyDeps = b
		return nil
	}
}

func OnlyTarget(b bool) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.PackageTargetOnly = b
		return nil
	}
}

func NoDeps(b bool) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.NoDeps = b
		return nil
	}
}

func Concurrency(i int) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		if i == 0 {
			i = runtime.NumCPU()
		}
		cfg.Concurrency = i
		return nil
	}
}

func WithCompressionType(t compression.Implementation) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.CompressionType = t
		return nil
	}
}

func WithSolverOptions(c types.BhojpurSolverOptions) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.SolverOptions = c
		return nil
	}
}

func WithContext(c types.Context) func(cfg *Compiler) error {
	return func(cfg *Compiler) error {
		cfg.Context = c
		return nil
	}
}
