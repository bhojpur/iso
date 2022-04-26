package types_test

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
	"strings"

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	Context("Inits paths", func() {
		t, _ := ioutil.TempDir("", "tests")
		defer os.RemoveAll(t)
		c := &types.BhojpurConfig{
			System: types.BhojpurSystemConfig{
				Rootfs:        t,
				PkgsCachePath: "foo",
				DatabasePath:  "baz",
			}}
		It("sets default", func() {
			err := c.Init()
			Expect(err).ToNot(HaveOccurred())
			Expect(c.System.Rootfs).To(Equal(t))
			Expect(c.System.PkgsCachePath).To(Equal(filepath.Join(t, "baz", "foo")))
			Expect(c.System.DatabasePath).To(Equal(filepath.Join(t, "baz")))
		})
	})

	Context("Load Repository1", func() {
		var ctx *context.Context
		BeforeEach(func() {
			ctx = context.NewContext(context.WithConfig(&types.BhojpurConfig{
				RepositoriesConfDir: []string{
					"../../../../tests/fixtures/repos.conf.d",
				},
			}))
			ctx.Config.Init()
		})

		It("Check Load Repository 1", func() {
			Expect(len(ctx.GetConfig().SystemRepositories)).Should(Equal(2))
			Expect(ctx.GetConfig().SystemRepositories[0].Name).Should(Equal("test1"))
			Expect(ctx.GetConfig().SystemRepositories[0].Priority).Should(Equal(999))
			Expect(ctx.GetConfig().SystemRepositories[0].Type).Should(Equal("disk"))
			Expect(len(ctx.GetConfig().SystemRepositories[0].Urls)).Should(Equal(1))
			Expect(ctx.GetConfig().SystemRepositories[0].Urls[0]).Should(Equal("tests/repos/test1"))
		})

		It("Chec Load Repository 2", func() {
			Expect(len(ctx.GetConfig().SystemRepositories)).Should(Equal(2))
			Expect(ctx.GetConfig().SystemRepositories[1].Name).Should(Equal("test2"))
			Expect(ctx.GetConfig().SystemRepositories[1].Priority).Should(Equal(1000))
			Expect(ctx.GetConfig().SystemRepositories[1].Type).Should(Equal("disk"))
			Expect(len(ctx.GetConfig().SystemRepositories[1].Urls)).Should(Equal(1))
			Expect(ctx.GetConfig().SystemRepositories[1].Urls[0]).Should(Equal("tests/repos/test2"))
		})
	})

	Context("Simple temporary directory creation", func() {
		ctx := context.NewContext(context.WithConfig(&types.BhojpurConfig{
			System: types.BhojpurSystemConfig{
				TmpDirBase: os.TempDir() + "/tmpiso",
			},
		}))

		BeforeEach(func() {
			ctx = context.NewContext(context.WithConfig(&types.BhojpurConfig{
				System: types.BhojpurSystemConfig{
					TmpDirBase: os.TempDir() + "/tmpiso",
				},
			}))

		})

		It("Create Temporary directory", func() {
			tmpDir, err := ctx.TempDir("test1")
			Expect(err).ToNot(HaveOccurred())
			Expect(strings.HasPrefix(tmpDir, filepath.Join(os.TempDir(), "tmpiso"))).To(BeTrue())
			Expect(fileHelper.Exists(tmpDir)).To(BeTrue())

			defer os.RemoveAll(tmpDir)
		})

		It("Create Temporary file", func() {
			tmpFile, err := ctx.TempFile("testfile1")
			Expect(err).ToNot(HaveOccurred())
			Expect(strings.HasPrefix(tmpFile.Name(), filepath.Join(os.TempDir(), "tmpiso"))).To(BeTrue())
			Expect(fileHelper.Exists(tmpFile.Name())).To(BeTrue())

			defer os.Remove(tmpFile.Name())
		})

	})

})
