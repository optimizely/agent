#!/usr/bin/env bash

set -e
BRANCH_NAME=${1:-master}
mkdir $HOME/travisci-tools && pushd $HOME/travisci-tools && git init && git pull https://$CI_USER_TOKEN@github.com/optimizely/travisci-tools.git $BRANCH_NAME  && popd

