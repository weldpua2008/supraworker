#!/usr/bin/env bash
#
# git clone https://github.com/weldpua2008/supraworker.git
# cd supraworker
#
cd "$(dirname "$0}")" || exit 1
cd ..
docker run -ti --rm   -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang:1.15.11