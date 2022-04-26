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
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	bhojpurtypes "github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/go-units"
	"github.com/pkg/errors"

	"github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	"github.com/bhojpur/iso/pkg/manager/helpers"

	"github.com/bhojpur/iso/pkg/manager/helpers/docker"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"
)

const (
	errImageDownloadMsg = "failed downloading image %s: %s"
)

type DockerClient struct {
	RepoData RepoData
	auth     *types.AuthConfig
	Cache    *artifact.ArtifactCache
	context  bhojpurtypes.Context
}

func NewDockerClient(r RepoData, ctx bhojpurtypes.Context) *DockerClient {
	auth := &types.AuthConfig{}

	dat, _ := json.Marshal(r.Authentication)
	json.Unmarshal(dat, auth)

	return &DockerClient{RepoData: r, auth: auth,
		Cache:   artifact.NewCache(ctx.GetConfig().System.PkgsCachePath),
		context: ctx,
	}
}

func (c *DockerClient) DownloadArtifact(a *artifact.PackageArtifact) (*artifact.PackageArtifact, error) {
	//var u *url.URL = nil
	var err error

	c.context.Spinner()
	defer c.context.SpinnerStop()

	downloaded := false

	resultingArtifact, err := c.CacheGet(a)
	if err == nil {
		return resultingArtifact, nil
	}

	// TODO:
	// Files are in URI/packagename:version (GetPackageImageName() method)
	// use downloadAndExtract .. and egenrate an archive to consume. Checksum should be already checked while downloading the image
	// with the above functions, because Docker images already contain such metadata
	// - Check how verification is done when calling DownloadArtifact outside, similarly we need to check DownloadFile, and how verification
	// is done in such cases (see repository.go)

	// We discard checksum, that are checked while during pull and unpack by containerd
	resultingArtifact.Checksums = artifact.Checksums{}

	temp, err := c.context.TempDir("image")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(temp)

	tempArtifact, err := c.context.TempFile("artifact")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempArtifact.Name())
	for _, uri := range c.RepoData.Urls {

		imageName := fmt.Sprintf("%s:%s", uri, a.CompileSpec.GetPackage().ImageID())
		c.context.Info("Downloading image", imageName)

		// imageName := fmt.Sprintf("%s/%s", uri, artifact.GetCompileSpec().GetPackage().GetPackageImageName())
		info, err := docker.DownloadAndExtractDockerImage(c.context, imageName, temp, c.auth, c.RepoData.Verify)
		if err != nil {
			c.context.Warning(fmt.Sprintf(errImageDownloadMsg, imageName, err.Error()))
			continue
		}

		c.context.Info(
			fmt.Sprintf("Image: %s. Pulled: %s. Size: %s",
				imageName,
				info.Target.Digest,
				units.BytesSize(float64(info.Target.Size)),
			),
		)
		c.context.Debug("\nCompressing result ", filepath.Join(temp), "to", tempArtifact.Name())

		resultingArtifact.Path = tempArtifact.Name() // First set to cache file
		err = resultingArtifact.Compress(temp, 1)
		if err != nil {
			c.context.Error(fmt.Sprintf("Failed compressing package %s: %s", imageName, err.Error()))
			continue
		}

		_, _, err = c.Cache.Put(resultingArtifact)
		if err != nil {
			c.context.Error(fmt.Sprintf("Failed storing package %s from cache: %s", imageName, err.Error()))
			continue
		}
		downloaded = true

		return c.CacheGet(resultingArtifact)
	}

	if !downloaded {
		return nil, errors.Wrap(err, "no image available from repositories")
	}

	return resultingArtifact, nil
}

func (c *DockerClient) DownloadFile(name string) (string, error) {
	var file *os.File = nil
	var err error
	var temp string
	// Files should be in URI/repository:<file>
	ok := false

	temp, err = c.context.TempDir("tree")
	if err != nil {
		return "", err
	}

	for _, uri := range c.RepoData.Urls {
		file, err = c.context.TempFile("DockerClient")
		if err != nil {
			continue
		}

		imageName := fmt.Sprintf("%s:%s", uri, helpers.SanitizeImageString(name))
		c.context.Info("Downloading", imageName)

		info, err := docker.DownloadAndExtractDockerImage(c.context, imageName, temp, c.auth, c.RepoData.Verify)
		if err != nil {
			c.context.Warning(fmt.Sprintf(errImageDownloadMsg, imageName, err.Error()))
			continue
		}

		c.context.Info(fmt.Sprintf("Pulled: %s", info.Target.Digest))
		c.context.Info(fmt.Sprintf("Size: %s", units.BytesSize(float64(info.Target.Size))))

		c.context.Debug("\nCopying file ", filepath.Join(temp, name), "to", file.Name())
		err = fileHelper.CopyFile(filepath.Join(temp, name), file.Name())
		if err != nil {
			continue
		}
		ok = true
		break
	}

	if !ok {
		return "", err
	}

	return file.Name(), err
}

func (c *DockerClient) CacheGet(a *artifact.PackageArtifact) (*artifact.PackageArtifact, error) {
	resultingArtifact := a.ShallowCopy()
	// TODO:
	// Files are in URI/packagename:version (GetPackageImageName() method)
	// use downloadAndExtract .. and egenrate an archive to consume. Checksum should be already checked while downloading the image
	// with the above functions, because Docker images already contain such metadata
	// - Check how verification is done when calling DownloadArtifact outside, similarly we need to check DownloadFile, and how verification
	// is done in such cases (see repository.go)

	// We discard checksum, that are checked while during pull and unpack by containerd
	resultingArtifact.Checksums = artifact.Checksums{}

	// Check if file is already in cache
	fileName, err := c.Cache.Get(resultingArtifact)
	// Check if file is already in cache
	if err == nil {
		artifactName := path.Base(a.Path)
		c.context.Debug("Use artifact", artifactName, "from cache.")
		resultingArtifact = a
		resultingArtifact.Path = fileName
		resultingArtifact.Checksums = artifact.Checksums{}
	}

	return resultingArtifact, err
}
