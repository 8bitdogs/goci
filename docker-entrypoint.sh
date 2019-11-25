#!/bin/sh
set -eu

git config --global url."https://${GIT_USERNAME}:${GIT_PASSWORD}@github".insteadOf https://github
