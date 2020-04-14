#!/usr/bin/env bash
set -euo pipefail

cd $TRAVIS_BUILD_DIR
docker run -v $PWD:/workdir -v /tmp/output_packages:/output -it golang:${GIMME_GO_VERSION%.x} bash -c "cd /workdir && make build_generate_secret && cp /workdir/bin/generate_secret /output"

# try running generate_secret
/tmp/output_packages/generate_secret
md5sum /tmp/output_packages/generate_secret
