package version

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
	"errors"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	semver "github.com/hashicorp/go-version"
	debversion "github.com/knqyf263/go-deb-version"
)

const (
	selectorGreaterThen        = iota
	selectorLessThen           = iota
	selectorGreaterOrEqualThen = iota
	selectorLessOrEqualThen    = iota
	selectorNotEqual           = iota
)

type packageSelector struct {
	Condition int
	Version   string
}

var selectors = map[string]int{
	">=": selectorGreaterOrEqualThen,
	">":  selectorGreaterThen,
	"<=": selectorLessOrEqualThen,
	"<":  selectorLessThen,
	"!":  selectorNotEqual,
}

func readPackageSelector(selector string) packageSelector {
	selectorType := 0
	v := ""

	k := []string{}
	for kk, _ := range selectors {
		k = append(k, kk)
	}

	sort.Slice(k, func(i, j int) bool {
		return len(k[i]) > len(k[j])
	})
	for _, p := range k {
		if strings.HasPrefix(selector, p) {
			selectorType = selectors[p]
			v = strings.TrimPrefix(selector, p)
			break
		}
	}
	return packageSelector{
		Condition: selectorType,
		Version:   v,
	}
}

func semverCheck(vv string, selector string) (bool, error) {
	c, err := semver.NewConstraint(selector)
	if err != nil {
		// Handle constraint not being parsable.

		return false, err
	}

	v, err := semver.NewVersion(vv)
	if err != nil {
		// Handle version not being parsable.

		return false, err
	}

	// Check if the version meets the constraints.
	return c.Check(v), nil
}

// WrappedVersioner uses different means to return unique result that is understendable by Bhojpur ISO
// It tries different approaches to sort, validate, and sanitize to a common versioning format
// that is understendable by the whole code
type WrappedVersioner struct{}

func DefaultVersioner() Versioner {
	return &WrappedVersioner{}
}

func (w *WrappedVersioner) Validate(version string) error {
	if !debversion.Valid(version) {
		return errors.New("invalid version")
	}
	return nil
}

func (w *WrappedVersioner) ValidateSelector(vv string, selector string) bool {
	if vv == "" {
		return true
	}
	vv = w.Sanitize(vv)
	selector = w.Sanitize(selector)

	sel := readPackageSelector(selector)

	selectorV, err := version.NewVersion(sel.Version)
	if err != nil {
		f, _ := semverCheck(vv, selector)
		return f
	}
	v, err := version.NewVersion(vv)
	if err != nil {
		f, _ := semverCheck(vv, selector)
		return f
	}

	switch sel.Condition {
	case selectorGreaterOrEqualThen:
		return v.GreaterThan(selectorV) || v.Equal(selectorV)
	case selectorLessOrEqualThen:
		return v.LessThan(selectorV) || v.Equal(selectorV)
	case selectorLessThen:
		return v.LessThan(selectorV)
	case selectorGreaterThen:
		return v.GreaterThan(selectorV)
	case selectorNotEqual:
		return !v.Equal(selectorV)
	}

	return false
}

func (w *WrappedVersioner) Sanitize(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "_", "-"))
}

func (w *WrappedVersioner) Sort(toSort []string) []string {
	if len(toSort) == 0 {
		return toSort
	}
	var versionsMap map[string]string = make(map[string]string)
	versionsRaw := []string{}
	result := []string{}
	for _, v := range toSort {
		sanitizedVersion := w.Sanitize(v)
		versionsMap[sanitizedVersion] = v
		versionsRaw = append(versionsRaw, sanitizedVersion)
	}

	vs := make([]debversion.Version, len(versionsRaw))
	for i, r := range versionsRaw {
		v, _ := debversion.NewVersion(r)
		vs[i] = v
	}

	sort.Slice(vs, func(i, j int) bool {
		return vs[i].LessThan(vs[j])
	})

	for _, v := range vs {
		result = append(result, versionsMap[v.String()])
	}
	return result
}
