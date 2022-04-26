package database_test

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
	"regexp"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	. "github.com/bhojpur/iso/pkg/manager/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BoltDB Database", func() {

	tmpfile, _ := ioutil.TempFile(os.TempDir(), "tests")
	defer os.Remove(tmpfile.Name()) // clean up
	var db types.PackageDatabase

	BeforeEach(func() {
		tmpfile, _ = ioutil.TempFile(os.TempDir(), "tests")
		defer os.Remove(tmpfile.Name()) // clean up
		db = NewBoltDatabase(tmpfile.Name())
	})
	Context("Simple package", func() {
		a := types.NewPackage("A", ">=1.0", []*types.Package{}, []*types.Package{})

		It("Find packages", func() {
			ID, err := db.CreatePackage(a)
			Expect(err).ToNot(HaveOccurred())

			pack, err := db.GetPackage(ID)
			Expect(err).ToNot(HaveOccurred())

			Expect(pack).To(Equal(a))
			ids := db.GetPackages()

			Expect(ids).To(Equal([]string{"1"}))

			pack, err = db.FindPackage(a)
			Expect(err).ToNot(HaveOccurred())
			Expect(pack).To(Equal(a))

		})

		It("Find package files", func() {
			a := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})
			a1 := types.NewPackage("A", "1.1", []*types.Package{}, []*types.Package{})
			a3 := types.NewPackage("A", "1.3", []*types.Package{}, []*types.Package{})
			_, err := db.CreatePackage(a)
			Expect(err).ToNot(HaveOccurred())

			_, err = db.CreatePackage(a1)
			Expect(err).ToNot(HaveOccurred())

			_, err = db.CreatePackage(a3)
			Expect(err).ToNot(HaveOccurred())

			err = db.SetPackageFiles(&types.PackageFile{PackageFingerprint: a.GetFingerPrint(), Files: []string{"foo"}})
			Expect(err).ToNot(HaveOccurred())

			err = db.SetPackageFiles(&types.PackageFile{PackageFingerprint: a1.GetFingerPrint(), Files: []string{"bar"}})
			Expect(err).ToNot(HaveOccurred())

			pack, err := db.FindPackageByFile("fo")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(pack)).To(Equal(1))
			Expect(pack[0]).To(Equal(a))
		})

		It("Expands correctly", func() {

			a := types.NewPackage("A", ">=1.0", []*types.Package{}, []*types.Package{})
			a1 := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})
			a11 := types.NewPackage("A", "1.1", []*types.Package{}, []*types.Package{})
			a01 := types.NewPackage("A", "0.1", []*types.Package{}, []*types.Package{})
			re := regexp.MustCompile("project[0-9][=].*")
			for _, p := range []*types.Package{a1, a11, a01} {
				_, err := db.CreatePackage(p)
				Expect(err).ToNot(HaveOccurred())
			}
			lst, err := a.Expand(db)
			Expect(err).ToNot(HaveOccurred())
			Expect(lst).To(ContainElement(a11))
			Expect(lst).To(ContainElement(a1))
			Expect(lst).ToNot(ContainElement(a01))
			Expect(len(lst)).To(Equal(2))
			p := lst.Best(nil)
			Expect(p).To(Equal(a11))
			// Test annotation with null map
			Expect(a.MatchAnnotation(re)).To(Equal(false))
		})

		It("Find best package candidate", func() {
			a := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})
			a1 := types.NewPackage("A", "1.1", []*types.Package{}, []*types.Package{})
			a3 := types.NewPackage("A", "1.3", []*types.Package{}, []*types.Package{})
			_, err := db.CreatePackage(a)
			Expect(err).ToNot(HaveOccurred())

			_, err = db.CreatePackage(a1)
			Expect(err).ToNot(HaveOccurred())

			_, err = db.CreatePackage(a3)
			Expect(err).ToNot(HaveOccurred())
			s := types.NewPackage("A", ">=1.0", []*types.Package{}, []*types.Package{})

			pack, err := db.FindPackageCandidate(s)
			Expect(err).ToNot(HaveOccurred())
			Expect(pack).To(Equal(a3))

		})

		It("Find specific package candidate", func() {
			a := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})
			a1 := types.NewPackage("A", "1.1", []*types.Package{}, []*types.Package{})
			a3 := types.NewPackage("A", "1.3", []*types.Package{}, []*types.Package{})
			_, err := db.CreatePackage(a)
			Expect(err).ToNot(HaveOccurred())

			_, err = db.CreatePackage(a1)
			Expect(err).ToNot(HaveOccurred())

			_, err = db.CreatePackage(a3)
			Expect(err).ToNot(HaveOccurred())
			s := types.NewPackage("A", "=1.0", []*types.Package{}, []*types.Package{})

			pack, err := db.FindPackageCandidate(s)
			Expect(err).ToNot(HaveOccurred())
			Expect(pack).To(Equal(a))

		})

		It("Provides replaces definitions", func() {
			a := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})
			a1 := types.NewPackage("A", "1.1", []*types.Package{}, []*types.Package{})
			a3 := types.NewPackage("A", "1.3", []*types.Package{}, []*types.Package{})

			a3.SetProvides([]*types.Package{{Name: "A", Category: "", Version: "1.0"}})
			Expect(a3.GetProvides()).To(Equal([]*types.Package{{Name: "A", Category: "", Version: "1.0"}}))

			_, err := db.CreatePackage(a)
			Expect(err).ToNot(HaveOccurred())

			_, err = db.CreatePackage(a1)
			Expect(err).ToNot(HaveOccurred())

			_, err = db.CreatePackage(a3)
			Expect(err).ToNot(HaveOccurred())

			s := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})

			pack, err := db.FindPackage(s)
			Expect(err).ToNot(HaveOccurred())
			Expect(pack).To(Equal(a3))
		})

		Context("Provides", func() {

			It("replaces definitions", func() {
				a := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})
				a1 := types.NewPackage("A", "1.1", []*types.Package{}, []*types.Package{})
				a3 := types.NewPackage("A", "1.3", []*types.Package{}, []*types.Package{})

				a3.SetProvides([]*types.Package{{Name: "A", Category: "", Version: "1.0"}})
				Expect(a3.GetProvides()).To(Equal([]*types.Package{{Name: "A", Category: "", Version: "1.0"}}))

				_, err := db.CreatePackage(a)
				Expect(err).ToNot(HaveOccurred())

				_, err = db.CreatePackage(a1)
				Expect(err).ToNot(HaveOccurred())

				_, err = db.CreatePackage(a3)
				Expect(err).ToNot(HaveOccurred())

				s := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})

				pack, err := db.FindPackage(s)
				Expect(err).ToNot(HaveOccurred())
				Expect(pack).To(Equal(a3))
			})

			It("replaces definitions", func() {
				a := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})
				a1 := types.NewPackage("A", "1.1", []*types.Package{}, []*types.Package{})
				a3 := types.NewPackage("A", "1.3", []*types.Package{}, []*types.Package{})

				a3.SetProvides([]*types.Package{{Name: "A", Category: "", Version: "1.0"}})
				Expect(a3.GetProvides()).To(Equal([]*types.Package{{Name: "A", Category: "", Version: "1.0"}}))

				_, err := db.CreatePackage(a)
				Expect(err).ToNot(HaveOccurred())

				_, err = db.CreatePackage(a1)
				Expect(err).ToNot(HaveOccurred())

				_, err = db.CreatePackage(a3)
				Expect(err).ToNot(HaveOccurred())

				s := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})

				packs, err := db.FindPackages(s)
				Expect(err).ToNot(HaveOccurred())
				Expect(packs).To(ContainElement(a3))
			})

			It("replaces definitions", func() {
				a := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})
				a1 := types.NewPackage("A", "1.1", []*types.Package{}, []*types.Package{})
				z := types.NewPackage("Z", "1.3", []*types.Package{}, []*types.Package{})

				z.SetProvides([]*types.Package{{Name: "A", Category: "", Version: ">=1.0"}})
				Expect(z.GetProvides()).To(Equal([]*types.Package{{Name: "A", Category: "", Version: ">=1.0"}}))

				_, err := db.CreatePackage(a)
				Expect(err).ToNot(HaveOccurred())

				_, err = db.CreatePackage(a1)
				Expect(err).ToNot(HaveOccurred())

				_, err = db.CreatePackage(z)
				Expect(err).ToNot(HaveOccurred())

				s := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})

				packs, err := db.FindPackages(s)
				Expect(err).ToNot(HaveOccurred())
				Expect(packs).To(ContainElement(z))
			})

			It("replaces definitions of unexisting packages", func() {
				a1 := types.NewPackage("A", "1.1", []*types.Package{}, []*types.Package{})
				z := types.NewPackage("Z", "1.3", []*types.Package{}, []*types.Package{})

				z.SetProvides([]*types.Package{{Name: "A", Category: "", Version: ">=1.0"}})
				Expect(z.GetProvides()).To(Equal([]*types.Package{{Name: "A", Category: "", Version: ">=1.0"}}))

				_, err := db.CreatePackage(a1)
				Expect(err).ToNot(HaveOccurred())

				_, err = db.CreatePackage(z)
				Expect(err).ToNot(HaveOccurred())

				s := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})

				packs, err := db.FindPackages(s)
				Expect(err).ToNot(HaveOccurred())
				Expect(packs).To(ContainElement(z))
			})

			It("replaces definitions of a required package", func() {

				c := types.NewPackage("C", "1.1", []*types.Package{{Name: "A", Category: "", Version: ">=0"}}, []*types.Package{})
				z := types.NewPackage("Z", "1.3", []*types.Package{}, []*types.Package{})

				z.SetProvides([]*types.Package{{Name: "A", Category: "", Version: ">=1.0"}})
				Expect(z.GetProvides()).To(Equal([]*types.Package{{Name: "A", Category: "", Version: ">=1.0"}}))

				_, err := db.CreatePackage(z)
				Expect(err).ToNot(HaveOccurred())
				_, err = db.CreatePackage(c)
				Expect(err).ToNot(HaveOccurred())

				s := types.NewPackage("A", "1.0", []*types.Package{}, []*types.Package{})

				packs, err := db.FindPackages(s)
				Expect(err).ToNot(HaveOccurred())
				Expect(packs).To(ContainElement(z))
			})

			When("Searching with selectors", func() {
				It("replaces definitions of a required package", func() {

					c := types.NewPackage("C", "1.1", []*types.Package{{Name: "A", Category: "", Version: ">=0"}}, []*types.Package{})
					z := types.NewPackage("Z", "1.3", []*types.Package{}, []*types.Package{})

					z.SetProvides([]*types.Package{{Name: "A", Category: "", Version: ">=1.0"}})
					Expect(z.GetProvides()).To(Equal([]*types.Package{{Name: "A", Category: "", Version: ">=1.0"}}))

					_, err := db.CreatePackage(z)
					Expect(err).ToNot(HaveOccurred())
					_, err = db.CreatePackage(c)
					Expect(err).ToNot(HaveOccurred())

					s := types.NewPackage("A", ">=1.0", []*types.Package{}, []*types.Package{})

					packs, err := db.FindPackages(s)
					Expect(err).ToNot(HaveOccurred())
					Expect(packs).To(ContainElement(z))
				})
			})

		})

	})

})
