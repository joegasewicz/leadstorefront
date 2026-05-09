#!/usr/bin/env bash

set -e
source ../../.env

terraform output app_floating_ip
