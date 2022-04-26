package database

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
	"regexp"

	"github.com/bhojpur/iso/pkg/manager/api/core/types"
	"github.com/pkg/errors"
)

func clone(src, dst types.PackageDatabase) error {
	for _, i := range src.World() {
		_, err := dst.CreatePackage(i)
		if err != nil {
			return errors.Wrap(err, "Failed create package "+i.HumanReadableString())
		}
	}
	return nil
}

func copy(src types.PackageDatabase) (types.PackageDatabase, error) {
	dst := NewInMemoryDatabase(false)

	if err := clone(src, dst); err != nil {
		return dst, errors.Wrap(err, "Failed create temporary in-memory db")
	}

	return dst, nil
}

func findPackageByFile(db types.PackageDatabase, pattern string) (types.Packages, error) {

	var ans []*types.Package

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid regex "+pattern+"!")
	}

PACKAGE:
	for _, pack := range db.World() {
		files, err := db.GetPackageFiles(pack)
		if err == nil {
			for _, f := range files {
				if re.MatchString(f) {
					ans = append(ans, pack)
					continue PACKAGE
				}
			}
		}
	}

	return types.Packages(ans), nil

}
