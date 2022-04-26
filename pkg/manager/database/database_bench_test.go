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
	"strconv"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	. "github.com/bhojpur/iso/pkg/manager/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database  benchmark", func() {

	Context("BoltDB", func() {

		a := types.NewPackage("A", ">=1.0", []*types.Package{}, []*types.Package{})

		tmpfile, _ := ioutil.TempFile(os.TempDir(), "tests")
		defer os.Remove(tmpfile.Name()) // clean up
		var db types.PackageSet

		BeforeEach(func() {

			tmpfile, _ = ioutil.TempFile(os.TempDir(), "tests")
			defer os.Remove(tmpfile.Name()) // clean up
			db = NewBoltDatabase(tmpfile.Name())
			if os.Getenv("BENCHMARK_TESTS") != "true" {
				Skip("BENCHMARK_TESTS not enabled")
			}
		})

		Measure("it should be fast in computing world from a 50000 dataset", func(b Benchmarker) {
			for i := 0; i < 50000; i++ {
				a = types.NewPackage("A"+strconv.Itoa(i), ">=1.0", []*types.Package{}, []*types.Package{})

				_, err := db.CreatePackage(a)
				Expect(err).ToNot(HaveOccurred())
			}
			runtime := b.Time("runtime", func() {
				packs := db.World()
				Expect(len(packs)).To(Equal(50000))
			})

			Ω(runtime.Seconds()).Should(BeNumerically("<", 30), "World() shouldn't take too long.")

		}, 1)

		Measure("it should be fast in computing world from a 100000 dataset", func(b Benchmarker) {
			for i := 0; i < 100000; i++ {
				a = types.NewPackage("A"+strconv.Itoa(i), ">=1.0", []*types.Package{}, []*types.Package{})

				_, err := db.CreatePackage(a)
				Expect(err).ToNot(HaveOccurred())
			}
			runtime := b.Time("runtime", func() {
				packs := db.World()
				Expect(len(packs)).To(Equal(100000))
			})

			Ω(runtime.Seconds()).Should(BeNumerically("<", 30), "World() shouldn't take too long.")

		}, 1)
	})

	Context("InMemory", func() {

		a := types.NewPackage("A", ">=1.0", []*types.Package{}, []*types.Package{})

		tmpfile, _ := ioutil.TempFile(os.TempDir(), "tests")
		defer os.Remove(tmpfile.Name()) // clean up
		var db types.PackageSet

		BeforeEach(func() {

			tmpfile, _ = ioutil.TempFile(os.TempDir(), "tests")
			defer os.Remove(tmpfile.Name()) // clean up
			db = NewInMemoryDatabase(false)
			if os.Getenv("BENCHMARK_TESTS") != "true" {
				Skip("BENCHMARK_TESTS not enabled")
			}
		})

		Measure("it should be fast in computing world from a 100000 dataset", func(b Benchmarker) {

			runtime := b.Time("runtime", func() {
				for i := 0; i < 100000; i++ {
					a = types.NewPackage("A"+strconv.Itoa(i), ">=1.0", []*types.Package{}, []*types.Package{})

					_, err := db.CreatePackage(a)
					Expect(err).ToNot(HaveOccurred())
				}
				packs := db.World()
				Expect(len(packs)).To(Equal(100000))
			})

			Ω(runtime.Seconds()).Should(BeNumerically("<", 10), "World() shouldn't take too long.")

		}, 2)
	})
})
