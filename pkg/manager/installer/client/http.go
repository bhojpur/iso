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
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"

	"github.com/cavaliercoder/grab"
)

type HttpClient struct {
	RepoData RepoData
	Cache    *artifact.ArtifactCache
	context  types.Context
}

func NewHttpClient(r RepoData, ctx types.Context) *HttpClient {
	return &HttpClient{
		RepoData: r,
		Cache:    artifact.NewCache(ctx.GetConfig().System.PkgsCachePath),
		context:  ctx,
	}
}

func NewGrabClient(timeout int) *grab.Client {
	return &grab.Client{
		UserAgent: "grab",
		HTTPClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
}

func (c *HttpClient) prepareReq(dst, url string) (*grab.Request, error) {

	req, err := grab.NewRequest(dst, url)
	if err != nil {
		return nil, err
	}

	if val, ok := c.RepoData.Authentication["token"]; ok {
		req.HTTPRequest.Header.Set("Authorization", "token "+val)
	} else if val, ok := c.RepoData.Authentication["basic"]; ok {
		req.HTTPRequest.Header.Set("Authorization", "Basic "+val)
	}

	return req, err
}

func Round(input float64) float64 {
	if input < 0 {
		return math.Ceil(input - 0.5)
	}
	return math.Floor(input + 0.5)
}

func (c *HttpClient) DownloadFile(p string) (string, error) {
	var file *os.File = nil
	var downloaded bool
	temp, err := c.context.TempDir("download")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(temp)

	client := NewGrabClient(c.context.GetConfig().General.HTTPTimeout)

	for _, uri := range c.RepoData.Urls {
		file, err = c.context.TempFile("HttpClient")
		if err != nil {
			c.context.Debug("Failed downloading", p, "from", uri)

			continue
		}
		c.context.Debug("Downloading artifact", p, "from", uri)

		u, err := url.Parse(uri)
		if err != nil {
			continue
		}
		u.Path = path.Join(u.Path, p)

		req, err := c.prepareReq(file.Name(), u.String())
		if err != nil {
			continue
		}

		resp := client.Do(req)

		// Initialize a progressbar only if we have one in the current context
		var pb *pterm.ProgressbarPrinter
		pbb := c.context.GetAnnotation("progressbar")
		switch v := pbb.(type) {
		case *pterm.ProgressbarPrinter:
			pb, _ = v.WithTotal(int(resp.Size())).WithTitle(filepath.Base(resp.Request.HTTPRequest.URL.RequestURI())).Start()
		}

		// start download loop
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()

	download_loop:

		for {
			select {
			case <-t.C:
				//	update the progress bar
				if pb != nil {
					pb.Increment().Current = int(resp.BytesComplete())
				}
			case <-resp.Done:
				//	update the progress bar
				if pb != nil {
					pb.Increment().Current = int(resp.BytesComplete())
				}
				// download is complete
				break download_loop
			}
		}

		if err = resp.Err(); err != nil {
			continue
		}

		c.context.Info("Downloaded", p, "of",
			fmt.Sprintf("%.2f", (float64(resp.BytesComplete())/1000)/1000), "MB (",
			fmt.Sprintf("%.2f", (float64(resp.BytesPerSecond())/1024)/1024), "MiB/s )")

		if pb != nil {
			// stop the progressbar if active
			pb.Stop()
		}
		//bar.Finish()
		downloaded = true
		break
	}

	if !downloaded {
		return "", errors.Wrap(err, "artifact not available in any of the specified url locations")
	}
	return file.Name(), nil
}

func (c *HttpClient) CacheGet(a *artifact.PackageArtifact) (*artifact.PackageArtifact, error) {
	newart := a.ShallowCopy()

	fileName, err := c.Cache.Get(a)

	newart.Path = fileName

	return newart, err
}

func (c *HttpClient) DownloadArtifact(a *artifact.PackageArtifact) (*artifact.PackageArtifact, error) {
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
