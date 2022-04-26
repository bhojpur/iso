#!/bin/bash

# Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

set -e

DIR="${BASH_SOURCE%/*}"
if [[ ! -d "$DIR" ]]; then DIR="$PWD"; fi
. "$DIR/func.sh"

LOOP_DEVICE_HDD=""

init() {
  # Remove the old ISO generation area if it exists.
  echo "Removing old ISO image work area. This may take a while."
  rm -rf $ISOIMAGE

  echo "Preparing new ISO image work area."
  mkdir -p $ISOIMAGE
}

## XXX: Move it away. Else, we ship kernels in a different package, instead of finding symlinks
#  or either we ask it in the spec explictly.
prepare_mll_bios() {
  # This is the folder where we keep legacy BIOS boot artifacts.
  mkdir -p $ISOIMAGE/boot

  # Now we copy the kernel.
  cp $KERNEL_INSTALLED/kernel \
    $ISOIMAGE/boot/kernel.xz

  # Now we copy the root file system.
  cp $WORKDIR/rootfs.cpio.xz \
    $ISOIMAGE/boot/rootfs.xz
}

prepare_overlay() {
  # Now we copy the overlay content if it exists.
  if [ -e $WORKDIR/rootfs.squashfs ] ; then

    echo "The ISO image will have overlay structure."
    cp -r $WORKDIR/rootfs.squashfs $ISOIMAGE
  else
    echo "The ISO image will have no overlay structure."
  fi
}

prepare_boot_bios() {
  # Add the Syslinux configuration files for legacy BIOS and additional
  # UEFI startup script.
  #
  # The existing UEFI startup script does not guarantee that you can run
  # MLL on UEFI systems. This script is invoked only in case your system
  # drops you in UEFI shell with support level 1 or above. See UEFI shell
  # specification 2.2, section 3.1. Depending on your system configuration
  # you may not end up with UEFI shell even if your system supports it.
  # In this case MLL will not boot and you will end up with some kind of
  # UEFI error message.

  bhojpur_install $ISOIMAGE "$ISOIMAGE_PACKAGES" "${BHOJPUR_REPOS}"
}

# Genrate 'El Torito' boot image as per UEFI sepcification 2.7,
# sections 13.3.1.x and 13.3.2.x.
prepare_boot_uefi() {
  # Find the build architecture based on the Busybox executable.
  BUSYBOX_ARCH=$(file $ROOTFS/bin/busybox | cut -d' ' -f3)

  # Determine the proper UEFI configuration. The default image file
  # names are described in UEFI specification 2.7, section 3.5.1.1.
  # Note that the x86_64 UEFI image file name indeed contains small
  # letter 'x'.
  rm -rf $WORKDIR/uefitmp
  mkdir -p $WORKDIR/uefitmp

  bhojpur_install $WORKDIR/uefitmp "$UEFI_PACKAGES" "${BHOJPUR_REPOS}"

  # Find the kernel size in bytes.
  kernel_size=`du -b $KERNEL_INSTALLED/kernel | awk '{print \$1}'`

  # Find the initramfs size in bytes.
  rootfs_size=`du -b $WORKDIR/rootfs.cpio.xz | awk '{print \$1}'`

  loader_size=`du -bs $WORKDIR/uefitmp | awk '{print \$1}'`

  # The EFI boot image is 64KB bigger than the kernel size.
  image_size=$((kernel_size + rootfs_size*2 + loader_size + 65536)) ## XXX: rootfsize is doubled

  echo "Creating UEFI boot image file '$WORKDIR/uefi.img'."
  rm -f $WORKDIR/uefi.img
  truncate -s $image_size $WORKDIR/uefi.img

  echo "Attaching hard disk image file to loop device."
  LOOP_DEVICE_HDD=$(losetup -f)
  losetup $LOOP_DEVICE_HDD $WORKDIR/uefi.img

  echo "Formatting hard disk image with FAT filesystem."
  mkfs.vfat $LOOP_DEVICE_HDD

  echo "Preparing 'uefi' work area."
  rm -rf $WORKDIR/uefi
  mkdir -p $WORKDIR/uefi
  mount $WORKDIR/uefi.img $WORKDIR/uefi

  cp -rf  $WORKDIR/uefitmp/* $WORKDIR/uefi

  echo "Preparing kernel and rootfs."
  mkdir -p $WORKDIR/uefi/minimal/$ARCH
  cp $KERNEL_INSTALLED/kernel \
    $WORKDIR/uefi/minimal/$ARCH/kernel.xz
  cp $WORKDIR/rootfs.cpio.xz \
    $WORKDIR/uefi/minimal/$ARCH/rootfs.xz

  echo "Unmounting UEFI boot image file."
  sync
  umount $WORKDIR/uefi
  sync
  sleep 1

  # The directory is now empty (mount point for loop device).
  rm -rf $WORKDIR/uefi $WORKDIR/uefitmp

  # Make sure the UEFI boot image is readable.
  chmod ugo+r $WORKDIR/uefi.img

  mkdir -p $ISOIMAGE/boot
  cp $WORKDIR/uefi.img \
    $ISOIMAGE/boot
}

check_root() {
  if [ ! "$(id -u)" = "0" ] ; then
    cat << CEOF

  ISO image preparation process for UEFI systems requires root permissions
  but you don't have such permissions. Restart this script with root
  permissions in order to generate UEFI compatible ISO structure.

CEOF
    exit 1
  fi
}

cleanup_on_exit () {
  [ -z "${LOOP_DEVICE_HDD}" ] || {

    losetup -d ${LOOP_DEVICE_HDD} || {

      echo "ATTENTION: Something goes wrong on detach loopback device ${LOOP_DEVICE_HDD}."
      echo "Ignoring it for now."

    }

  }

  return 0
}

trap "cleanup_on_exit" EXIT INT TERM

echo "*** PREPARE ISO BEGIN ***"

check_root
init
prepare_boot_uefi
prepare_boot_bios
prepare_mll_bios
prepare_overlay

echo "*** PREPARE ISO END ***"