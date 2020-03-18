/**
 * Copyright 2019 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package settings

/*Setting parameter names as found in kubernetes Config Map or Storage Class
 */
const (
	FilesetType  string = "filesetType"
	InodeLimit   string = "inodeLimit"
	Uid          string = "uid"
	Gid          string = "gid"
	ClusterId    string = "clusterId"
	ParentFset   string = "parentFileset"
	VolBackendFs string = "volBackendFs"
	VolDirPath   string = "volDirBasePath"
)

/* deprecated setting parameter names
 */
const (
	FilesetTypeDep string = "fileset-type"
	InodeLimitDep  string = "inode-limit"
)
