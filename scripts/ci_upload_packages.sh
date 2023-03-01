#!/usr/bin/env bash
set -euo pipefail

if [[ $TRAVIS_OS_NAME == "linux" ]]; then
  echo "we're on linux"

  # push docker images to dockerhub
  echo "$DOCKERHUB_PASS" | docker login -u "$DOCKERHUB_USER" --password-stdin
  # if you dont specify the tag, it'll push all image versions --> No longer the default
  # https://docs.docker.com/engine/release-notes/#20100
  # Add -a/--all-tags to docker push docker/cli#2220
  #docker push -a optimizely/agent

elif [[ $TRAVIS_OS_NAME == "osx" ]]; then
  echo "we're on osx"
else
  echo "we're lost!"
fi
