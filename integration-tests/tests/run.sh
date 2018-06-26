#!/bin/bash

sed -i "s/%ip%/${ENV_HOST}/" common.py
set -x

flake8 .
py.test -v . $@
