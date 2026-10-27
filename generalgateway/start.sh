#!/bin/sh
set -eu

exec ./generalgateway -f="$(pwd)/conf/config.yaml"
