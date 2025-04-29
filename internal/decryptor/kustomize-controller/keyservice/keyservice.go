// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package keyservice

import (
	"github.com/getsops/sops/v3/age"
	"github.com/getsops/sops/v3/keys"
	"github.com/getsops/sops/v3/pgp"
)

// IsOfflineMethod returns true for offline decrypt methods or false otherwise.
func IsOfflineMethod(mk keys.MasterKey) bool {
	switch mk.(type) {
	case *pgp.MasterKey, *age.MasterKey:
		return true
	default:
		return false
	}
}
