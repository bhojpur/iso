package installer_test

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
	//	. "github.com/bhojpur/iso/pkg/manager/installer"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	pkg "github.com/bhojpur/iso/pkg/manager/database"
	. "github.com/bhojpur/iso/pkg/manager/installer"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("System", func() {
	Context("Files", func() {
		var s *System
		var db types.PackageDatabase
		var a, b *types.Package
		ctx := context.NewContext()

		BeforeEach(func() {
			db = pkg.NewInMemoryDatabase(false)
			s = &System{Database: db}

			a = &types.Package{Name: "test", Version: "1", Category: "t"}

			db.CreatePackage(a)
			db.SetPackageFiles(&types.PackageFile{PackageFingerprint: a.GetFingerPrint(), Files: []string{"foo", "f"}})

			b = &types.Package{Name: "test2", Version: "1", Category: "t"}

			db.CreatePackage(b)
			db.SetPackageFiles(&types.PackageFile{PackageFingerprint: b.GetFingerPrint(), Files: []string{"barz", "f"}})
		})

		It("detects when are already shipped by other packages", func() {
			r, p, err := s.ExistsPackageFile("foo")
			Expect(r).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
			Expect(p).To(Equal(a))
			r, p, err = s.ExistsPackageFile("baz")
			Expect(r).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
			Expect(p).To(BeNil())

			r, p, err = s.ExistsPackageFile("f")
			Expect(r).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
			Expect(p).To(Or(Equal(b), Equal(a))) // This fails
			r, p, err = s.ExistsPackageFile("barz")
			Expect(r).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
			Expect(p).To(Equal(b))
		})

		It("detect missing files", func() {
			dir, err := ioutil.TempDir("", "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(dir)
			s.Target = dir
			notfound := s.OSCheck(ctx)
			Expect(len(notfound)).To(Equal(2))
			ioutil.WriteFile(filepath.Join(dir, "f"), []byte{}, os.ModePerm)
			ioutil.WriteFile(filepath.Join(dir, "foo"), []byte{}, os.ModePerm)
			notfound = s.OSCheck(ctx)
			Expect(len(notfound)).To(Equal(1))
		})
	})
})
