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
    -i golang:${GIMME_GO_VERSION%.x} \
    bash -c "cd /workdir && make clean && make build_generate_secret \
    && cp /workdir/bin/generate_secret /output/generate_secret${windows:+.exe} \
    && tar cvfz /output/generate_secret-$os-amd64-$APP_VERSION.tar.gz -C /output generate_secret${windows:+.exe} \
    && rm /output/generate_secret${windows:+.exe}"
  ls -al /tmp/output_packages/generate_secret-$os-amd64-$APP_VERSION.tar.gz
  md5sum /tmp/output_packages/generate_secret-$os-amd64-$APP_VERSION.tar.gz
done
