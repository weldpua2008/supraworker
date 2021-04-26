#!/usr/bin/env bash
cd "$(dirname "$0")" || true

docker-compose up -d

export PYTHONPATH=$PTHONPATH:./It/
python3 -m unittest discover -p 'test_*.py'