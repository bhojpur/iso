package context

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
	"context"
	"os"
	"path/filepath"

	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	gc "github.com/bhojpur/iso/pkg/manager/api/core/garbagecollector"
	"github.com/bhojpur/iso/pkg/manager/api/core/logger"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"

	"github.com/pkg/errors"
)

type Context struct {
	*logger.Logger
	context.Context
	types.GarbageCollector
	Config      *types.BhojpurConfig
	NoSpinner   bool
	annotations map[string]interface{}
}

// SetAnnotation sets generic annotations to hold in a context
func (c *Context) SetAnnotation(s string, i interface{}) {
	c.annotations[s] = i
}

// GetAnnotation gets generic annotations to hold in a context
func (c *Context) GetAnnotation(s string) interface{} {
	return c.annotations[s]
}

type ContextOption func(c *Context) error

// WithLogger sets the logger
func WithLogger(l *logger.Logger) ContextOption {
	return func(c *Context) error {
		c.Logger = l
		return nil
	}
}

// WithConfig sets the Bhojpur ISO config
func WithConfig(cc *types.BhojpurConfig) ContextOption {
	return func(c *Context) error {
		c.Config = cc
		return nil
	}
}

// NOTE: GC needs to be instantiated when a new context is created from system TmpDirBase

// WithGarbageCollector sets the Garbage collector for the given context
func WithGarbageCollector(l types.GarbageCollector) ContextOption {
	return func(c *Context) error {
		if !filepath.IsAbs(l.String()) {
			abs, err := fileHelper.Rel2Abs(l.String())
			if err != nil {
				return errors.Wrap(err, "while converting relative path to absolute path")
			}
			l = gc.GarbageCollector(abs)
		}

		c.GarbageCollector = l
		return nil
	}
}

// NewContext returns a new context.
// It accepts a Garbage collector, a config and a logger as an option
func NewContext(opts ...ContextOption) *Context {
	l, _ := logger.New()
	d := &Context{
		annotations:      make(map[string]interface{}),
		Logger:           l,
		GarbageCollector: gc.GarbageCollector(filepath.Join(os.TempDir(), "tmpiso")),
		Config: &types.BhojpurConfig{
			ConfigFromHost: true,
			Logging:        types.BhojpurLoggingConfig{},
			General:        types.BhojpurGeneralConfig{},
			System: types.BhojpurSystemConfig{
				DatabasePath:  filepath.Join("var", "db"),
				PkgsCachePath: filepath.Join("var", "db", "packages"),
			},
			Solver: types.BhojpurSolverOptions{},
		},
	}

	for _, o := range opts {
		o(d)
	}
	return d
}

// WithLoggingContext returns a copy of the context with a contextualized logger
func (c *Context) WithLoggingContext(name string) types.Context {
	configCopy := *c.Config
	configCopy.System = c.Config.System
	configCopy.General = c.Config.General
	configCopy.Logging = c.Config.Logging

	ctx := *c
	ctxCopy := &ctx
	ctxCopy.Config = &configCopy
	ctxCopy.annotations = ctx.annotations

	ctxCopy.Logger, _ = c.Logger.Copy(logger.WithContext(name))

	return ctxCopy
}

// Copy returns a context copy with a reset logging context
func (c *Context) Copy() types.Context {
	return c.WithLoggingContext("")
}

func (c *Context) Warning(mess ...interface{}) {
	c.Logger.Warn(mess...)
	if c.Config.General.FatalWarns {
		panic("panic on warning")
	}
}

func (c *Context) Warn(mess ...interface{}) {
	c.Warning(mess...)
}

func (c *Context) Warnf(t string, mess ...interface{}) {
	c.Logger.Warnf(t, mess...)
	if c.Config.General.FatalWarns {
		panic("panic on warning")
	}
}

func (c *Context) Warningf(t string, mess ...interface{}) {
	c.Warnf(t, mess...)
}

func (c *Context) GetConfig() types.BhojpurConfig {
	return *c.Config
}
