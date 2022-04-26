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
	"bufio"
	"context"
	"io"
	"os"

	archive "github.com/containerd/containerd/archive"
	dockerarchive "github.com/docker/docker/pkg/archive"
	docker "github.com/fsouza/go-dockerclient"
	layer "github.com/opencontainers/umoci/oci/layer"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/urfave/cli"
)

type ExtractOpts struct {
	Source, Destination                string
	Compressed, KeepDirlinks, Rootless bool
	UnpackMode                         string
}

func ExtractLayer(opts *ExtractOpts) error {
	file, err := os.Open(opts.Source)
	if err != nil {
		return err
	}
	var r io.Reader
	r = file

	if opts.Compressed {
		decompressedArchive, err := dockerarchive.DecompressStream(bufio.NewReader(file))
		if err != nil {
			return err
		}
		defer decompressedArchive.Close()
		r = decompressedArchive
	}

	buf := bufio.NewReader(r)
	switch opts.UnpackMode {
	case "umoci": // more fixes are in there
		return layer.UnpackLayer(opts.Destination, buf, &layer.UnpackOptions{KeepDirlinks: opts.KeepDirlinks, MapOptions: layer.MapOptions{Rootless: opts.Rootless}})
	case "containerd": // more cross-compatible
		_, err := archive.Apply(context.Background(), opts.Destination, buf)
		return err
	default: // moby way
		return Untar(buf, opts.Destination, !opts.Compressed)
	}
}

// PullImage pull the specified image
func PullImage(client *docker.Client, image string) error {
	var err error
	// Pulling the image
	jww.INFO.Printf("Pulling the docker image %s\n", image)
	if err = client.PullImage(docker.PullImageOptions{Repository: image}, docker.AuthConfiguration{}); err != nil {
		jww.ERROR.Printf("error pulling %s image: %s\n", image, err)
		return err
	}

	jww.INFO.Println("Image", image, "pulled correctly")

	return nil
}

// NewDocker Creates a new instance of *docker.Client, respecting env settings
func NewDocker() (*docker.Client, error) {
	var err error
	var client *docker.Client
	if os.Getenv("DOCKER_SOCKET") != "" {
		client, err = docker.NewClient(os.Getenv("DOCKER_SOCKET"))
	} else {
		client, err = docker.NewClient("unix:///var/run/docker.sock")
	}
	if err != nil {
		return nil, cli.NewExitError("could not connect to the Docker daemon", 87)
	}
	return client, nil
}