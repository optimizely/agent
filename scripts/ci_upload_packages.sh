#!/usr/bin/env bash

cd $TRAVIS_BUILD_DIR/ci

if [[ $TRAVIS_OS_NAME == "linux" ]]; then
  echo "we're on linux"
  for deb in `ls *.deb`; do
    curl -H "X-JFrog-Art-Api:${ARTIFACTORY_PASSWORD}" -XPUT "https://optimizely.jfrog.io/optimizely/deb-optimizely/pool/$deb;deb.distribution=xenial;deb.component=main;deb.architecture=amd64" -T $deb
  done
  for rpm in `ls *.rpm`; do
    curl -H "X-JFrog-Art-Api:${ARTIFACTORY_PASSWORD}" -XPUT https://optimizely.jfrog.io/optimizely/rpm-optimizely/ -T $rpm
  done
elif [[ $TRAVIS_OS_NAME == "osx" ]]; then
  echo "we're on osx"
else
  echo "we're lost!"
fi
