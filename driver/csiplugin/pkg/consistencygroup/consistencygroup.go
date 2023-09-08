/**
 * Copyright 2023 IBM Corp.
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

package consistencygroup

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

type StorageClassType int

const (
	ClassicStorageClass          StorageClassType = 0 // aka version 1 storage class
	ConsistencyGroupStorageClass StorageClassType = 1 // aka version 2 storage class
)

type VolumeType int

const (
	LightweightVolume             VolumeType = 0
	DependentFilesetBasedVolume   VolumeType = 1
	IndependentFilesetBasedVolume VolumeType = 2
)

var (
	ErrNoCsiVolume            = errors.New("no CSI volume")
	ErrInvalidCsiVolumeHandle = errors.New("invalid CSI volume handle format")
)

// VolumeHandle represents the VolumeHandle parameter that exists in the CSI PV spec.
type VolumeHandle struct {
	StorageClassType StorageClassType
	VolumeType       VolumeType
	ClusterID        string // ID of owning cluster where fileset resides
	FilesystemUID    string
	ConsistencyGroup string // Matches to the name of the independent fileset name. Format: <OCP cluster ID>-<namespace>
	FilesetName      string // Name of dependent fileset that represents the PV
	FilesetLinkPath  string
}

// GetVolumeHandle get's the CSI volume handle from a CSI PV source.
// VolumeHandle format:
// <storageclass_type>;<volume_type>;<cluster_id>;<filesystem_uuid>;<consistency_group>;<fileset_name>;<path>
func GetVolumeHandle(pvs *corev1.CSIPersistentVolumeSource) (VolumeHandle, error) {
	var vh VolumeHandle
	if pvs == nil {
		return vh, ErrNoCsiVolume
	}
	split := strings.Split(pvs.VolumeHandle, ";")
	if len(split) < 7 {
		return vh, ErrInvalidCsiVolumeHandle
	}
	i, err := strconv.Atoi(split[0])
	if err != nil {
		return vh, err
	}
	vh.StorageClassType = StorageClassType(i)
	i, err = strconv.Atoi(split[1])
	if err != nil {
		return vh, err
	}
	vh.VolumeType = VolumeType(i)
	vh.ClusterID = split[2]
	vh.FilesystemUID = split[3]
	vh.ConsistencyGroup = split[4]
	vh.FilesetName = split[5]
	vh.FilesetLinkPath = split[6]
	return vh, nil
}

// GetFilesystem reads the filesystem name from a CSI PV source.
func GetFilesystem(pvs *corev1.CSIPersistentVolumeSource) (string, error) {
	filesystem, ok := pvs.VolumeAttributes["volBackendFs"]
	if !ok {
		return filesystem, errors.New("CSI volume attribute 'volBackendFs' missing")
	}
	return filesystem, nil
}

// GetConsistencyGroupFileset reads the consistency group fileset name from a CSI PV source.
func GetConsistencyGroupFileset(pvs *corev1.CSIPersistentVolumeSource) (string, error) {
	volHandle, err := GetVolumeHandle(pvs)
	// The consistency group name is the same as the name of the independent fileset that represents the consistency group.
	return volHandle.ConsistencyGroup, err
}

// GetConsistencyGroupFilesetLinkPath returns the link path of the consistency group fileset.
func GetConsistencyGroupFilesetLinkPath(pvs *corev1.CSIPersistentVolumeSource) (string, error) {
	volHandle, err := GetVolumeHandle(pvs)
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(volHandle.FilesetLinkPath, "/"+volHandle.FilesetName) {
		return "", fmt.Errorf("unexpected format of fileset link path %s", volHandle.FilesetLinkPath)
	}
	// We can assume that the consistency group independent filesets link path is the parent directory of the PV fileset link path.
	// Querying the link path using mmlsfileset is not required.
	cgFsetLinkPath := strings.TrimSuffix(volHandle.FilesetLinkPath, "/"+volHandle.FilesetName)

	return cgFsetLinkPath, nil
}

var _ fmt.Stringer = VolumeHandle{}

// VolumeHandle implements fmt.Stringer interface to return the volume handle in string format
func (vh VolumeHandle) String() string {
	return fmt.Sprintf("%x;%x;%s;%s;%s;%s;%s", vh.StorageClassType, vh.VolumeType, vh.ClusterID, vh.FilesystemUID, vh.ConsistencyGroup, vh.FilesetName, vh.FilesetLinkPath)
}
