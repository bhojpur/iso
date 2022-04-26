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

info "Copy operating system Kernel file"

rm -rf $KERNEL_INSTALLED || true
mkdir -p $KERNEL_INSTALLED

# Try to find the kernel file in the overlay or initramfs areas
if [[ -e "$ROOTFS_DIR/boot/$INITRAMFS_KERNEL" ]] || [[ -L "$ROOTFS_DIR/boot/$INITRAMFS_KERNEL" ]]; then
  BOOT_DIR=$ROOTFS_DIR/boot
elif [[ -e "$OVERLAY_DIR/boot/$INITRAMFS_KERNEL" ]] || [[ -L "$OVERLAY_DIR/boot/$INITRAMFS_KERNEL" ]]; then
  BOOT_DIR=$OVERLAY_DIR/boot
fi

if [[ -L "$BOOT_DIR/$INITRAMFS_KERNEL" ]]; then
  bz=$(readlink -f $BOOT_DIR/$INITRAMFS_KERNEL)
  # Install the kernel file.
  cp $BOOT_DIR/$(basename $bz) \
    $KERNEL_INSTALLED/kernel
else
  cp $BOOT_DIR/$INITRAMFS_KERNEL \
    $KERNEL_INSTALLED/kernel
fi