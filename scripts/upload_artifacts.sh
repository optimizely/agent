#!/usr/bin/env bash
set -e

choco install awscli
export PATH=$PATH:'/c/Program Files/Amazon/AWSCLI/bin'
aws s3 cp "bin\optimizely.exe" "s3://${AWS_BUCKET}/${TRAVIS_REPO_SLUG}/${TRAVIS_BUILD_NUMBER}/${TRAVIS_JOB_NUMBER}/optimizely.exe" --quiet
