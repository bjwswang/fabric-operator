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
#set -x
export TERM=xterm-color

KindName="kind"
TimeoutSeconds=${TimeoutSeconds:-"600"}
HelmTimeout=${HelmTimeout:-"1800s"}
KindVersion=${KindVersion:-"v1.24.4"}
TempFilePath=${TempFilePath:-"/tmp/fabric-operator-example-test"}
KindConfigPath=${TempFilePath}/kind-config.yaml
InstallDirPath=${TempFilePath}/installer
DefaultPassWord=${DefaultPassWord:-'passw0rd'}
LOG_DIR=${LOG_DIR:-"/tmp/fabric-operator-example-test/logs"}
ComponentImageFile=${ComponentImageFile:-"/tmp/all.image.tar"}

Timeout="${TimeoutSeconds}s"
mkdir ${TempFilePath} || true

function debugInfo {
	if [[ $? -eq 0 ]]; then
		exit 0
	fi
	if [[ $debug -ne 0 ]]; then
		exit 1
	fi

	warning "debugInfo start 🧐"
	mkdir -p $LOG_DIR

	warning "1. Try to get all resources "
	kubectl api-resources --verbs=list -o name | xargs -n 1 kubectl get -A --ignore-not-found=true --show-kind=true >$LOG_DIR/get-all-resources-list.log
	kubectl api-resources --verbs=list -o name | xargs -n 1 kubectl get -A -oyaml --ignore-not-found=true --show-kind=true >$LOG_DIR/get-all-resources-yaml.log

	warning "2. Try to describe all resources "
	kubectl api-resources --verbs=list -o name | xargs -n 1 kubectl describe -A >$LOG_DIR/describe-all-resources.log

	warning "3. Try to export kind logs to $LOG_DIR..."
	kind export logs --name=${KindName} $LOG_DIR
	sudo chown -R $USER:$USER $LOG_DIR

	warning "debugInfo finished ! "
	warning "This means that some tests have failed. Please check the log. 🌚"
	debug=1
	exit 1
}
trap 'debugInfo $LINENO' ERR
trap 'debugInfo $LINENO' EXIT
debug=0

function cecho() {
	declare -A colors
	colors=(
		['black']='\E[0;47m'
		['red']='\E[0;31m'
		['green']='\E[0;32m'
		['yellow']='\E[0;33m'
		['blue']='\E[0;34m'
		['magenta']='\E[0;35m'
		['cyan']='\E[0;36m'
		['white']='\E[0;37m'
	)
	local defaultMSG="No message passed."
	local defaultColor="black"
	local defaultNewLine=true
	while [[ $# -gt 1 ]]; do
		key="$1"
		case $key in
		-c | --color)
			color="$2"
			shift
			;;
		-n | --noline)
			newLine=false
			;;
		*)
			# unknown option
			;;
		esac
		shift
	done
	message=${1:-$defaultMSG}     # Defaults to default message.
	color=${color:-$defaultColor} # Defaults to default color, if not specified.
	newLine=${newLine:-$defaultNewLine}
	echo -en "${colors[$color]}"
	echo -en "$message"
	if [ "$newLine" = true ]; then
		echo
	fi
	tput sgr0 #  Reset text attributes to normal without clearing screen.
	return
}

function warning() {
	cecho -c 'yellow' "$@"
}

function error() {
	cecho -c 'red' "$@"
}

function info() {
	cecho -c 'blue' "$@"
}

info "1. create kind cluster"
git clone https://github.com/bestchains/installer.git ${InstallDirPath}
cd ${InstallDirPath}
export IGNORE_FIXED_IMAGE_LOAD=YES
. ./scripts/kind.sh
if [ -f $ComponentImageFile ]; then
	info "reload component images from cache."
	source ${InstallDirPath}/scripts/cache-image.sh
	load_all_images $KindName /tmp/all.image.tar
fi

info "2. install component in kubernetes..."
info "2.1 install u4a component and u4a services"
export IGNORE_FABRIC_OPERATOR=YES
. ./scripts/e2e.sh
cd -

info "2.2 install fabric-operator"
docker tag hyperledgerk8s/fabric-operator:latest hyperledgerk8s/fabric-operator:example-e2e
kind load docker-image --name=${KindName} hyperledgerk8s/fabric-operator:example-e2e
export IMG=hyperledgerk8s/fabric-operator:example-e2e
export IMAGE_PULL_POLICY=IfNotPresent
make deploy
kubectl set env -n operator-system deployment/operator-controller-manager OPERATOR_INGRESS_DOMAIN=${ingressNodeIP}.nip.io
kubectl wait deploy -n operator-system operator-controller-manager --for condition=Available=True

info "3. create user and get user's token"
info "3.1 create all test users"
kubectl apply -f config/rbac/blockchain_cluster_role.yaml
# TODO oidc-server readinessProbe need readd.
function adduser() {
	START_TIME=$(date +%s)
	while true; do
		kubectl apply -f config/samples/users
		if [[ $? -eq 0 ]]; then
			break
		fi
		CURRENT_TIME=$(date +%s)
		ELAPSED_TIME=$((CURRENT_TIME - START_TIME))
		if [ $ELAPSED_TIME -gt $TimeoutSeconds ]; then
			error "Timeout reached"
			exit 1
		fi
		sleep 5
	done
}
adduser
# kubectl apply -f config/samples/users

Domain=$(kubectl get ing -n u4a-system bff-server-ingress --no-headers=true -o wide | awk '{print $3}')
function getToken() {
	local domain=$1
	local username=$2
	local password=$3
	UserNameBase64=$(echo -n ${username} | base64)
	PassWordBase64=$(echo -n ${password} | jq -sRr @uri | awk '{ printf "%s", $0 }' | base64)
	LoginURL=$(curl -Lks -o /dev/null -w %{url_effective} \
		"https://${domain}/oidc/auth?redirect_uri=https://${domain}/&response_type=code&scope=openid+profile+email+groups+offline_access")
	LoginResp=$(curl -Lks ${LoginURL} --data-raw '{"login":"'$UserNameBase64'","password":"'$PassWordBase64'","responseType":"JSON"}' \
		-H 'content-type: application/json;charset=UTF-8')
	StateURI=$(echo ${LoginResp} | jq -r .redirect)
	CodeURI=$(curl -Lks -o /dev/null -w %{url_effective} 'https://'$domain''$StateURI'')
	Code=$(echo $CodeURI | gawk -F '&' '{for(i=1; i<=NF; i++){if (match($i,"code")){print $i}}}' | awk -F "=" '{print $2}')
	query='query getToken($oidc: OidcTokenInput!) {\n token(oidc: $oidc) {\n access_token\n token_type\n expires_in\n refresh_token\n id_token\n }\n}'
	TokenResp=$(curl -Lks "https://${domain}/bff" --data-raw \
		'{"query":"'"$query"'","variables":{"oidc":{"grant_type":"authorization_code","redirect_uri":"https://'$domain'/","code":"'$Code'"}},"operationName":"getToken"}' \
		-H 'content-type: application/json;charset=UTF-8')
	Token=$(echo $TokenResp | jq -r .data.token.id_token)
}

info "3.2 get all test user's token"
getToken $Domain "org1admin" $DefaultPassWord
Admin1Token=$Token
getToken $Domain "org2admin" $DefaultPassWord
Admin2Token=$Token
getToken $Domain "org3admin" $DefaultPassWord
Admin3Token=$Token

info "3.3 get default ingress class and storage class"
IngressClassName=$(kubectl get ingressclass --no-headers | awk '{print $1}')
StorageClassName=$(kubectl get sc -o json |
	jq -r '.items[] | select(.metadata.annotations."storageclass.kubernetes.io/is-default-class"=true) | .metadata.name')

info "4. example test..."

info "4.1 create organizations: org1 org2 org3"
sed -i -e "s/<org1AdminToken>/${Admin1Token}/g" config/samples/orgs/org1.yaml
sed -i -e "s/<org2AdminToken>/${Admin2Token}/g" config/samples/orgs/org2.yaml
sed -i -e "s/<org3AdminToken>/${Admin3Token}/g" config/samples/orgs/org3.yaml

info "4.1.1 create org=org1, wait for the relevant components to start up."
kubectl create -f config/samples/orgs/org1.yaml --dry-run=client -o json |
	jq '.spec.caSpec.ingress.class = "'$IngressClassName'"' | jq '.spec.caSpec.storage.ca.class = "'$StorageClassName'"' |
	kubectl create --token=${Admin1Token} -f -
function waitOrgReady() {
	orgName=$1
	wantFedName=$2
	token=$3
	START_TIME=$(date +%s)
	while true; do
		status=$(kubectl get org $orgName --token=${token} --ignore-not-found=true -o json | jq -r .status.type)
		if [ "$status" == "Deployed" ]; then
			if [[ $wantFedName != "" ]]; then
				getFedName=$(kubectl get org $orgName --token=${token} --ignore-not-found=true -o json | jq -r '.status.federations[0]')
				if [[ $wantFedName == $getFedName ]]; then
					break
				fi
			fi
			break
		fi

		CURRENT_TIME=$(date +%s)
		ELAPSED_TIME=$((CURRENT_TIME - START_TIME))
		if [ $ELAPSED_TIME -gt $TimeoutSeconds ]; then
			error "Timeout reached"
			kubectl describe pod -n $orgName
			kubectl get org $orgName -oyaml
			kubectl get ibpca -n $orgName $orgName -oyaml
			exit 1
		fi
		sleep 5
	done
}
waitOrgReady org1 "" ${Admin1Token}

info "4.1.2 create org=org2, wait for the relevant components to start up."
kubectl create -f config/samples/orgs/org2.yaml --dry-run=client -o json |
	jq '.spec.caSpec.ingress.class = "'$IngressClassName'"' | jq '.spec.caSpec.storage.ca.class = "'$StorageClassName'"' |
	kubectl create --token=${Admin2Token} -f -
waitOrgReady org2 "" ${Admin2Token}

info "4.1.3 create org=org3, wait for the relevant components to start up."
kubectl create -f config/samples/orgs/org3.yaml --dry-run=client -o json |
	jq '.spec.caSpec.ingress.class = "'$IngressClassName'"' | jq '.spec.caSpec.storage.ca.class = "'$StorageClassName'"' |
	kubectl create --token=${Admin3Token} -f -
waitOrgReady org3 "" ${Admin3Token}

info "4.2 create federation resources: federation-sample"
kubectl create -f config/samples/ibp.com_v1beta1_federation.yaml --token=${Admin1Token}
function waitFed() {
	fedName=$1
	check=$2
	token=$3
	START_TIME=$(date +%s)
	while true; do
		if [[ $check == "Exist" ]]; then
			name=$(kubectl get fed --token=${token} $fedName --no-headers --ignore-not-found=true | awk '{print $1}')
			if [[ $name != "" ]]; then
				break
			fi
		elif [[ $check == "Activated" ]]; then
			status=$(kubectl get fed --token=${token} $fedName -o json | jq -r '.status.type')
			if [ "$status" == "FederationActivated" ]; then
				break
			fi
		fi
		CURRENT_TIME=$(date +%s)
		ELAPSED_TIME=$((CURRENT_TIME - START_TIME))
		if [ $ELAPSED_TIME -gt $TimeoutSeconds ]; then
			error "Timeout reached"
			exit 1
		fi
		sleep 5
	done
}
waitFed federation-sample "Exist" ${Admin1Token}
waitOrgReady "org1" "federation-sample" ${Admin1Token}
waitOrgReady "org1" "federation-sample" ${Admin1Token}

info "4.3 create federation create proposal for fed=federation-sample"

info "4.3.1 create proposal pro=create-federation-sample"
kubectl create -f config/samples/ibp.com_v1beta1_proposal_create_federation.yaml --token=${Admin1Token}

info "4.3.2 user=org2admin vote for pro=create-federation-sample"
function waitVoteExist() {
	ns=$1
	proposalName=$2
	token=$3
	START_TIME=$(date +%s)
	while true; do
		voteName=$(kubectl get vote --token=${token} -n $ns -l "bestchains.vote.proposal=$proposalName" --no-headers=true --ignore-not-found=true | awk '{print $1}')
		if [[ $voteName != "" ]]; then
			break
		fi
		CURRENT_TIME=$(date +%s)
		ELAPSED_TIME=$((CURRENT_TIME - START_TIME))
		if [ $ELAPSED_TIME -gt $TimeoutSeconds ]; then
			error "Timeout reached"
			exit 1
		fi
		sleep 5
	done
}
waitVoteExist org2 create-federation-sample ${Admin2Token}
kubectl patch vote -n org2 vote-org2-create-federation-sample --type='json' \
	-p='[{"op": "replace", "path": "/spec/decision", "value": true}]' --token=${Admin2Token}

info "4.3.3 pro=create-federation-sample become Succeeded"
function waitProposalSucceeded() {
	proposalName=$1
	token=$2
	START_TIME=$(date +%s)
	while true; do
		Type=$(kubectl get pro --token=${token} $proposalName --ignore-not-found=true -o json | jq -r '.status.conditions[] | select(.status=="True") | .type')
		if [[ $Type == "Succeeded" ]]; then
			break
		fi
		CURRENT_TIME=$(date +%s)
		ELAPSED_TIME=$((CURRENT_TIME - START_TIME))
		if [ $ELAPSED_TIME -gt $TimeoutSeconds ]; then
			error "Timeout reached"
			kubectl describe pro $proposalName
			exit 1
		fi
		sleep 5
	done
}
waitProposalSucceeded create-federation-sample ${Admin1Token}

info "4.3.4 fed=federation-sample become Activated, federation create finish!"
waitFed federation-sample "Activated" ${Admin1Token}

info "4.4 network management"
info "4.4.1 create single orderer node network"
sed -i -e "s/<org1AdminToken>/${Admin1Token}/g" config/samples/ibp.com_v1beta1_network.yaml
kubectl create -f config/samples/ibp.com_v1beta1_network.yaml --dry-run=client -o json |
	jq '.spec.orderSpec.ingress.class = "'$IngressClassName'"' | jq '.spec.orderSpec.storage.orderer.class = "'$StorageClassName'"' |
	kubectl create --token=${Admin1Token} -f -
function waitNetwork() {
	networkName=$1
	orderNs=$2
	want=$3
	channelName=$4
	token=$5
	START_TIME=$(date +%s)
	while true; do
		if [[ $want == "NoExist" ]]; then
			name=$(kubectl get network --token=${token} $networkName --no-headers=true --ignore-not-found=true | awk '{print $1}')
			if [[ $name == "" ]]; then
				break
			fi
		elif [[ $want == "Ready" ]]; then
			Type=$(kubectl get network ${networkName} --token=${token} --ignore-not-found=true -o json | jq -r '.status.type')
			if [[ $Type == "Deployed" ]]; then
				if [[ $channelName != "" ]]; then
					get=$(kubectl get network ${networkName} --token=${token} --ignore-not-found=true -o json | jq -r '.status.channels[0]')
					if [[ $get == $channelName ]]; then
						break
					fi
				fi
				break
			fi
		fi
		CURRENT_TIME=$(date +%s)
		ELAPSED_TIME=$((CURRENT_TIME - START_TIME))
		if [ $ELAPSED_TIME -gt $TimeoutSeconds ]; then
			error "Timeout reached"
			kubectl describe network $networkName
			exit 1
		fi
		sleep 5
	done
}
waitNetwork network-sample "org1" "Ready" "" ${Admin1Token}

info "4.4.2 create 3 orderer node network"
sed -i -e "s/<org1AdminToken>/${Admin1Token}/g" config/samples/ibp.com_v1beta1_network_size_3.yaml
kubectl create -f config/samples/ibp.com_v1beta1_network_size_3.yaml --dry-run=client -o json |
	jq '.spec.orderSpec.ingress.class = "'$IngressClassName'"' | jq '.spec.orderSpec.storage.orderer.class = "'$StorageClassName'"' |
	kubectl create --token=${Admin1Token} -f -
waitNetwork network-sample3 "org1" "Ready" "" ${Admin1Token}

info "4.4.3 delete network need create a federation dissolve network proposal for fed=federation-sample network=network-sample"

info "4.4.3.1 create proposal pro=dissolve-network-sample"
kubectl create -f config/samples/ibp.com_v1beta1_proposal_dissolve_network.yaml --token=${Admin1Token}

info "4.4.3.2 user=org2admin vote for pro=dissolve-network-sample"
waitVoteExist org2 dissolve-network-sample ${Admin2Token}
kubectl patch vote -n org2 vote-org2-dissolve-network-sample --type='json' -p='[{"op": "replace", "path": "/spec/decision", "value": true}]' --token=${Admin2Token}

info "4.4.3.3 pro=dissolve-network-sample become Activated"
waitProposalSucceeded dissolve-network-sample ${Admin2Token}

info "4.4.3.4 network=network-sample cant find, deletion finished"
waitNetwork network-sample "" "NoExist" "" ${Admin2Token}

info "4.7 channel management"
info "4.7.1 create channel channel=channel-sample"
kubectl create -f config/samples/ibp.com_v1beta1_channel_create.yaml --token=${Admin1Token}
function waitChannelReady() {
	channelName=$1
	want=$2
	token=$3
	START_TIME=$(date +%s)
	while true; do
		if [[ $want == "ChannelCreated" ]]; then
			Type=$(kubectl get channel --token=${token} $channelName --ignore-not-found=true -o json | jq -r '.status.type')
			if [[ $Type == $want ]]; then
				break
			fi
		elif [[ $want == "ChannelArchived" ]]; then
			Type=$(kubectl get channel --token=${token} $channelName --ignore-not-found=true -o json | jq -r '.status.type')
			if [[ $Type == $want ]]; then
				break
			fi
		fi
		CURRENT_TIME=$(date +%s)
		ELAPSED_TIME=$((CURRENT_TIME - START_TIME))
		if [ $ELAPSED_TIME -gt $TimeoutSeconds ]; then
			error "Timeout reached"
			kubectl describe channel $channelName
			exit 1
		fi
		sleep 5
	done
}
waitChannelReady channel-sample "ChannelCreated" ${Admin1Token}

info "4.7.2 create peer node peer=org1peer1"
Org1CaCert=$(kubectl get cm --token=${Admin1Token} -norg1 org1-connection-profile -ojson | jq -r '.binaryData."profile.json"' | base64 -d | jq -r '.tls.cert')
Org1CaURI=$(kubectl get cm --token=${Admin1Token} -norg1 org1-connection-profile -ojson | jq -r '.binaryData."profile.json"' | base64 -d | jq -r '.endpoints.api')
function parseURI() {
	uri=$1
	# https://stackoverflow.com/a/6174447/5939892
	# extract the protocol
	proto="$(echo $1 | grep :// | sed -e's,^\(.*://\).*,\1,g')"
	# remove the protocol
	url="$(echo ${1/$proto/})"
	# extract the user (if any)
	user="$(echo $url | grep @ | cut -d@ -f1)"
	# extract the host and port
	hostport="$(echo ${url/$user@/} | cut -d/ -f1)"
	# by request host without port
	host="$(echo $hostport | sed -e 's,:.*,,g')"
	# by request - try to extract the port
	port="$(echo $hostport | sed -e 's,^.*:,:,g' -e 's,.*:\([0-9]*\).*,\1,g' -e 's,[^0-9],,g')"
	# extract the path (if any)
	path="$(echo $url | grep / | cut -d/ -f2-)"
	if [[ $port == "" && $proto == "http" ]]; then
		port="80"
	elif [[ $port == "" && $proto == "https" ]]; then
		port="443"
	fi
}
parseURI ${Org1CaURI}
Org1CaHost=${host}
Org1CaPort=${port}
sed -i -e "s/<org1AdminToken>/${Admin1Token}/g" config/samples/peers/ibp.com_v1beta1_peer_org1peer1.yaml
sed -i -e "s/<org1-ca-cert>/${Org1CaCert}/g" config/samples/peers/ibp.com_v1beta1_peer_org1peer1.yaml
kubectl create -f config/samples/peers/ibp.com_v1beta1_peer_org1peer1.yaml --dry-run=client -o json |
	jq '.spec.ingress.class = "'$IngressClassName'"' |
	jq '.spec.storage.peer.class = "'$StorageClassName'"' | jq '.spec.storage.statedb.class = "'$StorageClassName'"' |
	jq '.spec.secret.enrollment.component.cahost = "'$Org1CaHost'"' | jq '.spec.secret.enrollment.tls.cahost = "'$Org1CaHost'"' |
	jq '.spec.secret.enrollment.component.caport = "'$Org1CaPort'"' | jq '.spec.secret.enrollment.tls.caport = "'$Org1CaPort'"' |
	kubectl create --token=${Admin1Token} -f -
function waitPeerReady() {
	peerName=$1
	ns=$2
	want=$3
	token=$4
	START_TIME=$(date +%s)
	while true; do
		if [[ $want == "" ]]; then
			Type=$(kubectl get ibppeer --token=${token} -n $ns $peerName --ignore-not-found=true -o json | jq -r '.status.type')
			if [[ $Type == "Deployed" ]]; then
				break
			fi
		fi
		CURRENT_TIME=$(date +%s)
		ELAPSED_TIME=$((CURRENT_TIME - START_TIME))
		if [ $ELAPSED_TIME -gt $TimeoutSeconds ]; then
			error "Timeout reached"
			kubectl describe --token=${token} ibppeer -n $ns $peerName
			exit 1
		fi
		sleep 5
	done
}
waitPeerReady org1peer1 org1 "" ${Admin1Token}

info "4.7.3 create peer node peer=org2peer1"
Org2CaCert=$(kubectl get cm --token=${Admin2Token} -norg2 org2-connection-profile -ojson | jq -r '.binaryData."profile.json"' | base64 -d | jq -r '.tls.cert')
Org2CaURI=$(kubectl get cm --token=${Admin2Token} -norg2 org2-connection-profile -ojson | jq -r '.binaryData."profile.json"' | base64 -d | jq -r '.endpoints.api')
parseURI ${Org2CaURI}
Org2CaHost=${host}
Org2CaPort=${port}
sed -i -e "s/<org2AdminToken>/${Admin2Token}/g" config/samples/peers/ibp.com_v1beta1_peer_org2peer1.yaml
sed -i -e "s/<org2-ca-cert>/${Org2CaCert}/g" config/samples/peers/ibp.com_v1beta1_peer_org2peer1.yaml
kubectl create -f config/samples/peers/ibp.com_v1beta1_peer_org2peer1.yaml --dry-run=client -o json |
	jq '.spec.ingress.class = "'$IngressClassName'"' |
	jq '.spec.storage.peer.class = "'$StorageClassName'"' | jq '.spec.storage.statedb.class = "'$StorageClassName'"' |
	jq '.spec.secret.enrollment.component.cahost = "'$Org2CaHost'"' | jq '.spec.secret.enrollment.tls.cahost = "'$Org2CaHost'"' |
	jq '.spec.secret.enrollment.component.caport = "'$Org2CaPort'"' | jq '.spec.secret.enrollment.tls.caport = "'$Org2CaPort'"' |
	kubectl create --token=${Admin2Token} -f -
waitPeerReady org2peer1 org2 "" ${Admin2Token}

info "4.7.4 add peer node to channel peer=org1peer1 channel=channel-sample"
kubectl apply -f config/samples/ibp.com_v1beta1_channel_join_org1.yaml --token=${Admin1Token}
# todo Verify that peers successfully join channel
sleep 5

info "4.7.5 add peer node to channel peer=org2peer1 channel=channel-sample"
kubectl apply -f config/samples/ibp.com_v1beta1_channel_join_org2.yaml --token=${Admin2Token}
# todo Verify that peers successfully join channel
sleep 5

info "4.7.6 create a proposal to archive channel-sample"
kubectl --token=${Admin1Token} apply -f config/samples/ibp.com_v1beta1_proposal_archive_channel.yaml

info "4.7.7 user=org2admin vote for pro=archive-channel-sample"
waitVoteExist org2 archive-channel-sample ${Admin2Token}
kubectl patch vote -n org2 vote-org2-archive-channel-sample --type='json' \
	-p='[{"op": "replace", "path": "/spec/decision", "value": true}]' --token=${Admin2Token}

info "4.7.8 pro=archive-channel-sample become Succeeded"
waitProposalSucceeded archive-channel-sample ${Admin1Token}

info "4.7.9 channel=channel-sample become Archived"
waitChannelReady channel-sample "ChannelArchived" ${Admin1Token}

info "4.7.10 create a proposal to unarchive channel-sample"
kubectl --token=${Admin1Token} apply -f config/samples/ibp.com_v1beta1_proposal_unarchive_channel.yaml

info "4.7.11 user=org2admin vote for pro=unarchive-channel-sample"
waitVoteExist org2 unarchive-channel-sample ${Admin2Token}
kubectl patch vote -n org2 vote-org2-unarchive-channel-sample --type='json' \
	-p='[{"op": "replace", "path": "/spec/decision", "value": true}]' --token=${Admin2Token}

info "4.7.12 pro=unarchive-channel-sample become Succeeded"
waitProposalSucceeded unarchive-channel-sample ${Admin1Token}

info "4.7.13 channel=channel-sample become Archived"
waitChannelReady channel-sample "ChannelCreated" ${Admin1Token}

info "cache component image"
source ${InstallDirPath}/scripts/cache-image.sh
save_all_images /tmp/all.image.tar /tmp/all.image.list
info "Do we need to update the cache image? $UPLOAD_IMAGE"
echo "UPLOAD_IMAGE=${UPLOAD_IMAGE}" >>$GITHUB_ENV

info "all finished! ✅"
