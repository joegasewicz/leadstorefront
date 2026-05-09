#!/usr/bin/env bash

set -e
source ../../.env

terraform plan \
  -var "do_token=${DO_PAT}" \
  -var "pvt_key=${HOME}/.ssh/leadstorefront" \
  -var "pub_key=${HOME}/.ssh/leadstorefront.pub" \
  -var "my_ip=$(curl -4 ifconfig.me)/32" \
  -var "docker_username=${DOCKER_USERNAME}" \
  -var "docker_password=${DOCKER_PASSWORD}" \
  -var "github_username=${GITHUB_USERNAME:-}" \
  -var "github_password=${GITHUB_PASSWORD:-}"
