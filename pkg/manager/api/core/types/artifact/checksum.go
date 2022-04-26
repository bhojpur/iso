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
	//"strconv"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"sort"

	//	. "github.com/bhojpur/iso/pkg/manager/logger"
	"github.com/pkg/errors"
)

type HashImplementation string

const (
	SHA256 HashImplementation = "sha256"
)

type Checksums map[string]string

type HashOptions struct {
	Hasher hash.Hash
	Type   HashImplementation
}

func (c Checksums) List() (res [][]string) {
	keys := make([]string, 0)
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		res = append(res, []string{k, c[k]})
	}
	return
}

// Generate generates all Checksums supported for the artifact
func (c *Checksums) Generate(a *PackageArtifact) error {
	return c.generateSHA256(a)
}

func (c Checksums) Compare(d Checksums) error {
	for t, sum := range d {
		if v, ok := c[t]; ok && v != sum {
			return errors.New("Checksum mismsatch")
		}
	}
	return nil
}

func (c *Checksums) generateSHA256(a *PackageArtifact) error {
	return c.generateSum(a, HashOptions{Hasher: sha256.New(), Type: SHA256})
}

func (c *Checksums) generateSum(a *PackageArtifact, opts HashOptions) error {

	f, err := os.Open(a.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(opts.Hasher, f); err != nil {
		return err
	}

	sum := fmt.Sprintf("%x", opts.Hasher.Sum(nil))

	(*c)[string(opts.Type)] = sum
	return nil
}
