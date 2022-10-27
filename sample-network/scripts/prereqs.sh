#!/bin/bash
#
# Copyright contributors to the Hyperledger Fabric Operator project
#
# SPDX-License-Identifier: Apache-2.0
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at:
#
# 	  http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

ARCH=$(echo "$(uname -s|tr '[:upper:]' '[:lower:]'|sed 's/mingw64_nt.*/windows/')-$(uname -m |sed 's/x86_64/amd64/g')" |sed 's/darwin-arm64/darwin-amd64/g')

# Double check that kind, kubectl, docker, and all required images are present.
function check_prereqs() {

  set +e

  # disable docker check for now
  #${CONTAINER_CLI} version > /dev/null
  #if [[ $? -ne 0 ]]; then
  #  echo "No '${CONTAINER_CLI}' binary available?"
  #  exit 1
  #fi

  if [ "${CLUSTER_RUNTIME}" == "kind" ]; then
    kind version > /dev/null
    if [[ $? -ne 0 ]]; then
      echo "No 'kind' binary available? (https://kind.sigs.k8s.io/docs/user/quick-start/#installation)"
      exit 1
    fi
  fi

  kubectl > /dev/null
  if [[ $? -ne 0 ]]; then
    echo "No 'kubectl' binary available? (https://kubernetes.io/docs/tasks/tools/)"
    exit 1
  fi

  # disable jq check for now
  #jq --version > /dev/null
  #if [[ $? -ne 0 ]]; then
  #  echo "No 'jq' binary available? (https://stedolan.github.io/jq/)"
  #  exit 1
  #fi

  # Use the local fabric binaries if available.  If not, go get them.
  # TODO
  bin/peer version &> /dev/null
  if [[ $? -ne 0 ]]; then
    # echo "Downloading LATEST Fabric binaries and config"
    echo "Unzipping the v$FABRIC_VERSION Fabric binaries and config"
    mkdir -p $TEMP_DIR

    # Unzip the package offline to avoid downloading files
    # https://github.com/hyperledger/fabric/releases/download/v2.4.7/hyperledger-fabric-linux-amd64-2.4.7.tar.gz
    # https://github.com/hyperledger/fabric-ca/releases/download/v1.5.5/hyperledger-fabric-ca-linux-amd64-1.5.5.tar.gz
    tar xf fabric-release-files/$ARCH/hyperledger-fabric-$ARCH-2.4.7.tar.gz -C $TEMP_DIR
    tar xf fabric-release-files/$ARCH/hyperledger-fabric-ca-$ARCH-1.5.5.tar.gz -C $TEMP_DIR
    # The download / installation of binaries will also transfer a core.yaml, which overlaps with a local configuration.
    # Pull the binaries into a temp folder and then move them into the target location.
    # (pushd $TEMP_DIR && curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/bootstrap.sh \
    #   | bash -s -- $FABRIC_VERSION $FABRIC_CA_VERSION -s -d)
    mkdir -p bin && mv $TEMP_DIR/bin/* bin

    # delete config files transferred by the installer
    rm $TEMP_DIR/config/configtx.yaml
    rm $TEMP_DIR/config/core.yaml
    rm $TEMP_DIR/config/orderer.yaml
  fi

  export PATH=bin:$PATH

  # Double-check that the binary transfer was OK
  peer version > /dev/null
  if [[ $? -ne 0 ]]; then
    log "No 'peer' binary available?"
    exit 1
  fi

  set -e
}