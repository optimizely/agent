#!/usr/bin/env bash
set -euo pipefail

if [[ $TRAVIS_OS_NAME == "linux" ]]; then
  echo "we're on linux"

  # push docker images to dockerhub
  echo "$DOCKERHUB_PASS" | docker login -u "$DOCKERHUB_USER" --password-stdin
  # if you dont specify the tag, it'll push all image versions
  docker push optimizely/agent

elif [[ $TRAVIS_OS_NAME == "osx" ]]; then
  echo "we're on osx"
else
  echo "we're lost!"
fi
