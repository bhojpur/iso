package artifact_test

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
	. "github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"
	compilerspec "github.com/bhojpur/iso/pkg/manager/compiler/types/spec"
	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {
	Context("CacheID", func() {
		It("Get and retrieve files", func() {
			tmpdir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up

			tmpdirartifact, err := ioutil.TempDir(os.TempDir(), "testartifact")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdirartifact) // clean up

			err = ioutil.WriteFile(filepath.Join(tmpdirartifact, "foo"), []byte(string("foo")), os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			a := NewPackageArtifact(filepath.Join(tmpdir, "foo.tar.gz"))
			err = a.Compress(tmpdirartifact, 1)
			Expect(err).ToNot(HaveOccurred())

			cache := NewCache(tmpdir)

			// Put an artifact in the cache and retrieve it later
			// the artifact is NOT hashed so it is referenced just by the path in the cache
			_, _, err = cache.Put(a)
			Expect(err).ToNot(HaveOccurred())

			path, err := cache.Get(a)
			Expect(err).ToNot(HaveOccurred())

			b := NewPackageArtifact(path)
			ctx := context.NewContext()
			err = b.Unpack(ctx, tmpdir, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(fileHelper.Exists(filepath.Join(tmpdir, "foo"))).To(BeTrue())

			bb, err := ioutil.ReadFile(filepath.Join(tmpdir, "foo"))
			Expect(err).ToNot(HaveOccurred())

			Expect(string(bb)).To(Equal("foo"))

			// After the artifact is hashed, the fingerprint mutates so the cache doesn't see it hitting again
			// the test we did above fails as we expect to.
			a.Hash()
			_, err = cache.Get(a)
			Expect(err).To(HaveOccurred())

			a.CompileSpec = &compilerspec.BhojpurCompilationSpec{Package: &types.Package{Name: "foo", Category: "bar"}}
			_, _, err = cache.Put(a)
			Expect(err).ToNot(HaveOccurred())

			c := NewPackageArtifact(filepath.Join(tmpdir, "foo.tar.gz"))
			c.Hash()
			c.CompileSpec = &compilerspec.BhojpurCompilationSpec{Package: &types.Package{Name: "foo", Category: "bar"}}
			_, err = cache.Get(c)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
