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
	"archive/tar"
	"io"
	"regexp"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
)

func compileRegexes(regexes []string) []*regexp.Regexp {
	var result []*regexp.Regexp
	for _, i := range regexes {
		r, e := regexp.Compile(i)
		if e != nil {
			continue
		}
		result = append(result, r)
	}
	return result
}

type ImageDiffNode struct {
	Name string `json:"Name"`
	Size int    `json:"Size"`
}
type ImageDiff struct {
	Additions []ImageDiffNode `json:"Adds"`
	Deletions []ImageDiffNode `json:"Dels"`
	Changes   []ImageDiffNode `json:"Mods"`
}

func Delta(srcimg, dstimg v1.Image) (res ImageDiff, err error) {
	srcReader := mutate.Extract(srcimg)
	defer srcReader.Close()

	dstReader := mutate.Extract(dstimg)
	defer dstReader.Close()

	filesSrc, filesDst := map[string]int64{}, map[string]int64{}

	srcTar := tar.NewReader(srcReader)
	dstTar := tar.NewReader(dstReader)

	for {
		var hdr *tar.Header
		hdr, err = srcTar.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return
		}
		filesSrc[hdr.Name] = hdr.Size
	}

	for {
		var hdr *tar.Header
		hdr, err = dstTar.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return
		}
		filesDst[hdr.Name] = hdr.Size
	}
	err = nil

	for f, size := range filesDst {
		if size2, exist := filesSrc[f]; exist && size2 != size {
			res.Changes = append(res.Changes, ImageDiffNode{
				Name: f,
				Size: int(size),
			})
		} else if !exist {
			res.Additions = append(res.Additions, ImageDiffNode{
				Name: f,
				Size: int(size),
			})
		}
	}
	for f, size := range filesSrc {
		if _, exist := filesDst[f]; !exist {
			res.Deletions = append(res.Deletions, ImageDiffNode{
				Name: f,
				Size: int(size),
			})
		}
	}

	return
}
