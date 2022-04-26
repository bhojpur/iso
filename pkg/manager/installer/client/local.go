package client

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

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"
	"github.com/pkg/errors"
)

type LocalClient struct {
	RepoData RepoData
	Cache    *artifact.ArtifactCache
	context  types.Context
}

func NewLocalClient(r RepoData, ctx types.Context) *LocalClient {
	return &LocalClient{
		Cache:    artifact.NewCache(ctx.GetConfig().System.PkgsCachePath),
		RepoData: r,
		context:  ctx,
	}
}

func (c *LocalClient) DownloadArtifact(a *artifact.PackageArtifact) (*artifact.PackageArtifact, error) {
	var err error

	artifactName := path.Base(a.Path)

	newart, err := c.CacheGet(a)
	// Check if file is already in cache
	if err == nil {
		return newart, nil
	}

	d, err := c.DownloadFile(artifactName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed downloading %s", artifactName)
	}
	defer os.RemoveAll(d)

	newart.Path = d
	c.Cache.Put(newart)

	return c.CacheGet(newart)
}

func (c *LocalClient) CacheGet(a *artifact.PackageArtifact) (*artifact.PackageArtifact, error) {
	newart := a.ShallowCopy()
	fileName, err := c.Cache.Get(a)

	newart.Path = fileName

	return newart, err
}

func (c *LocalClient) DownloadFile(name string) (string, error) {
	var err error
	var file *os.File = nil

	rootfs := ""

	if !c.context.GetConfig().ConfigFromHost {
		rootfs = c.context.GetConfig().System.Rootfs
	}

	ok := false
	for _, uri := range c.RepoData.Urls {

		uri = filepath.Join(rootfs, uri)

		c.context.Info("Copying file", name, "from", uri)
		file, err = c.context.TempFile("localclient")
		if err != nil {
			continue
		}
		//defer os.Remove(file.Name())

		err = fileHelper.CopyFile(filepath.Join(uri, name), file.Name())
		if err != nil {
			continue
		}
		ok = true
		break
	}

	if ok {
		return file.Name(), nil
	}

	return "", err
}
