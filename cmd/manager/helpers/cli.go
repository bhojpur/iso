package cmd_helpers

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
	"fmt"
	"regexp"
	"strings"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"

	_gentoo "github.com/Sabayon/pkgs-checker/pkg/gentoo"
	"github.com/bhojpur/iso/cmd/manager/util"
)

func CreateRegexArray(rgx []string) ([]*regexp.Regexp, error) {
	ans := make([]*regexp.Regexp, len(rgx))
	if len(rgx) > 0 {
		for idx, reg := range rgx {
			re := regexp.MustCompile(reg)
			if re == nil {
				return nil, errors.New("Invalid regex " + reg + "!")
			}
			ans[idx] = re
		}
	}

	return ans, nil
}

func packageData(p string) (string, string) {
	cat := ""
	name := ""
	if strings.Contains(p, "/") {
		packagedata := strings.Split(p, "/")
		cat = packagedata[0]
		name = packagedata[1]
	} else {
		name = p
	}
	return cat, name
}

func packageHasGentooSelector(v string) bool {
	return (strings.HasPrefix(v, "=") || strings.HasPrefix(v, ">") ||
		strings.HasPrefix(v, "<"))
}

func gentooVersion(gp *_gentoo.GentooPackage) string {

	condition := gp.Condition.String()
	if condition == "=" {
		condition = ""
	}

	pkgVersion := fmt.Sprintf("%s%s%s",
		condition,
		gp.Version,
		gp.VersionSuffix,
	)
	if gp.VersionBuild != "" {
		pkgVersion = fmt.Sprintf("%s%s%s+%s",
			condition,
			gp.Version,
			gp.VersionSuffix,
			gp.VersionBuild,
		)
	}
	return pkgVersion
}

func ParsePackageStr(p string) (*types.Package, error) {

	if packageHasGentooSelector(p) {
		gp, err := _gentoo.ParsePackageStr(p)
		if err != nil {
			return nil, err
		}
		if gp.Version == "" {
			gp.Version = "0"
			gp.Condition = _gentoo.PkgCondGreaterEqual
		}

		return &types.Package{
			Name:     gp.Name,
			Category: gp.Category,
			Version:  gentooVersion(gp),
			Uri:      make([]string, 0),
		}, nil
	}

	ver := ">=0"
	cat := ""
	name := ""

	if strings.Contains(p, "@") {
		packageinfo := strings.Split(p, "@")
		ver = packageinfo[1]
		cat, name = packageData(packageinfo[0])
	} else {
		cat, name = packageData(p)
	}

	return &types.Package{
		Name:     name,
		Category: cat,
		Version:  ver,
		Uri:      make([]string, 0),
	}, nil
}

func CheckErr(err error) {
	if err != nil {
		util.DefaultContext.Fatal(err)
	}
}
