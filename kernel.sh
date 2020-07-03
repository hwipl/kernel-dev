#!/bin/bash

# config
FABFILES=~/git/fabfiles
BUILDDIR=~/git/archlinux-kernel-build/net-next
VMS=10.0.0.2,10.0.0.3

cd $FABFILES || exit

echo "Making kernel package"
fab archlinux.makepkg -a "-f" -p $BUILDDIR

echo "Deploying kernel package to VMs and rebooting them"
fab --prompt-for-sudo-password \
	-H $VMS archlinux.install-package-file \
	-p $BUILDDIR/linux-custom-git-1-x86_64.pkg.tar.xz \
	shutdown.reboot
