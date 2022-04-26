package client_test

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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	"github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"
	. "github.com/bhojpur/iso/pkg/manager/installer/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Http client", func() {
	Context("With repository", func() {
		ctx := context.NewContext()

		It("Downloads single files", func() {
			// setup small staticfile webserver with content
			tmpdir, err := ioutil.TempDir("", "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up
			Expect(err).ToNot(HaveOccurred())
			ts := httptest.NewServer(http.FileServer(http.Dir(tmpdir)))
			defer ts.Close()
			err = ioutil.WriteFile(filepath.Join(tmpdir, "test.txt"), []byte(`test`), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			c := NewHttpClient(RepoData{Urls: []string{ts.URL}}, ctx)
			path, err := c.DownloadFile("test.txt")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Read(path)).To(Equal("test"))
			os.RemoveAll(path)
		})

		It("Downloads artifacts", func() {
			// setup small staticfile webserver with content
			tmpdir, err := ioutil.TempDir("", "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up
			Expect(err).ToNot(HaveOccurred())
			ts := httptest.NewServer(http.FileServer(http.Dir(tmpdir)))
			defer ts.Close()
			err = ioutil.WriteFile(filepath.Join(tmpdir, "test.txt"), []byte(`test`), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			c := NewHttpClient(RepoData{Urls: []string{ts.URL}}, ctx)
			path, err := c.DownloadArtifact(&artifact.PackageArtifact{Path: "test.txt"})
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Read(path.Path)).To(Equal("test"))
			os.RemoveAll(path.Path)
		})

	})
})
