package cmd_helpers_test

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
	. "github.com/bhojpur/iso/cmd/manager/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CLI Helpers", func() {
	Context("Can parse package strings correctly", func() {
		It("accept single package names", func() {
			pack, err := ParsePackageStr("foo")
			Expect(err).ToNot(HaveOccurred())
			Expect(pack.GetName()).To(Equal("foo"))
			Expect(pack.GetCategory()).To(Equal(""))
			Expect(pack.GetVersion()).To(Equal(">=0"))
		})
		It("accept unversioned packages with category", func() {
			pack, err := ParsePackageStr("cat/foo")
			Expect(err).ToNot(HaveOccurred())
			Expect(pack.GetName()).To(Equal("foo"))
			Expect(pack.GetCategory()).To(Equal("cat"))
			Expect(pack.GetVersion()).To(Equal(">=0"))
		})
		It("accept versioned packages with category", func() {
			pack, err := ParsePackageStr("cat/foo@1.1")
			Expect(err).ToNot(HaveOccurred())
			Expect(pack.GetName()).To(Equal("foo"))
			Expect(pack.GetCategory()).To(Equal("cat"))
			Expect(pack.GetVersion()).To(Equal("1.1"))
		})
		It("accept versioned ranges with category", func() {
			pack, err := ParsePackageStr("cat/foo@>=1.1")
			Expect(err).ToNot(HaveOccurred())
			Expect(pack.GetName()).To(Equal("foo"))
			Expect(pack.GetCategory()).To(Equal("cat"))
			Expect(pack.GetVersion()).To(Equal(">=1.1"))
		})
		It("accept gentoo regex parsing without versions", func() {
			pack, err := ParsePackageStr("=cat/foo")
			Expect(err).ToNot(HaveOccurred())
			Expect(pack.GetName()).To(Equal("foo"))
			Expect(pack.GetCategory()).To(Equal("cat"))
			Expect(pack.GetVersion()).To(Equal(">=0"))
		})
		It("accept gentoo regex parsing with versions", func() {
			pack, err := ParsePackageStr("=cat/foo-1.2")
			Expect(err).ToNot(HaveOccurred())
			Expect(pack.GetName()).To(Equal("foo"))
			Expect(pack.GetCategory()).To(Equal("cat"))
			Expect(pack.GetVersion()).To(Equal("1.2"))
		})

		It("accept gentoo regex parsing with with condition", func() {
			pack, err := ParsePackageStr(">=cat/foo-1.2")
			Expect(err).ToNot(HaveOccurred())
			Expect(pack.GetName()).To(Equal("foo"))
			Expect(pack.GetCategory()).To(Equal("cat"))
			Expect(pack.GetVersion()).To(Equal(">=1.2"))
		})

		It("accept gentoo regex parsing with with condition2", func() {
			pack, err := ParsePackageStr("<cat/foo-1.2")
			Expect(err).ToNot(HaveOccurred())
			Expect(pack.GetName()).To(Equal("foo"))
			Expect(pack.GetCategory()).To(Equal("cat"))
			Expect(pack.GetVersion()).To(Equal("<1.2"))
		})

		It("accept gentoo regex parsing with with condition3", func() {
			pack, err := ParsePackageStr(">cat/foo-1.2")
			Expect(err).ToNot(HaveOccurred())
			Expect(pack.GetName()).To(Equal("foo"))
			Expect(pack.GetCategory()).To(Equal("cat"))
			Expect(pack.GetVersion()).To(Equal(">1.2"))
		})

		It("accept gentoo regex parsing with with condition4", func() {
			pack, err := ParsePackageStr("<=cat/foo-1.2")
			Expect(err).ToNot(HaveOccurred())
			Expect(pack.GetName()).To(Equal("foo"))
			Expect(pack.GetCategory()).To(Equal("cat"))
			Expect(pack.GetVersion()).To(Equal("<=1.2"))
		})
	})
})
