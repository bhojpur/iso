package gc

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
)

type GarbageCollector string

func (c GarbageCollector) String() string {
	return string(c)
}

func (c GarbageCollector) init() error {
	if _, err := os.Stat(string(c)); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(string(c), os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c GarbageCollector) Clean() error {
	return os.RemoveAll(string(c))
}

func (c GarbageCollector) TempDir(pattern string) (string, error) {
	err := c.init()
	if err != nil {
		return "", err
	}
	return ioutil.TempDir(string(c), pattern)
}

func (c GarbageCollector) TempFile(s string) (*os.File, error) {
	err := c.init()
	if err != nil {
		return nil, err
	}
	return ioutil.TempFile(string(c), s)
}
