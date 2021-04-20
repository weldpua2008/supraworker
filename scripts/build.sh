#!/usr/bin/env bash
cd "$(dirname "$0}")" || exit 1
cd ..
docker run -ti --rm   -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.15.11