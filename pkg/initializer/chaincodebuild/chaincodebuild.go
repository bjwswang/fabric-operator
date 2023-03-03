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

package chaincodebuild

type Config struct {
	// PipelinRunNamespace namespace where pipelinerun executes
	PipelinRunNamespace string

	// Minio configurations
	MinioHost      string
	MinioAccessKey string
	MinioSecretKey string

	// PipelineRun tempaltes
	PipelineRunPVCFile string
	PipelineRunFile    string
}
