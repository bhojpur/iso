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

XZ="${XZ:-xz}"


if [[ -z "$INITRAMFS_ROOTFS" ]]; then
    info "Packing initramfs"

    # Remove the old 'initramfs' archive if it exists.
    rm -f $WORKDIR/rootfs.cpio.xz || true

    pushd $ROOTFS_DIR  > /dev/null 2>&1

    # Packs the current 'initramfs' folder structure in 'cpio.xz' archive.
    find . | cpio -R root:root -H newc -o | $XZ -9 --check=none > $WORKDIR/rootfs.cpio.xz

    echo "Packing of initramfs has finished."

    popd  > /dev/null 2>&1
else
    echo "Copying initramfs"
    # Try to find the rootfs file in the overlay or initramfs areas
    if [[ -e "$ROOTFS_DIR/boot/$INITRAMFS_ROOTFS" ]] || [[ -L "$ROOTFS_DIR/boot/$INITRAMFS_ROOTFS" ]]; then
        BOOT_DIR=$ROOTFS_DIR/boot
    elif [[ -e "$OVERLAY_DIR/boot/$INITRAMFS_ROOTFS" ]] || [[ -L "$OVERLAY_DIR/boot/$INITRAMFS_ROOTFS" ]]; then
        BOOT_DIR=$OVERLAY_DIR/boot
    fi

    if [[ -L "$BOOT_DIR/$INITRAMFS_ROOTFS" ]]; then
        bz=$(readlink -f $BOOT_DIR/$INITRAMFS_ROOTFS)
        # Install the kernel file.
        cp $BOOT_DIR/$(basename $bz) \
            $WORKDIR/rootfs.cpio.xz
    else
        cp $BOOT_DIR/$INITRAMFS_ROOTFS \
            $WORKDIR/rootfs.cpio.xz 
    fi
fi

info "Packing overlayfs"
rm -f $WORKDIR/rootfs.squashfs || true
mksquashfs "$OVERLAY_DIR" $WORKDIR/rootfs.squashfs -b 1024k -comp xz -Xbcj x86