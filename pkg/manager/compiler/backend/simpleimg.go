package backend

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
	"os/exec"
	"strings"

	bus "github.com/bhojpur/iso/pkg/manager/api/core/bus"
	"github.com/bhojpur/iso/pkg/manager/api/core/image"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/pkg/errors"
)

type SimpleImg struct {
	ctx types.Context
}

func NewSimpleImgBackend(ctx types.Context) *SimpleImg {
	return &SimpleImg{ctx: ctx}
}

func (s *SimpleImg) LoadImage(string) error {
	return errors.New("Not supported")
}

// TODO: Missing still: labels, and build args expansion
func (s *SimpleImg) BuildImage(opts Options) error {
	name := opts.ImageName
	bus.Manager.Publish(bus.EventImagePreBuild, opts)

	buildarg := genBuildCommand(opts)

	s.ctx.Info(":tea: Building image " + name)

	cmd := exec.Command("img", buildarg...)
	cmd.Dir = opts.SourcePath
	err := runCommand(s.ctx, cmd)
	if err != nil {
		return err
	}
	bus.Manager.Publish(bus.EventImagePostBuild, opts)

	s.ctx.Info(":tea: Building image " + name + " done")

	return nil
}

func (s *SimpleImg) RemoveImage(opts Options) error {
	name := opts.ImageName
	buildarg := []string{"rm", name}
	s.ctx.Spinner()
	defer s.ctx.SpinnerStop()
	out, err := exec.Command("img", buildarg...).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Failed removing image: "+string(out))
	}

	s.ctx.Info(":tea: Image " + name + " removed")
	return nil
}

func (s *SimpleImg) ImageReference(a string, ondisk bool) (v1.Image, error) {

	f, err := s.ctx.TempFile("snapshot")
	if err != nil {
		return nil, err
	}
	buildarg := []string{"save", a, "-o", f.Name()}
	s.ctx.Spinner()
	defer s.ctx.SpinnerStop()

	out, err := exec.Command("img", buildarg...).CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "Failed saving image: "+string(out))
	}

	img, err := crane.Load(f.Name())
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (s *SimpleImg) DownloadImage(opts Options) error {
	name := opts.ImageName
	bus.Manager.Publish(bus.EventImagePrePull, opts)

	buildarg := []string{"pull", name}
	s.ctx.Debug(":tea: Downloading image " + name)

	s.ctx.Spinner()
	defer s.ctx.SpinnerStop()

	cmd := exec.Command("img", buildarg...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Failed downloading image: "+string(out))
	}

	s.ctx.Info(":tea: Image " + name + " downloaded")
	bus.Manager.Publish(bus.EventImagePostPull, opts)

	return nil
}
func (s *SimpleImg) CopyImage(src, dst string) error {
	s.ctx.Spinner()
	defer s.ctx.SpinnerStop()

	s.ctx.Debug(":tea: Tagging image", src, dst)
	cmd := exec.Command("img", "tag", src, dst)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Failed tagging image: "+string(out))
	}
	s.ctx.Info(":tea: Image " + dst + " tagged")

	return nil
}

func (s *SimpleImg) ImageAvailable(imagename string) bool {
	return image.Available(imagename)
}

// ImageExists check if the given image is available locally
func (*SimpleImg) ImageExists(imagename string) bool {
	cmd := exec.Command("img", "ls")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	if strings.Contains(string(out), imagename) {
		return true
	}
	return false
}

func (s *SimpleImg) ImageDefinitionToTar(opts Options) error {
	if err := s.BuildImage(opts); err != nil {
		return errors.Wrap(err, "Failed building image")
	}
	if err := s.ExportImage(opts); err != nil {
		return errors.Wrap(err, "Failed exporting image")
	}
	if err := s.RemoveImage(opts); err != nil {
		return errors.Wrap(err, "Failed removing image")
	}
	return nil
}

func (s *SimpleImg) ExportImage(opts Options) error {
	name := opts.ImageName
	path := opts.Destination
	buildarg := []string{"save", "-o", path, name}
	s.ctx.Debug(":tea: Saving image " + name)

	s.ctx.Spinner()
	defer s.ctx.SpinnerStop()

	out, err := exec.Command("img", buildarg...).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Failed exporting image: "+string(out))
	}
	s.ctx.Info(":tea: Image " + name + " saved")
	return nil
}

func (s *SimpleImg) Push(opts Options) error {
	name := opts.ImageName
	bus.Manager.Publish(bus.EventImagePrePush, opts)

	pusharg := []string{"push", name}
	out, err := exec.Command("img", pusharg...).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Failed pushing image: "+string(out))
	}
	s.ctx.Info(":tea: Pushed image:", name)
	bus.Manager.Publish(bus.EventImagePostPush, opts)

	//s.ctx.Info(string(out))
	return nil
}
