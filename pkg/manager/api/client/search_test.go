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
	. "github.com/bhojpur/iso/pkg/manager/api/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client CLI API", func() {
	Context("Reads a package tree from the Bhojpur CLI", func() {
		It("Correctly detect packages", func() {
			t, err := TreePackages("../../../tests/fixtures/alpine")
			Expect(err).ToNot(HaveOccurred())
			Expect(t).ToNot(BeNil())
			Expect(len(t.Packages)).To(Equal(1))
			Expect(t.Packages[0].Name).To(Equal("alpine"))
			Expect(t.Packages[0].Category).To(Equal("seed"))
			Expect(t.Packages[0].Version).To(Equal("1.0"))
			Expect(t.Packages[0].ImageAvailable("foo")).To(BeFalse())
			Expect(t.Packages[0].Equal(t.Packages[0])).To(BeTrue())
			Expect(t.Packages[0].Equal(Package{})).To(BeFalse())
			Expect(t.Packages[0].EqualNoV(Package{Name: "alpine", Category: "seed"})).To(BeTrue())
			Expect(t.Packages[0].EqualS("seed/alpine")).To(BeTrue())
			Expect(t.Packages[0].EqualS("seed/alpinev")).To(BeFalse())
			Expect(t.Packages[0].EqualSV("seed/alpine@1.0")).To(BeTrue())
			Expect(t.Packages[0].Image("foo")).To(Equal("foo:alpine-seed-1.0"))
			Expect(Packages(t.Packages).Exist(t.Packages[0])).To(BeTrue())
			Expect(Packages(t.Packages).Exist(Package{})).To(BeFalse())

		})
	})
})
