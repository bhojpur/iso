package artifact

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
	"crypto/sha512"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rancher-sandbox/gofilecache"
)

type ArtifactCache struct {
	gofilecache.Cache
}

func NewCache(dir string) *ArtifactCache {
	return &ArtifactCache{Cache: *gofilecache.InitCache(dir)}
}

func (c *ArtifactCache) cacheID(a *PackageArtifact) [64]byte {
	fingerprint := filepath.Base(a.Path)
	if a.CompileSpec != nil && a.CompileSpec.Package != nil {
		fingerprint = a.CompileSpec.Package.GetFingerPrint()
	}
	if len(a.Checksums) > 0 {
		for _, cs := range a.Checksums.List() {
			t := cs[0]
			result := cs[1]
			fingerprint += fmt.Sprintf("+%s:%s", t, result)
		}
	}
	return sha512.Sum512([]byte(fingerprint))
}

func (c *ArtifactCache) Get(a *PackageArtifact) (string, error) {
	fileName, _, err := c.Cache.GetFile(c.cacheID(a))
	return fileName, err
}

func (c *ArtifactCache) Put(a *PackageArtifact) (gofilecache.OutputID, int64, error) {
	file, err := os.Open(a.Path)
	if err != nil {
		return [64]byte{}, 0, errors.Wrapf(err, "failed opening %s", a.Path)
	}
	defer file.Close()
	return c.Cache.Put(c.cacheID(a), file)
}
