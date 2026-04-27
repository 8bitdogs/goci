#!/bin/sh
set -eu

if [ ! -z ${GITHUB_TOKEN:-} ]; then
    git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
else
    echo "GITHUB_TOKEN is not set, using public github.com"
fi

exec "$@"
