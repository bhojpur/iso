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
	"os"
	"os/exec"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	box "github.com/bhojpur/iso/pkg/manager/box"
	"github.com/ghodss/yaml"

	"github.com/pkg/errors"
)

type BhojpurFinalizer struct {
	Shell     []string `json:"shell"`
	Install   []string `json:"install"`
	Uninstall []string `json:"uninstall"` // TODO: Where to store?
}

func (f *BhojpurFinalizer) RunInstall(ctx types.Context, s *System) error {
	var cmd string
	var args []string
	if len(f.Shell) == 0 {
		// Default to sh otherwise
		cmd = "sh"
		args = []string{"-c"}
	} else {
		cmd = f.Shell[0]
		if len(f.Shell) > 1 {
			args = f.Shell[1:]
		}
	}

	for _, c := range f.Install {
		toRun := append(args, c)
		ctx.Info(":shell: Executing finalizer on ", s.Target, cmd, toRun)
		if s.Target == string(os.PathSeparator) {
			cmd := exec.Command(cmd, toRun...)
			cmd.Env = ctx.GetConfig().FinalizerEnvs.Slice()
			stdoutStderr, err := cmd.CombinedOutput()
			if err != nil {
				return errors.Wrap(err, "Failed running command: "+string(stdoutStderr))
			}
			ctx.Info(string(stdoutStderr))
		} else {
			b := box.NewBox(cmd, toRun, []string{}, ctx.GetConfig().FinalizerEnvs.Slice(), s.Target, false, true, true)
			err := b.Run()
			if err != nil {
				return errors.Wrap(err, "Failed running command ")
			}
		}
	}
	return nil
}

// TODO: We don't store uninstall finalizers ?!
func (f *BhojpurFinalizer) RunUnInstall(ctx types.Context) error {
	for _, c := range f.Uninstall {
		ctx.Debug("finalizer:", "sh", "-c", c)
		cmd := exec.Command("sh", "-c", c)
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			return errors.Wrap(err, "Failed running command: "+string(stdoutStderr))
		}
		ctx.Info(string(stdoutStderr))
	}
	return nil
}

func NewBhojpurFinalizerFromYaml(data []byte) (*BhojpurFinalizer, error) {
	var p BhojpurFinalizer
	err := yaml.Unmarshal(data, &p)
	if err != nil {
		return &p, err
	}
	return &p, err
}
