#!/usr/bin/env bash
set -o pipefail

# Creates a draft GitHub release for the given tag, using the latest
# CHANGELOG.md entry as the release description.
#
# This is a `gh`-based replacement for travisci-tools/release_github/release_github_v2.sh,
# which relied on the deprecated `hub` CLI. `gh` is preinstalled on GitHub-hosted runners.
#
# Run this in the same directory where CHANGELOG.md lives.
# grep -P below is GNU grep specific, so this is meant to run on Linux.

CHANGELOG="CHANGELOG.md"

# the tag being released, e.g. v4.6.0
GIT_TAG="$1"

# set to true for debugging
debug=false

if [ $# -ne 1 ]; then
  echo "example: $0 v3.1.1"
  exit 1
fi

# VERSION is the git tag with the "v" prefix removed
VERSION=${GIT_TAG#v}

# The release is often already created (the tag push that triggers CI usually
# comes from a GitHub release created in the UI). In that case there is nothing
# to create -- just let the caller move on to uploading assets.
if gh release view "${GIT_TAG}" >/dev/null 2>&1; then
  echo "Release ${GIT_TAG} already exists; skipping creation."
  exit 0
fi

# check that CHANGELOG.md has been updated
# (first version entry in the changelog should match our VERSION argument)
NEW_VERSION_CHECK=$(grep -P '^## \[\d+\.\d+\.\d+\]' ${CHANGELOG} | cut -d[ -f2 | cut -d] -f1 | awk 'NR==1')

$debug && echo "NEW_VERSION_CHECK $NEW_VERSION_CHECK"
$debug && echo "VERSION $VERSION"

if [[ "${NEW_VERSION_CHECK}" != "${VERSION}" ]]; then
  echo "ERROR: ${CHANGELOG} has not been updated yet."
  exit 1
fi

NEW_VERSION=$(grep -P '^## \[\d+\.\d+\.\d+\]' ${CHANGELOG} | awk 'NR==1' | sed -e 's/\[/\\\[/' | sed -e 's/\]/\\\]/')
LAST_VERSION=$(grep -P '^## \[\d+\.\d+\.\d+\]' ${CHANGELOG} | awk 'NR==2' | sed -e 's/\[/\\\[/' | sed -e 's/\]/\\\]/')

DESCRIPTION=$(awk "/^${NEW_VERSION}$/,/^${LAST_VERSION:-nothingmatched}$/" "${CHANGELOG}" | grep -v "^${LAST_VERSION:-nothingmatched}$")

$debug && echo "NEW_VERSION $NEW_VERSION"
$debug && echo "LAST_VERSION $LAST_VERSION"
$debug && echo "DESCRIPTION $DESCRIPTION"

# --draft creates a draft release; the tag already exists (release is triggered by the tag push)
gh release create "${GIT_TAG}" --draft --title "Release ${VERSION}" --notes "${DESCRIPTION}"
