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
	"os"
	"path/filepath"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	"github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	compilerspec "github.com/bhojpur/iso/pkg/manager/compiler/types/spec"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	. "github.com/bhojpur/iso/pkg/manager/installer/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This test expect that the repository defined in UNIT_TEST_DOCKER_IMAGE is in zstd format.
// the repository is built by the 01_simple_docker.sh integration test fileHelper.
// This test also require root. At the moment, unpacking docker images with 'img' requires root permission to
// mount/unmount layers.
var _ = Describe("Docker client", func() {
	Context("With repository", func() {
		ctx := context.NewContext()

		repoImage := os.Getenv("UNIT_TEST_DOCKER_IMAGE")
		var repoURL []string
		var c *DockerClient
		BeforeEach(func() {
			if repoImage == "" {
				Skip("UNIT_TEST_DOCKER_IMAGE not specified")
			}
			repoURL = []string{repoImage}
			c = NewDockerClient(RepoData{Urls: repoURL}, ctx)
		})

		It("Downloads single files", func() {
			f, err := c.DownloadFile("repository.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileHelper.Read(f)).To(ContainSubstring("Test Repo"))
			os.RemoveAll(f)
		})

		It("Downloads artifacts", func() {
			f, err := c.DownloadArtifact(&artifact.PackageArtifact{
				Path: "test.tar",
				CompileSpec: &compilerspec.BhojpurCompilationSpec{
					Package: &types.Package{
						Name:     "c",
						Category: "test",
						Version:  "1.0",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())
			tmpdir, err := ioutil.TempDir("", "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			Expect(f.Unpack(ctx, tmpdir, false)).ToNot(HaveOccurred())
			Expect(fileHelper.Read(filepath.Join(tmpdir, "c"))).To(Equal("c\n"))
			Expect(fileHelper.Read(filepath.Join(tmpdir, "cd"))).To(Equal("c\n"))
			os.RemoveAll(f.Path)
		})
	})
})
