#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${SCRIPT_ROOT}/hack/lib/init.sh"
go::setup_env
cd "${ROOT_PATH}"

echo "codegen start."
echo "tools install..."

GENERATOR_VERSION=${GENERATOR_VERSION:-$(cat "${ROOT_PATH}"/go.mod | grep "k8s.io/code-generator" | awk '{print $2}')}
GO111MODULE=on go install k8s.io/code-generator/cmd/client-gen@${GENERATOR_VERSION}
GO111MODULE=on go install k8s.io/code-generator/cmd/lister-gen@${GENERATOR_VERSION}
GO111MODULE=on go install k8s.io/code-generator/cmd/informer-gen@${GENERATOR_VERSION}

echo "Generating with client-gen..."
client-gen \
	--go-header-file hack/boilerplate.go.txt \
	--input-base="github.com/IBM-Blockchain/fabric-operator/api" \
	--input="v1beta1" \
	--output-package=github.com/IBM-Blockchain/fabric-operator/pkg/generated/clientset \
	--clientset-name=versioned

echo "Generating with lister-gen..."
lister-gen \
	--go-header-file hack/boilerplate.go.txt \
	--input-dirs=github.com/IBM-Blockchain/fabric-operator/api/v1beta1 \
	--output-package=github.com/IBM-Blockchain/fabric-operator/pkg/generated/listers

echo "Generating with informer-gen..."
informer-gen \
	--go-header-file hack/boilerplate.go.txt \
	--input-dirs=github.com/IBM-Blockchain/fabric-operator/api/v1beta1 \
	--versioned-clientset-package=github.com/IBM-Blockchain/fabric-operator/pkg/generated/clientset/versioned \
	--listers-package=github.com/IBM-Blockchain/fabric-operator/pkg/generated/listers \
	--output-package=github.com/IBM-Blockchain/fabric-operator/pkg/generated/informers

# FIXME: this may be fixed by set right arg in client-gen or informer-gen ?
find pkg/generated/informers -type f -exec sed -i.bak 's/IbpV1beta1/Ibp/g' {} +
find pkg/generated/informers -type f -name '*.bak' -delete

echo "codegen done."
