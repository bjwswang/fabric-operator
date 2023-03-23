#!/usr/bin/env bash

unset CDPATH

ROOT_PATH=$(git rev-parse --show-toplevel)
readonly PACKAGE_NAME="github.com/IBM-Blockchain/fabric-operator"
readonly OUTPUT_DIR="${ROOT_PATH}/_output"
readonly BUILD_GOPATH="${OUTPUT_DIR}/go"
source "${ROOT_PATH}/hack/lib/golang.sh"
