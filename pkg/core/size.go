// Copyright 2026 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import "github.com/docker/go-units"

const kibPerMib = SizeKiB(units.MiB / units.KiB)

// SizeKiB is the canonical internal size unit used by PD.
type SizeKiB int64

// MiBToKiB converts MiB to KiB.
func MiBToKiB(size int64) SizeKiB {
	return SizeKiB(size) * kibPerMib
}

// BytesToKiB converts bytes to KiB.
func BytesToKiB(size uint64) SizeKiB {
	return SizeKiB(size / units.KiB)
}

// ToMiB converts KiB to MiB with floor division.
func (s SizeKiB) ToMiB() int64 {
	return int64(s / kibPerMib)
}
