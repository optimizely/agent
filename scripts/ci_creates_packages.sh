#!/usr/bin/env bash

if [[ $TRAVIS_OS_NAME == "linux" ]]; then
  echo "we're on linux"
  cd $TRAVIS_BUILD_DIR
  make ci_build_fpm_centos
  make ci_get_fpm_centos
  ls -al *.rpm
  make ci_build_fpm_ubuntu
  make ci_get_fpm_ubuntu
  ls -al *.deb
elif [[ $TRAVIS_OS_NAME == "osx" ]]; then
  echo "we're on osx"
else
  echo "we're lost!"
fi
