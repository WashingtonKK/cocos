// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

//go:build embed
// +build embed

package manager

import (
	"context"

	backendinfo "github.com/ultravioletrs/cocos/scripts/backend_info"
)

func (ms *managerService) FetchBackendInfo(_ context.Context, _ string) ([]byte, error) {
	return backendinfo.BackendInfo, nil
}
