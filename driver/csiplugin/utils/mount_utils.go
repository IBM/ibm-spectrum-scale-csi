/**
 * Copyright 2019, 2024 IBM Corp.
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

package utils

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
	"k8s.io/klog/v2"
)

// BindMount performs a bind mount of hostSource onto target using the system
// mount(8) binary, mirroring the behaviour of mount.Mounter.Mount with the
// "bind" option:
//
//  1. First pass: mount --bind <hostSource> <target>
//  2. Second pass: mount -o remount,bind[,ro,nodev,…] <hostSource> <target>
//     with any read-only / nodev / noexec / nosuid / noatime / relatime /
//     nodiratime flags inherited from the source filesystem via statfs(2).
//
// hostSource is the bare host path (e.g. /ibm/fs1/pvc-…/data) passed to the
// mount(8) binary — the external process resolves paths in the host mount
// namespace and does not see the /host prefix used inside the container.
//
// statfsSource is the /host-prefixed path (e.g. /host/ibm/fs1/pvc-…/data)
// used only for the in-process statfs(2) call, which succeeds because the
// container filesystem exposes host paths under /host.
func BindMount(ctx context.Context, hostSource, statfsSource, target string) error {
	loggerId := GetLoggerId(ctx)
	// First pass: establish the bind mount using the bare host path.
	if out, err := exec.Command("mount", "--bind", hostSource, target).CombinedOutput(); err != nil {
		klog.Errorf("[%s] mount --bind source [%s] target [%s] failed, Error: %v, Output: %s", loggerId, hostSource, target, err, string(out))
		return fmt.Errorf("mount --bind source [%s] target [%s] failed, Error: %v, Output: %s", hostSource, target, err, string(out))
	}

	// Derive the mount flags active on the source filesystem so the bind mount
	// inherits them (e.g. ro, nodev, nosuid, noexec, …). Use the /host-prefixed
	// path so statfs(2) resolves correctly inside the container.
	//
	// All options — including "remount" and "bind" — are passed as a single
	// comma-separated -o argument, which is the correct mount(8) syntax.
	// "--remount" is not a valid standalone flag for mount(8).
	mountOpts := []string{"remount", "bind"}
	st, err := StatfsWithTimeout(ctx, statfsSource)
	if err == nil {
		flagMap := map[int]string{
			unix.MS_RDONLY:     "ro",
			unix.MS_NODEV:      "nodev",
			unix.MS_NOEXEC:     "noexec",
			unix.MS_NOSUID:     "nosuid",
			unix.MS_NOATIME:    "noatime",
			unix.MS_RELATIME:   "relatime",
			unix.MS_NODIRATIME: "nodiratime",
		}
		for flag, opt := range flagMap {
			if int(st.Flags)&flag == flag {
				mountOpts = append(mountOpts, opt)
			}
		}
	}

	// Second pass: remount with inherited flags.
	opts := strings.Join(mountOpts, ",")
	if out, err := exec.Command("mount", "-o", opts, hostSource, target).CombinedOutput(); err != nil {
		klog.Errorf("[%s] mount -o %s %s %s failed: %v, Output: %s", loggerId, opts, hostSource, target, err, string(out))
		return fmt.Errorf("mount -o %s %s %s failed: %v, Output: %s", opts, hostSource, target, err, string(out))
	}

	return nil
}
