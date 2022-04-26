package config_test

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
	config "github.com/bhojpur/iso/pkg/manager/api/core/config"
	"github.com/bhojpur/iso/pkg/manager/api/core/context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	Context("Test config protect", func() {

		It("Protect1", func() {
			ctx := context.NewContext()
			files := []string{
				"etc/foo/my.conf",
				"usr/bin/foo",
				"usr/share/doc/foo.md",
			}

			cp := config.NewConfigProtect("/etc")
			cp.Map(files, ctx.Config.ConfigProtectConfFiles)

			Expect(cp.Protected("etc/foo/my.conf")).To(BeTrue())
			Expect(cp.Protected("/etc/foo/my.conf")).To(BeTrue())
			Expect(cp.Protected("usr/bin/foo")).To(BeFalse())
			Expect(cp.Protected("/usr/bin/foo")).To(BeFalse())
			Expect(cp.Protected("/usr/share/doc/foo.md")).To(BeFalse())

			Expect(cp.GetProtectFiles(false)).To(Equal(
				[]string{
					"etc/foo/my.conf",
				},
			))

			Expect(cp.GetProtectFiles(true)).To(Equal(
				[]string{
					"/etc/foo/my.conf",
				},
			))
		})

		It("Protect2", func() {
			ctx := context.NewContext()

			files := []string{
				"etc/foo/my.conf",
				"usr/bin/foo",
				"usr/share/doc/foo.md",
			}

			cp := config.NewConfigProtect("")
			cp.Map(files, ctx.Config.ConfigProtectConfFiles)

			Expect(cp.Protected("etc/foo/my.conf")).To(BeFalse())
			Expect(cp.Protected("/etc/foo/my.conf")).To(BeFalse())
			Expect(cp.Protected("usr/bin/foo")).To(BeFalse())
			Expect(cp.Protected("/usr/bin/foo")).To(BeFalse())
			Expect(cp.Protected("/usr/share/doc/foo.md")).To(BeFalse())

			Expect(cp.GetProtectFiles(false)).To(Equal(
				[]string{},
			))

			Expect(cp.GetProtectFiles(true)).To(Equal(
				[]string{},
			))
		})

		It("Protect3: Annotation dir without initial slash", func() {
			ctx := context.NewContext()

			files := []string{
				"etc/foo/my.conf",
				"usr/bin/foo",
				"usr/share/doc/foo.md",
			}

			cp := config.NewConfigProtect("etc")
			cp.Map(files, ctx.Config.ConfigProtectConfFiles)

			Expect(cp.Protected("etc/foo/my.conf")).To(BeTrue())
			Expect(cp.Protected("/etc/foo/my.conf")).To(BeTrue())
			Expect(cp.Protected("usr/bin/foo")).To(BeFalse())
			Expect(cp.Protected("/usr/bin/foo")).To(BeFalse())
			Expect(cp.Protected("/usr/share/doc/foo.md")).To(BeFalse())

			Expect(cp.GetProtectFiles(false)).To(Equal(
				[]string{
					"etc/foo/my.conf",
				},
			))

			Expect(cp.GetProtectFiles(true)).To(Equal(
				[]string{
					"/etc/foo/my.conf",
				},
			))
		})

	})

})
