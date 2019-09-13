#!/usr/bin/env bash
set -euo pipefail

# this directory gets created by ci_create_packages.sh when it is shared into the container's /output directory
cd /tmp/output_packages

if [[ $TRAVIS_OS_NAME == "linux" ]]; then
  echo "we're on linux"
  for deb in `ls *.deb`; do
    curl -H "X-JFrog-Art-Api:${ARTIFACTORY_PASSWORD}" -XPUT "https://optimizely.jfrog.io/optimizely/deb-optimizely/pool/$deb;deb.distribution=xenial-optimizely;deb.distribution=bionic-optimizely;deb.component=main;deb.architecture=amd64" -T $deb
  done
  for rpm in `ls *.rpm`; do
    curl -H "X-JFrog-Art-Api:${ARTIFACTORY_PASSWORD}" -XPUT https://optimizely.jfrog.io/optimizely/rpm-optimizely/ -T $rpm
  done
  # push docker images to artifactory
  docker login -u ${ARTIFACTORY_USER} -p ${ARTIFACTORY_PASSWORD} optimizely-docker.jfrog.io
  # if you dont specify the tag, it'll push all image versions
  docker push optimizely-docker.jfrog.io/sidedoor
elif [[ $TRAVIS_OS_NAME == "osx" ]]; then
  echo "we're on osx"
else
  echo "we're lost!"
fi
