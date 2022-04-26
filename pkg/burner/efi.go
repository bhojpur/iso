package burner

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

	"github.com/bhojpur/iso/pkg/utils"
	"github.com/pkg/errors"
	"github.com/twpayne/go-vfs"
)

func CreateEFIImage(source, diskImage string, f vfs.FS) error {
	align := int64(4 * 1024 * 1024)
	diskImg, _ := f.RawPath(diskImage)
	diskSize, _ := utils.DirSize(source)

	diskF, err := f.Create(diskImg)
	if err != nil {
		return errors.Wrapf(err, "failed creating image %s", diskImg)
	}

	// Align disk size to the next 4MB slot
	diskSize = diskSize/align*align + align

	err = diskF.Truncate(diskSize)
	if err != nil {
		diskF.Close()
		return errors.Wrapf(err, "failed setting file size to %d bytes", diskSize)
	}
	diskF.Close()

	err = run(fmt.Sprintf("mkfs.fat %s", diskImg))
	if err != nil {
		return errors.Wrapf(err, "failed formatting %s image", diskImg)
	}

	err = run(fmt.Sprintf("mcopy -s -i %s %s/* ::", diskImg, source))
	if err != nil {
		return errors.Wrapf(err, "failed copying '%s' files to image '%s'", source, diskImg)
	}

	return nil
}
