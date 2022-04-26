package spectooling_test

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
	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	. "github.com/bhojpur/iso/pkg/manager/spectooling"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Spec Tooling", func() {
	Context("Conversion1", func() {

		b := types.NewPackage("B", "1.0", []*types.Package{}, []*types.Package{})
		c := types.NewPackage("C", "1.0", []*types.Package{}, []*types.Package{})
		d := types.NewPackage("D", "1.0", []*types.Package{}, []*types.Package{})
		p1 := types.NewPackage("A", "1.0", []*types.Package{b, c}, []*types.Package{d})
		virtual := types.NewPackage("E", "1.0", []*types.Package{}, []*types.Package{})
		virtual.SetCategory("virtual")
		p1.Provides = []*types.Package{virtual}
		p1.AddLabel("label1", "value1")
		p1.AddLabel("label2", "value2")
		p1.SetDescription("Package1")
		p1.SetCategory("cat1")
		p1.SetLicense("GPL")
		p1.AddURI("https://github.com/bhojpur/iso")
		p1.AddUse("systemd")
		It("Convert pkg1", func() {
			res := NewDefaultPackageSanitized(p1)
			expected_res := &PackageSanitized{
				Name:     "A",
				Version:  "1.0",
				Category: "cat1",
				PackageRequires: []*PackageSanitized{
					&PackageSanitized{
						Name:    "B",
						Version: "1.0",
					},
					&PackageSanitized{
						Name:    "C",
						Version: "1.0",
					},
				},
				PackageConflicts: []*PackageSanitized{
					&PackageSanitized{
						Name:    "D",
						Version: "1.0",
					},
				},
				Provides: []*PackageSanitized{
					&PackageSanitized{
						Name:     "E",
						Category: "virtual",
						Version:  "1.0",
					},
				},
				Labels: map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
				Description: "Package1",
				License:     "GPL",
				Uri:         []string{"https://github.com/bhojpur/iso"},
				UseFlags:    []string{"systemd"},
			}

			Expect(res).To(Equal(expected_res))
		})

	})
})
