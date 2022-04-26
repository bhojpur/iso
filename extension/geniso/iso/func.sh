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

# Output functions

if ((BASH_VERSINFO[0] >= 4)) && [[ $'\u2388 ' != "\\u2388 " ]]; then
        ROCKET_IMG=$'\U1F680 '
        RECIPE_IMG=$'\U1F382 '
        ARROW_IMG=$'\U27A4 '
        INFO_IMG=$'\U2139 '
        WARN_IMG=$'\U26A0 '
        ERR_IMG=$'\U1F480 '
        OK_IMG=$'\U2705 '
    else
        ROCKET_IMG=$'\xF0\x9F\x9A\x80 '
        RECIPE_IMG=$'\xF0\x9F\x8E\x82 '
        ARROW_IMG=$'\xE2\x9E\xA4 '
        INFO_IMG=$'\xE2\x84\xB9 '
        WARN_IMG=$'\xE2\x9A\xA0 '
        ERR_IMG=$'\xF0\x9F\x92\x80 '
        OK_IMG=$'\xE2\x9C\x85 '
fi

{
# Reset
Color_Off='\033[0m'       # Text Reset
# Regular Colors
Black='\033[0;30m'        # Black
Red='\033[0;31m'          # Red
Green='\033[0;32m'        # Green
Yellow='\033[0;33m'       # Yellow
Blue='\033[0;34m'         # Blue
Purple='\033[0;35m'       # Purple
Cyan='\033[0;36m'         # Cyan
White='\033[0;37m'        # White

# Bold
BBlack='\033[1;30m'       # Black
BRed='\033[1;31m'         # Red
BGreen='\033[1;32m'       # Green
BYellow='\033[1;33m'      # Yellow
BBlue='\033[1;34m'        # Blue
BPurple='\033[1;35m'      # Purple
BCyan='\033[1;36m'        # Cyan
BWhite='\033[1;37m'       # White

# Underline
UBlack='\033[4;30m'       # Black
URed='\033[4;31m'         # Red
UGreen='\033[4;32m'       # Green
UYellow='\033[4;33m'      # Yellow
UBlue='\033[4;34m'        # Blue
UPurple='\033[4;35m'      # Purple
UCyan='\033[4;36m'        # Cyan
UWhite='\033[4;37m'       # White

# Background
On_Black='\033[40m'       # Black
On_Red='\033[41m'         # Red
On_Green='\033[42m'       # Green
On_Yellow='\033[43m'      # Yellow
On_Blue='\033[44m'        # Blue
On_Purple='\033[45m'      # Purple
On_Cyan='\033[46m'        # Cyan
On_White='\033[47m'       # White

# High Intensity
IBlack='\033[0;90m'       # Black
IRed='\033[0;91m'         # Red
IGreen='\033[0;92m'       # Green
IYellow='\033[0;93m'      # Yellow
IBlue='\033[0;94m'        # Blue
IPurple='\033[0;95m'      # Purple
ICyan='\033[0;96m'        # Cyan
IWhite='\033[0;97m'       # White

# Bold High Intensity
BIBlack='\033[1;90m'      # Black
BIRed='\033[1;91m'        # Red
BIGreen='\033[1;92m'      # Green
BIYellow='\033[1;93m'     # Yellow
BIBlue='\033[1;94m'       # Blue
BIPurple='\033[1;95m'     # Purple
BICyan='\033[1;96m'       # Cyan
BIWhite='\033[1;97m'      # White

# High Intensity backgrounds
On_IBlack='\033[0;100m'   # Black
On_IRed='\033[0;101m'     # Red
On_IGreen='\033[0;102m'   # Green
On_IYellow='\033[0;103m'  # Yellow
On_IBlue='\033[0;104m'    # Blue
On_IPurple='\033[0;105m'  # Purple
On_ICyan='\033[0;106m'    # Cyan
On_IWhite='\033[0;107m'   # White
}

function info {
    local message="$*"

    printf "${BBlue}${INFO_IMG}${BWhite}${ROCKET_IMG}${RECIPE_IMG}${BBlue} ${ARROW_IMG} ${BWhite}${On_Black}$message$Color_Off\n"
}


function err {
    local message="$*"

    printf "${BRed}${ERR_IMG}${BRed}${ARROW_IMG} ${BWhite}${On_Black}$message$Color_Off\n"
}

function ok {
    local message="$*"

    printf "${BGreen}${OK_IMG}${BGreen}${ARROW_IMG} ${BWhite}${On_Black}$message$Color_Off\n"
}

umount_rootfs() {
  local rootfs=$1
  #${SUDO} umount -l $rootfs/boot

  ${SUDO} umount -l $rootfs/dev/pts > /dev/null 2>&1 || true
  ${SUDO} umount -l $rootfs/dev/ > /dev/null 2>&1 || true
  ${SUDO} umount -l $rootfs/sys/ > /dev/null 2>&1 || true
  ${SUDO} umount -l $rootfs/proc/  > /dev/null 2>&1 || true
}

# Automatically Bhojpur ISO using box feature (pivot container)
# on run finalizer if the target dir is != /.
# In this case is not needed chroot.
bhojpur_box_install() {
  local rootfs=$1
  local packages="$2"
  local repositories="$3"
  local keep_db="$4"

  export BHOJPUR_ISO_NOLOCK=true

  # Create a valid FS structure in order to boot
  # we shouldn't really care to do this here, but let packages instead create those on need.
  # we do this here just for safety (who on earth would create a non-bootable ISO?)
  for d in "/dev" "/sys" "/proc" "/tmp" "/dev/pts" "/run" "/var/lock" "/isodb" "/etc"; do
    mkdir -p ${rootfs}${d} || true
  done

  cp -rf "${BHOJPUR_CONFIG}" "$rootfs/iso.yaml"

  # XXX: This is temporarly needed until we fix override from CLI of --system-target
  #      and the --system-dbpath options
  cat <<EOF >> "$rootfs/iso.yaml"
system:
  rootfs: "$rootfs"
  database_engine: "boltdb"
repos_confdir:
  - $rootfs/etc/bhojpur/repos.conf.d
repositories:
- name: "repository-index"
  description: "Bhojpur OS Repository Index"
  type: "http"
  enable: true
  cached: true
  priority: 1
  urls:
  - "https://get.bhojpur.net/iso/repository-index"

EOF

  # Required to connect to remote repositories
  if [ ! -f "$rootfs/etc/resolv.conf" ]; then
    echo "nameserver 8.8.8.8" > $rootfs/etc/resolv.conf
  fi
  if [ ! -f "$rootfs/etc/ssl/certs/ca-certificates.crt" ]; then
    mkdir -p $rootfs/etc/ssl/certs
    cp -rfv "${CA_CERTIFICATES}" $rootfs/etc/ssl/certs
  fi

  cp -rfv ${BHOJPUR_BIN} $rootfs/isomgr

  if [ -n "${repositories}" ]; then
    echo "Installing repositories ${repositories} in $rootfs, logs available at ${BHOJPUR_GENISO_OUTPUT}"
    ${SUDO} isomgr install --config $rootfs/iso.yaml ${BHOJPUR_OPTS} ${repositories} >> ${BHOJPUR_GENISO_OUTPUT} 2>&1
  fi

  echo "Installing packages ${packages} in $rootfs, logs available at ${BHOJPUR_GENISO_OUTPUT}"
  ${SUDO} isomgr install --config $rootfs/iso.yaml ${BHOJPUR_OPTS} ${packages} >> ${BHOJPUR_GENISO_OUTPUT} 2>&1
  ${SUDO} isomgr cleanup --config $rootfs/iso.yaml

  if [[ "$keep_db" != "true" ]]; then
    rm -rf "$rootfs/var/bhojpur/"
    rm -rf "$rootfs/etc/bhojpur/repos.conf.d/"
  fi

  mv "$rootfs/iso.yaml" "$rootfs/etc/bhojpur/"
}


bhojpur_install() {
  local rootfs=$1
  local packages="$2"
  local repositories="$3"
  local keep_db="$4"

  export BHOJPUR_ISO_NOLOCK=true

 
  # Create a valid FS structure in order to boot
  # we shouldn't really care to do this here, but let packages instead create those on need.
  # we do this here just for safety (who on earth would create a non-bootable ISO?)
  for d in "/dev" "/sys" "/proc" "/tmp" "/dev/pts" "/run" "/var/lock" "/isodb" "/etc"; do
    mkdir -p ${rootfs}${d} || true
  done

  cp -rf "${BHOJPUR_CONFIG}" "$rootfs/iso.yaml"

  # XXX: This is temporarly needed until we fix override from Bhojpur CLI of --system-target
  #      and the --system-dbpath options
  cat <<EOF >> "$rootfs/iso.yaml"
system:
  rootfs: "/"
  database_engine: "boltdb"
repos_confdir:
  - /etc/bhojpur/repos.conf.d
repositories:
- name: "repository-index"
  description: "Bhojpur OS Repository Index"
  type: "http"
  enable: true
  cached: true
  priority: 1
  urls:
  - "https://get.bhojpur.net/iso/repository-index"

EOF

  ${SUDO} mount --bind /dev $rootfs/dev/
  ${SUDO} mount --bind /sys $rootfs/sys/
  ${SUDO} mount --bind /proc $rootfs/proc/
  ${SUDO} mount --bind /dev/pts $rootfs/dev/pts

  pushd ${rootfs}

  # Required to connect to remote repositories
  if [ ! -f "etc/resolv.conf" ]; then
    echo "nameserver 8.8.8.8" > etc/resolv.conf
  fi
  if [ ! -f "etc/ssl/certs/ca-certificates.crt" ]; then
    mkdir -p etc/ssl/certs
    cp -rfv "${CA_CERTIFICATES}" etc/ssl/certs
  fi

  cp -rfv ${BHOJPUR_BIN} $rootfs/isomgr


  if [ -n "${repositories}" ]; then
    echo "Installing repositories ${repositories} in $rootfs, logs available at ${BHOJPUR_GENISO_OUTPUT}"
    ${SUDO} chroot . /isomgr install --config /iso.yaml ${BHOJPUR_OPTS} ${repositories} >> ${BHOJPUR_GENISO_OUTPUT} 2>&1
  fi

  echo "Installing packages ${packages} in $rootfs, logs available at ${BHOJPUR_GENISO_OUTPUT}"
  ${SUDO} chroot . /isomgr install --config /iso.yaml ${BHOJPUR_OPTS} ${packages} >> ${BHOJPUR_GENISO_OUTPUT} 2>&1
  ${SUDO} chroot . /isomgr cleanup --config /iso.yaml

  if [[ "$keep_db" != "true" ]]; then
    rm -rf "$rootfs/var/bhojpur/"
    rm -rf "$rootfs/etc/bhojpur/repos.conf.d/"
  fi

  # Cleanup/umount
  umount_rootfs $rootfs

  mv "$rootfs/iso.yaml" "$rootfs/etc/bhojpur/"

  rm $rootfs/isomgr

  popd
}

run_hook() {
  local rootfs=$1
  local script=$2

  if [ ! -f "${script}" ] ; then
    err "ERROR: Hook script ${script} not found!"
    exit 1
  fi

  # Create a valid FS structure in order to boot
  # we shouldn't really care to do this here, but let packages instead create those on need.
  # we do this here just for safety (who on earth would create a non-bootable ISO?)
  for d in "/dev" "/sys" "/proc" "/tmp" "/dev/pt" "/run" "/var/lock" "/isodb" "/etc"; do
    mkdir -p ${rootfs}${d} || true
  done

  ${SUDO} mount --bind /dev $rootfs/dev/
  ${SUDO} mount --bind /sys $rootfs/sys/
  ${SUDO} mount --bind /proc $rootfs/proc/
  ${SUDO} mount --bind /dev/pts $rootfs/dev/pts

  pushd ${rootfs}

  # Required to connect to remote repositories
  if [ ! -f "etc/resolv.conf" ]; then
    echo "nameserver 8.8.8.8" > etc/resolv.conf
  fi
  if [ ! -f "etc/ssl/certs/ca-certificates.crt" ]; then
    mkdir -p etc/ssl/certs
    cp -rfv "${CA_CERTIFICATES}" etc/ssl/certs
  fi

  cp -vf ${script} "$rootfs/hook.sh"
  chmod a+x "${rootfs}/hook.sh"

  echo "Run hook ${script} in $rootfs, logs available at ${BHOJPUR_GENISO_OUTPUT}"
  ${SUDO} chroot . /bin/sh -c /hook.sh >> ${BHOJPUR_GENISO_OUTPUT} 2>&1

  rm -v $rootfs/hook.sh

  # Cleanup/umount
  umount_rootfs $rootfs

  popd
}
