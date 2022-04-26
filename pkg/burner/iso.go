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
	"strings"

	"github.com/bhojpur/iso/pkg/schema"
	"github.com/bhojpur/iso/pkg/utils"
	"github.com/pkg/errors"
	"github.com/twpayne/go-vfs"
)

func GenISO(s *schema.SystemSpec, source string, f vfs.FS) error {

	diskImage := s.ISOName()
	label := s.Label
	var bootloader_args string

	// Detect syslinux usage. Probably a new flag for that in schema
	// would be better but for backward compatibility we need some sort
	// of automatic detection.
	if strings.Contains(s.BootFile, "isolinux") {
		bootloader_args = fmt.Sprintf(`\
		  -boot_image isolinux bin_path="%s" \
		  -boot_image isolinux system_area="%s/%s" \
		  -boot_image isolinux partition_table=on \`, s.BootFile, source, s.IsoHybridMBR)
	} else {
		bootloader_args = fmt.Sprintf(`\
		  -boot_image grub bin_path="%s" \
		  -boot_image grub grub2_mbr="%s/%s" \
		  -boot_image grub grub2_boot_info=on`, s.BootFile, source, s.IsoHybridMBR)
	}

	if err := run(fmt.Sprintf(
		`xorriso \
		  -volid "%s" \
		  -joliet on -padding 0 \
		  -outdev "%s" \
		  -map "%s" / -chmod 0755 -- %s \
		  -boot_image any partition_offset=16 \
		  -boot_image any cat_path="%s" \
		  -boot_image any cat_hidden=on \
		  -boot_image any boot_info_table=on \
		  -boot_image any platform_id=0x00 \
		  -boot_image any emul_type=no_emulation \
		  -boot_image any load_size=2048 \
		  -append_partition 2 0xef "%s/boot/uefi.img" \
		  -boot_image any next \
		  -boot_image any efi_path=--interval:appended_partition_2:all:: \
		  -boot_image any platform_id=0xef \
		  -boot_image any emul_type=no_emulation`,
		label, diskImage, source, bootloader_args, s.BootCatalog, source)); err != nil {
		info(err)
		return err
	}

	checksum, err := utils.Checksum(diskImage)
	if err != nil {
		return errors.Wrap(err, "while calculating checksum")
	}

	return f.WriteFile(diskImage+".sha256", []byte(fmt.Sprintf("%s %s", checksum, diskImage)), os.ModePerm)
}
