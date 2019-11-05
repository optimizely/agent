#!/usr/bin/env bash
set -euo pipefail

if [[ $TRAVIS_OS_NAME == "linux" ]]; then
  echo "we're on linux"
  cd $TRAVIS_BUILD_DIR
  make -e ci_build_fpm_centos
  make -e ci_get_fpm_centos
  make -e ci_build_fpm_ubuntu
  make -e ci_get_fpm_ubuntu
  make -e ci_build_dockerimage
  make -e ci_build_dockerimage_alpine
elif [[ $TRAVIS_OS_NAME == "osx" ]]; then
  echo "we're on osx"
  mkdir /tmp/output_packages # make osx happy
else
  echo "we're lost!"
fi
