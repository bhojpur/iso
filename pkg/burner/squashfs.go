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
	"os"

	"github.com/bhojpur/iso/pkg/schema"
	"github.com/bhojpur/iso/pkg/utils"
	"github.com/pkg/errors"
	"github.com/twpayne/go-vfs"

	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/fat32"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	"github.com/diskfs/go-diskfs/filesystem/squashfs"
)

func CreateFilesystem(d *disk.Disk, spec disk.FilesystemSpec) (filesystem.FileSystem, error) {
	// find out where the partition starts and ends, or if it is the entire disk
	var (
		size, start int64
	)
	switch {
	case spec.Partition == 0:
		size = d.Size
		start = 0
	case d.Table == nil:
		return nil, fmt.Errorf("cannot create filesystem on a partition without a partition table")
	default:
		partitions := d.Table.GetPartitions()
		// API indexes from 1, but slice from 0
		partition := spec.Partition - 1
		if spec.Partition > len(partitions) {
			return nil, fmt.Errorf("cannot create filesystem on partition %d greater than maximum partition %d", spec.Partition, len(partitions))
		}
		size = partitions[partition].GetSize()
		start = partitions[partition].GetStart()
	}

	switch spec.FSType {
	case filesystem.TypeFat32:
		return fat32.Create(d.File, size, start, d.LogicalBlocksize, spec.VolumeLabel)
	case filesystem.TypeISO9660:
		return iso9660.Create(d.File, size, start, d.LogicalBlocksize, spec.WorkDir)
	case filesystem.TypeSquashfs:
		return squashfs.Create(d.File, size, start, d.LogicalBlocksize)
	default:
		return nil, errors.New("Unknown filesystem type requested")
	}
}

// XXX: This doesn't work still
func nativeSquashfs(diskImage, source string, options schema.SquashfsOptions, f vfs.FS) error {

	diskImg, err := f.RawPath(diskImage)

	if diskImg == "" {
		return errors.New("must have a valid path for diskImg")
	}
	size, err := utils.DirSize(source)
	var diskSize int64 = size // 10 MB
	mydisk, err := diskfs.Create(diskImg, diskSize, diskfs.Raw)
	if err != nil {
		return errors.Wrapf(err, "while creating squashfs disk")
	}

	mydisk.LogicalBlocksize = 4096
	fspec := disk.FilesystemSpec{Partition: 0, FSType: filesystem.TypeSquashfs, VolumeLabel: options.Label}
	fs, err := CreateFilesystem(mydisk, fspec)
	if err != nil {
		return errors.Wrapf(err, "while creating squashfs size: %d", size)
	}

	if err := copyToFS(source, fs, f); err != nil {
		return errors.Wrapf(err, "while copying files")
	}

	sqs, ok := fs.(*squashfs.FileSystem)
	if !ok {
		return errors.Wrapf(err, "not a squashfs")
	}

	return sqs.Finalize(squashfs.FinalizeOptions{})
}

func CreateSquashfs(diskImage string, source string, options schema.SquashfsOptions, f vfs.FS) error {
	if os.Getenv(("NATIVE")) == "true" {
		return nativeSquashfs(diskImage, source, options, f)
	}
	cmd := fmt.Sprintf("mksquashfs %s %s -b 1024k -comp %s", source, diskImage, options.Compression)
	if options.CompressionOptions != "" {
		cmd = fmt.Sprintf("%s %s", cmd, options.CompressionOptions)
	}
	return run(cmd)
}
