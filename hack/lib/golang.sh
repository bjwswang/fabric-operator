#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

function go::create_gopath_tree() {
	local go_pkg_dir="${BUILD_GOPATH}/src/${PACKAGE_NAME}"
	local go_pkg_basedir
	go_pkg_basedir=$(dirname "${go_pkg_dir}")

	mkdir -p "${go_pkg_basedir}"

	if [[ ! -e ${go_pkg_dir} || "$(readlink "${go_pkg_dir}")" != "${ROOT_PATH}" ]]; then
		ln -snf "${ROOT_PATH}" "${go_pkg_dir}"
	fi
}

function go::setup_env() {
	go::create_gopath_tree

	export GOPATH="${BUILD_GOPATH}"
	export GOCACHE="${BUILD_GOPATH}/cache"

	export PATH="${BUILD_GOPATH}/bin:${PATH}"

	GOROOT=$(go env GOROOT)
	export GOROOT

	unset GOBIN

	cd "$BUILD_GOPATH/src/$PACKAGE_NAME"
}
