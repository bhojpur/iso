package template

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
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	fileHelper "github.com/bhojpur/iso/pkg/manager/helpers/file"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/imdario/mergo"
)

// String templates a string with the interface
func String(t string, i interface{}) (string, error) {
	b := bytes.NewBuffer([]byte{})

	f := funcMap()

	tmpl := template.New("")

	includedNames := make(map[string]int)

	// Add the 'include' function here so we can close over tmpl.
	f["include"] = includeTemplate(tmpl, includedNames)

	tmpl, err := tmpl.Funcs(f).Parse(t)
	if err != nil {
		return "", err
	}

	err = tmpl.Option("missingkey=zero").Execute(b, i)

	return b.String(), err
}

// Render renders the template string like helm
func Render(strings []string, values, d map[string]interface{}) (string, error) {

	// We slurp all the files into one here. This is not elegant, but still works.
	// As a reminder, the files passed here have on the head the templates in the 'templates/' folder
	// of each Bhojpur ISO tree, and it have at the bottom the package buildpsec to be templated.
	// TODO: Replace by correctly populating the files so that the helm render engine templates it
	// correctly
	toTemplate := ""
	for _, f := range strings {
		toTemplate += f
	}

	input := map[string]interface{}{"Values": d}

	if err := mergo.Merge(&input, map[string]interface{}{"Values": values}); err != nil {
		return "", err
	}

	return String(toTemplate, input)
}

// ReadFilesInDir reads a list of paths and reads all yaml file inside. It returns a
// slice of strings with the raw content of the yaml
func FilesInDir(path []string) (res []string, err error) {
	for _, t := range path {
		var rel string
		rel, err = fileHelper.Rel2Abs(t)
		if err != nil {
			return nil, err
		}

		if !fileHelper.Exists(rel) {
			continue
		}
		var files []string
		files, err = fileHelper.ListDir(rel)
		if err != nil {
			return
		}

		for _, f := range files {
			if strings.ToLower(filepath.Ext(f)) == ".yaml" {
				res = append(res, f)
			}
		}
	}
	return
}

// ReadFiles reads all the given files and returns a slice of []*chart.File
// containing the raw content and the file name for each file
func ReadFiles(s ...string) (res []string) {
	for _, c := range s {
		raw, err := ioutil.ReadFile(c)
		if err != nil {
			return
		}
		res = append(res, string(raw))
	}

	return
}

type templatedata map[string]interface{}

// UnMarshalValues unmarshal values files and joins them into a unique templatedata
// the join happens from right to left, so any rightmost value file overwrites the content of the ones before it.
func UnMarshalValues(values []string) (templatedata, error) {
	dst := templatedata{}
	if len(values) > 0 {
		for _, bv := range reverse(values) {
			current := templatedata{}

			defBuild, err := ioutil.ReadFile(bv)
			if err != nil {
				return nil, errors.Wrap(err, "rendering file "+bv)
			}
			err = yaml.Unmarshal(defBuild, &current)
			if err != nil {
				return nil, errors.Wrap(err, "rendering file "+bv)
			}
			if err := mergo.Merge(&dst, current); err != nil {
				return nil, errors.Wrap(err, "merging values file "+bv)
			}
		}
	}
	return dst, nil
}

func reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// RenderWithValues render a group of files with values files
func RenderWithValues(rawFiles []string, valuesFile string, defaultFile ...string) (string, error) {
	if !fileHelper.Exists(valuesFile) {
		return "", errors.New("file does not exist: " + valuesFile)
	}
	val, err := ioutil.ReadFile(valuesFile)
	if err != nil {
		return "", errors.Wrap(err, "reading file: "+valuesFile)
	}

	var values templatedata
	if err = yaml.Unmarshal(val, &values); err != nil {
		return "", errors.Wrap(err, "unmarshalling values")
	}

	dst, err := UnMarshalValues(defaultFile)
	if err != nil {
		return "", errors.Wrap(err, "unmarshalling values")
	}

	return Render(ReadFiles(rawFiles...), values, dst)
}
