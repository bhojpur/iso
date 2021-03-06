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
	"os"
	"path"
	"path/filepath"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
)

type Export struct {
	Entries map[string]*ExportedImage
	Path    string
}

func (e *Export) ExtractLayers(unpackmode string) error {

	jww.INFO.Println("Extracting layers...")

	for _, entry := range e.Entries {
		jww.INFO.Println("  - ", entry.LayerTarPath)
		err := entry.ExtractLayerDir(unpackmode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Export) UnPackLayers(order []string, layerDir string, unpackmode string) error {
	err := os.MkdirAll(layerDir, 0755)
	if err != nil {
		return err
	}

	for _, ee := range order {
		entry := e.Entries[ee]
		if _, err := os.Stat(entry.LayerTarPath); os.IsNotExist(err) {
			continue
		}

		err := ExtractLayer(&ExtractOpts{
			Source:       entry.LayerTarPath,
			Destination:  layerDir,
			Compressed:   true,
			KeepDirlinks: true,
			Rootless:     false,
			UnpackMode:   unpackmode})
		if err != nil {
			jww.INFO.Println(err.Error())
			return err
		}

		jww.INFO.Println("  -  Deleting whiteouts for layer " + ee)
		err = e.deleteWhiteouts(layerDir)
		if err != nil {
			return err
		}
	}
	return nil
}

const TarCmd = "tar"

func (e *Export) deleteWhiteouts(location string) error {
	return filepath.Walk(location, func(p string, info os.FileInfo, err error) error {
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		if info == nil {
			return nil
		}

		name := info.Name()
		parent := filepath.Dir(p)
		// if start with whiteout
		if strings.Index(name, ".wh.") == 0 {
			deletedFile := path.Join(parent, name[len(".wh."):len(name)])
			// remove deleted files
			if err := os.RemoveAll(deletedFile); err != nil {
				return err
			}
			// remove the whiteout itself
			if err := os.RemoveAll(p); err != nil {
				return err
			}
		}
		return nil
	})
}
