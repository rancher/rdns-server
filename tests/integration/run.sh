#!/bin/bash
set -ex

flake8 .
py.test -v . $@