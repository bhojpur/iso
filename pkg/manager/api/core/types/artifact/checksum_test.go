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

	. "github.com/bhojpur/iso/pkg/manager/api/core/types/artifact"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Checksum", func() {
	Context("Generation", func() {
		It("Compares successfully", func() {

			tmpdir, err := ioutil.TempDir("", "tree")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpdir) // clean up
			buildsum := Checksums{}
			definitionsum := Checksums{}
			definitionsum2 := Checksums{}

			Expect(len(buildsum)).To(Equal(0))
			Expect(len(definitionsum)).To(Equal(0))
			Expect(len(definitionsum2)).To(Equal(0))

			err = buildsum.Generate(NewPackageArtifact("../../../../../tests/fixtures/layers/alpine/build.yaml"))
			Expect(err).ToNot(HaveOccurred())

			err = definitionsum.Generate(NewPackageArtifact("../../../../../tests/fixtures/layers/alpine/definition.yaml"))
			Expect(err).ToNot(HaveOccurred())

			err = definitionsum2.Generate(NewPackageArtifact("../../../../../tests/fixtures/layers/alpine/definition.yaml"))
			Expect(err).ToNot(HaveOccurred())

			Expect(len(buildsum)).To(Equal(1))
			Expect(len(definitionsum)).To(Equal(1))
			Expect(len(definitionsum2)).To(Equal(1))

			Expect(definitionsum.Compare(buildsum)).To(HaveOccurred())
			Expect(definitionsum.Compare(definitionsum2)).ToNot(HaveOccurred())
		})
	})

})
