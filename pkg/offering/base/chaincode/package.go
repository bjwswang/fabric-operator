/*
 * Copyright contributors to the Hyperledger Fabric Operator project
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 * 	  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package chaincode

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"os"

	current "github.com/IBM-Blockchain/fabric-operator/api/v1beta1"
)

const (
	k8sPackageType  = "k8s"
	caasPakcageType = "caas"

	imageJson = `{
  "name": "%s",
  "digest": "%s"
}`
	connectionJson = `{
  "address": "%s",
  "dial_timeout": "10s",
  "tls_required": false
}`
	metadataJson = `{
  "type": "%s",
  "label": "%s"
}`
)

func (c *baseChaincode) PackageForK8s(instance *current.Chaincode) (string, error) {
	method := fmt.Sprintf("%s [base.chaincode.PackageForK8s]", stepPrefix)
	imageName := instance.Spec.Images.Name
	imageSha256 := instance.Spec.Images.Digest

	tmpDir := ChaincodeStorageDir("", instance)
	log.Info(fmt.Sprintf("%s package store dir %s\n", method, tmpDir))
	var (
		buf bytes.Buffer
		err error
	)
	if err = os.MkdirAll(tmpDir, 0755); err != nil {
		log.Error(err, "")
		return err.Error(), err
	}

	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	imageContent := fmt.Sprintf(imageJson, imageName, imageSha256)
	compressItems := []compressItem{
		{
			file:    "image.json",
			content: []byte(imageContent),
		},
	}
	log.Info(fmt.Sprintf("%s starting to compress first level", method))
	log.Info(fmt.Sprintf("%s compressItems %+v\n", method, compressItems))
	if err = compressFiles(tw, gw, compressItems); err != nil {
		return err.Error(), err
	}

	var b bytes.Buffer
	nextGW := gzip.NewWriter(&b)
	nextTw := tar.NewWriter(nextGW)

	metadataContent := fmt.Sprintf(metadataJson, k8sPackageType, instance.Spec.Label)
	compressItems = []compressItem{
		{
			file:    "code.tar.gz",
			content: buf.Bytes(),
		},
		{
			file:    "metadata.json",
			content: []byte(metadataContent),
		},
	}

	log.Info(fmt.Sprintf("%s starting to compress second level", method))
	log.Info(fmt.Sprintf("%s compressItems %+v\n", method, compressItems))
	if err = compressFiles(nextTw, nextGW, compressItems); err != nil {
		return err.Error(), err
	}

	writeFileName := ChaincodePacakgeFile(instance)
	absolutePath := fmt.Sprintf("%s/%s", tmpDir, writeFileName)
	log.Info(fmt.Sprintf("%s starting to write package info, path: %s\n", method, absolutePath))
	f, err := os.Create(absolutePath)
	if err != nil {
		log.Error(err, " try to create tar file error")
		return err.Error(), err
	}

	defer f.Close()
	_, err = f.Write(b.Bytes())
	return absolutePath, err
}

type compressItem struct {
	file    string
	content []byte
}

func compressFiles(tw *tar.Writer, gw *gzip.Writer, compressItems []compressItem) error {
	defer gw.Close()
	defer tw.Close()

	for _, item := range compressItems {
		hdr := tar.Header{
			Name: item.file,
			Mode: 0600,
			Size: int64(len(item.content)),
		}
		if err := tw.WriteHeader(&hdr); err != nil {
			log.Error(err, fmt.Sprintf("write %s file header error", item.file))
			return err
		}
		if _, err := tw.Write(item.content); err != nil {
			log.Error(err, fmt.Sprintf("write %s file content error", item.file))
		}
	}
	return nil
}
