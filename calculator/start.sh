#!/bin/sh
set -eu

exec ./calculator -f="$(pwd)/conf/config.yaml"
