package image_test

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
	"path/filepath"

	"github.com/bhojpur/iso/pkg/manager/api/core/context"
	. "github.com/bhojpur/iso/pkg/manager/api/core/image"
	"github.com/bhojpur/iso/pkg/manager/helpers/file"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {

	ctx := context.NewContext()
	Context("used as k/v store", func() {

		cache := &Cache{}
		var dir string

		BeforeEach(func() {
			ctx = context.NewContext()
			var err error
			dir, err = ctx.TempDir("foo")
			Expect(err).ToNot(HaveOccurred())
			cache = NewCache(dir, 10*1024*1024, 1) // 10MB Cache when upgrading to files. Max volatile memory of 1 row.
		})

		AfterEach(func() {
			cache.Clean()
		})

		It("does handle automatically memory upgrade", func() {
			cache.Set("foo", "bar")
			v, found := cache.Get("foo")
			Expect(found).To(BeTrue())
			Expect(v).To(Equal("bar"))
			Expect(file.Exists(filepath.Join(dir, "foo"))).To(BeFalse())
			cache.Set("baz", "bar")
			Expect(file.Exists(filepath.Join(dir, "foo"))).To(BeTrue())
			Expect(file.Exists(filepath.Join(dir, "baz"))).To(BeTrue())
			v, found = cache.Get("foo")
			Expect(found).To(BeTrue())
			Expect(v).To(Equal("bar"))

			Expect(cache.Count()).To(Equal(2))
		})

		It("does CRUD", func() {
			cache.Set("foo", "bar")

			v, found := cache.Get("foo")
			Expect(found).To(BeTrue())
			Expect(v).To(Equal("bar"))

			hit := false
			cache.All(func(c CacheResult) {
				hit = true
				Expect(c.Key()).To(Equal("foo"))
				Expect(c.Value()).To(Equal("bar"))
			})
			Expect(hit).To(BeTrue())

		})

		It("Unmarshals values", func() {
			type testStruct struct {
				Test string
			}

			cache.SetValue("foo", &testStruct{Test: "baz"})

			n := &testStruct{}

			cache.All(func(cr CacheResult) {
				err := cr.Unmarshal(n)
				Expect(err).ToNot(HaveOccurred())

			})
			Expect(n.Test).To(Equal("baz"))
		})
	})
})
