#!/usr/bin/env bash
set -euo pipefail

cd $TRAVIS_BUILD_DIR

for os in linux darwin windows; do
  [[ $os == 'windows' ]] && windows=1
  docker run \
    -e GOOS=$os \
    -e GOARCH=amd64 \
    -v $PWD:/workdir \
    -v /tmp/output_packages:/output \
    -it golang:${GIMME_GO_VERSION%.x} \
    bash -c "cd /workdir && make clean && make build_generate_secret && cp /workdir/bin/generate_secret /output/generate_secret-$os-amd64-$APP_VERSION${windows:+.exe}"
done
