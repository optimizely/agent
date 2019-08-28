#!/usr/bin/env bash

if [[ $TRAVIS_OS_NAME == "linux" ]]; then
  echo "we're on linux"
elif [[ $TRAVIS_OS_NAME == "osx" ]]; then
  echo "we're on osx"
else
  echo "we're lost!"
fi
