#!/bin/sh
set -eu

exec ./greeter -f="$(pwd)/conf/config.yaml"
