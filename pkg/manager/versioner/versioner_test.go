package version_test

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
	"fmt"

	. "github.com/bhojpur/iso/pkg/manager/versioner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Versioner", func() {
	Context("Invalid version", func() {
		versioner := DefaultVersioner()
		It("Sanitize", func() {
			sanitized := versioner.Sanitize("foo_bar")
			Expect(sanitized).Should(Equal("foo-bar"))
		})
	})

	Context("valid version", func() {
		versioner := DefaultVersioner()
		It("Validate", func() {
			err := versioner.Validate("1.0")
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("invalid version", func() {
		versioner := DefaultVersioner()
		It("Validate", func() {
			err := versioner.Validate("1.0_##")
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("Sorting", func() {
		versioner := DefaultVersioner()
		It("finds the correct ordering", func() {
			sorted := versioner.Sort([]string{"1.0", "0.1"})
			Expect(sorted).Should(Equal([]string{"0.1", "1.0"}))
		})
	})

	Context("Sorting with invalid characters", func() {
		versioner := DefaultVersioner()
		It("finds the correct ordering", func() {
			sorted := versioner.Sort([]string{"1.0_1", "0.1"})
			Expect(sorted).Should(Equal([]string{"0.1", "1.0_1"}))
		})
	})

	Context("Complex Sorting", func() {
		versioner := DefaultVersioner()
		It("finds the correct ordering", func() {
			sorted := versioner.Sort([]string{"1.0", "0.1", "0.22", "1.1", "1.9", "1.10", "11.1"})
			Expect(sorted).Should(Equal([]string{"0.1", "0.22", "1.0", "1.1", "1.9", "1.10", "11.1"}))
		})
	})

	Context("Sorting with +", func() {
		versioner := DefaultVersioner()
		It("finds the correct ordering", func() {
			sorted := versioner.Sort([]string{"1.0+1", "1.0+0", "0.1", "1.0+3", "1.0+2", "1.9"})
			Expect(sorted).Should(Equal([]string{"0.1", "1.0+0", "1.0+1", "1.0+2", "1.0+3", "1.9"}))
		})
	})

	// from: https://github.com/knqyf263/go-deb-version/blob/master/version_test.go#L8
	Context("Debian Sorting", func() {
		versioner := DefaultVersioner()
		It("finds the correct ordering", func() {
			sorted := versioner.Sort([]string{"2:7.4.052-1ubuntu3.1", "2:7.4.052-1ubuntu1", "2:7.4.052-1ubuntu2", "2:7.4.052-1ubuntu3"})
			Expect(sorted).Should(Equal([]string{"2:7.4.052-1ubuntu1", "2:7.4.052-1ubuntu2", "2:7.4.052-1ubuntu3", "2:7.4.052-1ubuntu3.1"}))
		})
	})

	It("finds the correct ordering", func() {
		versioner := DefaultVersioner()

		sorted := versioner.Sort([]string{"0.0.1-beta-9", "0.0.1-alpha08-9", "0.0.1-alpha07-9", "0.0.1-alpha07-8"})
		Expect(sorted).Should(Equal([]string{"0.0.1-alpha07-8", "0.0.1-alpha07-9", "0.0.1-alpha08-9", "0.0.1-beta-9"}))
	})
	It("finds the correct ordering", func() {
		versioner := DefaultVersioner()

		sorted := versioner.Sort([]string{"0.0.1-beta01", "0.0.1-alpha08", "0.0.1-alpha07"})
		Expect(sorted).Should(Equal([]string{"0.0.1-alpha07", "0.0.1-alpha08", "0.0.1-beta01"}))
	})

	Context("Matching a selector", func() {

		testCases := [][]string{
			{">=1", "2"},
			{"<=3", "2"},
			{">0", ""},
			{">0", "0.0.40-alpha"},
			{">=0.1.0+0.4", "0.1.0+0.5"},
			{">=0.0.20190406.4.9.172-r1", "1.0.111"},
			{">=0", "1.0.29+pre2_p20191024"},
			{">=0.1.0+4", "0.1.0+5"},
			{">0.1.0-4", "0.1.0-5"},
			{"<1.2.3-beta", "1.2.3-beta.1-1"},
			{"<1.2.3", "1.2.3-beta.1"},
			{">0.0.1-alpha07", "0.0.1-alpha07-8"},
			{">0.0.1-alpha07-1", "0.0.1-alpha07-8"},
			{">0.0.1-alpha07", "0.0.1-alpha08"},
		}
		versioner := DefaultVersioner()

		for i, t := range testCases {
			selector := testCases[i][0]
			version := testCases[i][1]
			It(fmt.Sprint(t), func() {

				Expect(versioner.ValidateSelector(version, selector)).Should(BeTrue())
			})
		}
	})

	Context("Not matching a selector", func() {

		testfalseCases := [][]string{
			{">0.0.1-alpha07", "0.0.1-alpha06"},
			{"<0.0.1-alpha07", "0.0.1-alpha08"},
			{">0.1.0+0.4", "0.1.0+0.3"},
			{">=0.0.20190406.4.9.172-r1", "0"},
			{"<=1", "2"},
			{">=3", "2"},
			{"<0", "0.0.40-alpha"},
			{"<0.1.0+0.4", "0.1.0+0.5"},
			{"<=0.0.20190406.4.9.172-r1", "1.0.111"},
			{"<0.1.0+4", "0.1.0+5"},
			{"<=0.1.0-4", "0.1.0-5"},
		}

		versioner := DefaultVersioner()

		for i, t := range testfalseCases {
			selector := testfalseCases[i][0]
			version := testfalseCases[i][1]
			It(fmt.Sprint(t), func() {
				Expect(versioner.ValidateSelector(version, selector)).Should(BeFalse())
			})
		}
	})

})
