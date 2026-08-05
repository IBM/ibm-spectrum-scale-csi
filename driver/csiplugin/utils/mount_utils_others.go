//go:build !linux

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

// Stub implementations for non-Linux platforms (e.g. macOS) so the package
// compiles for local development and tooling without Linux-specific syscalls.

package utils

import (
	"context"
	"fmt"
)

// BindMount is not supported on non-Linux platforms.
func BindMount(_ context.Context, hostSource, _, target string) error {
	return fmt.Errorf("BindMount: not supported on this platform (source=%s target=%s)", hostSource, target)
}
