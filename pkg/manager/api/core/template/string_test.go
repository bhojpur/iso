package template_test

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
	"path/filepath"

	. "github.com/bhojpur/iso/pkg/manager/api/core/template"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func writeFile(path string, content string) {
	err := ioutil.WriteFile(path, []byte(content), 0644)
	Expect(err).ToNot(HaveOccurred())
}

var _ = Describe("Templates", func() {
	Context("templates", func() {
		It("correctly templates input", func() {
			str, err := String("foo-{{.}}", "bar")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(str).Should(ContainSubstring("foo-bar"))
			Expect(len(str)).ToNot(Equal(4))

			str, err = String("foo-", nil)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(str).Should(ContainSubstring("foo-"))
			Expect(len(str)).To(Equal(4))
		})
		It("Renders templates", func() {
			out, err := Render([]string{"{{.Values.Test}}{{.Values.Bar}}"}, map[string]interface{}{"Test": "foo"}, map[string]interface{}{"Bar": "bar"})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("foobar"))
		})
		It("Renders templates with overrides", func() {
			out, err := Render([]string{"{{.Values.Test}}{{.Values.Bar}}"}, map[string]interface{}{"Test": "foo", "Bar": "baz"}, map[string]interface{}{"Bar": "bar"})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("foobar"))
		})

		It("Renders templates", func() {
			out, err := Render([]string{"{{.Values.Test}}{{.Values.Bar}}"}, map[string]interface{}{"Test": "foo", "Bar": "bar"}, map[string]interface{}{})
			Expect(err).ToNot(HaveOccurred())
			Expect(out).To(Equal("foobar"))
		})

		It("Render files default overrides", func() {
			testDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(testDir)

			toTemplate := filepath.Join(testDir, "totemplate.yaml")
			values := filepath.Join(testDir, "values.yaml")
			d := filepath.Join(testDir, "default.yaml")

			writeFile(toTemplate, `{{.Values.foo}}`)
			writeFile(values, `
foo: "bar"
`)
			writeFile(d, `
foo: "baz"
`)

			Expect(err).ToNot(HaveOccurred())

			res, err := RenderWithValues([]string{toTemplate}, values, d)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("baz"))

		})

		It("Render files from values", func() {
			testDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(testDir)

			toTemplate := filepath.Join(testDir, "totemplate.yaml")
			values := filepath.Join(testDir, "values.yaml")
			d := filepath.Join(testDir, "default.yaml")

			writeFile(toTemplate, `{{.Values.foo}}`)
			writeFile(values, `
foo: "bar"
`)
			writeFile(d, `
faa: "baz"
`)

			Expect(err).ToNot(HaveOccurred())

			res, err := RenderWithValues([]string{toTemplate}, values, d)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("bar"))

		})

		It("Render files from values if no default", func() {
			testDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(testDir)

			toTemplate := filepath.Join(testDir, "totemplate.yaml")
			values := filepath.Join(testDir, "values.yaml")

			writeFile(toTemplate, `{{.Values.foo}}`)
			writeFile(values, `
foo: "bar"
`)

			Expect(err).ToNot(HaveOccurred())

			res, err := RenderWithValues([]string{toTemplate}, values)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("bar"))
		})

		It("Render files merging defaults", func() {
			testDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(testDir)

			toTemplate := filepath.Join(testDir, "totemplate.yaml")
			values := filepath.Join(testDir, "values.yaml")
			d := filepath.Join(testDir, "default.yaml")
			d2 := filepath.Join(testDir, "default2.yaml")

			writeFile(toTemplate, `{{.Values.foo}}{{.Values.bar}}{{.Values.b}}`)
			writeFile(values, `
foo: "bar"
b: "f"
`)
			writeFile(d, `
foo: "baz"
`)

			writeFile(d2, `
foo: "do"
bar: "nei"
`)

			Expect(err).ToNot(HaveOccurred())

			res, err := RenderWithValues([]string{toTemplate}, values, d2, d)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("bazneif"))

			res, err = RenderWithValues([]string{toTemplate}, values, d, d2)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("doneif"))
		})

		It("doesn't interpolate if no one provides the values", func() {
			testDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(testDir)

			toTemplate := filepath.Join(testDir, "totemplate.yaml")
			values := filepath.Join(testDir, "values.yaml")
			d := filepath.Join(testDir, "default.yaml")

			writeFile(toTemplate, `{{if .Values.foo}}{{.Values.foo}}{{end}}`)
			writeFile(values, `
foao: "bar"
`)
			writeFile(d, `
faa: "baz"
`)

			Expect(err).ToNot(HaveOccurred())

			res, err := RenderWithValues([]string{toTemplate}, values, d)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(""))

		})

		It("correctly parses `include`", func() {
			testDir, err := ioutil.TempDir(os.TempDir(), "test")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(testDir)

			toTemplate := filepath.Join(testDir, "totemplate.yaml")
			values := filepath.Join(testDir, "values.yaml")
			d := filepath.Join(testDir, "default.yaml")

			writeFile(toTemplate, `
{{- define "app" -}}
app_name: {{if .Values.foo}}{{.Values.foo}}{{end}}
{{- end -}}
{{ include "app" . | indent 4 }}
`)
			writeFile(values, `
foo: "bar"
`)
			writeFile(d, ``)

			Expect(err).ToNot(HaveOccurred())

			res, err := RenderWithValues([]string{toTemplate}, values, d)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(`    app_name: bar
`))
		})
	})
})
