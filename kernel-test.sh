#!/bin/bash

# config
FABFILES=~/git/fabfiles
VMS=10.0.0.2,10.0.0.3
TEST=test-setup.sh

cd $FABFILES || exit

echo "Run test setup script on VMs"
fab --prompt-for-sudo-password run.bash -H $VMS -e $TEST -s
