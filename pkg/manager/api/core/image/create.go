package image

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
	"io"
	"os"

	containerdCompression "github.com/containerd/containerd/archive/compression"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/pkg/errors"
)

func imageFromTar(imagename, architecture, OS string, opener func() (io.ReadCloser, error)) (name.Reference, v1.Image, error) {
	newRef, err := name.ParseReference(imagename)
	if err != nil {
		return nil, nil, err
	}

	layer, err := tarball.LayerFromOpener(opener)
	if err != nil {
		return nil, nil, err
	}

	baseImage := empty.Image
	cfg, err := baseImage.ConfigFile()
	if err != nil {
		return nil, nil, err
	}

	cfg.Architecture = architecture
	cfg.OS = OS

	baseImage, err = mutate.ConfigFile(baseImage, cfg)
	if err != nil {
		return nil, nil, err
	}
	img, err := mutate.Append(baseImage, mutate.Addendum{
		Layer: layer,
		History: v1.History{
			CreatedBy: "Bhojpur",
			Comment:   "Custom image",
		},
	})
	if err != nil {
		return nil, nil, err
	}

	return newRef, img, nil
}

// CreateTar a imagetarball from a standard tarball
func CreateTar(srctar, dstimageTar, imagename, architecture, OS string) error {

	dstFile, err := os.Create(dstimageTar)
	if err != nil {
		return errors.Wrap(err, "Cannot create "+dstimageTar)
	}
	defer dstFile.Close()

	newRef, img, err := imageFromTar(imagename, architecture, OS, func() (io.ReadCloser, error) {
		f, err := os.Open(srctar)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot open "+srctar)
		}
		decompressed, err := containerdCompression.DecompressStream(f)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot open "+srctar)
		}

		return decompressed, nil
	})
	if err != nil {
		return err
	}

	// NOTE: We might also stream that back to the daemon with daemon.Write(tag, img)
	return tarball.Write(newRef, img, dstFile)

}
