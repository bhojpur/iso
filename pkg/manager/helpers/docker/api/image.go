package api

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
	"os"

	"os/exec"

	jww "github.com/spf13/jwalterweatherman"
)

type ExportedImage struct {
	Path         string
	JsonPath     string
	VersionPath  string
	LayerTarPath string
	LayerDirPath string
}

func (e *ExportedImage) CreateDirs() error {
	return os.MkdirAll(e.Path, 0755)
}

func (e *ExportedImage) TarLayer() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	err = os.Chdir(e.LayerDirPath)
	if err != nil {
		return err
	}
	defer os.Chdir(cwd)

	cmd := exec.Command("sudo", "/bin/sh", "-c", fmt.Sprintf("%s cvf ../layer.tar ./", TarCmd))
	out, err := cmd.CombinedOutput()
	if err != nil {
		jww.INFO.Println(out)
		return err
	}
	return nil
}

func (e *ExportedImage) RemoveLayerDir() error {
	return os.RemoveAll(e.LayerDirPath)
}

func (e *ExportedImage) ExtractLayerDir(unpackmode string) error {
	err := os.MkdirAll(e.LayerDirPath, 0755)
	if err != nil {
		return err
	}

	if err := ExtractLayer(&ExtractOpts{
		Source:       e.LayerTarPath,
		Destination:  e.LayerDirPath,
		Compressed:   true,
		KeepDirlinks: true,
		Rootless:     false,
		UnpackMode:   unpackmode}); err != nil {
		return err
	}
	return nil
}
