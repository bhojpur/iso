package helpers

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
	"os"
	"os/exec"
	"os/user"
	"syscall"

	"github.com/pkg/errors"
)

// This allows a multi-platform switch in the future
func Exec(cmd string, args []string, env []string) error {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return errors.Wrap(err, "Could not find binary in path: "+cmd)
	}
	return syscall.Exec(path, args, env)
}

func GetHomeDir() (ans string) {
	// os/user doesn't work in from scratch environments
	u, err := user.Current()
	if err == nil {
		ans = u.HomeDir
	} else {
		ans = ""
	}
	if os.Getenv("HOME") != "" {
		ans = os.Getenv("HOME")
	}
	return ans
}
