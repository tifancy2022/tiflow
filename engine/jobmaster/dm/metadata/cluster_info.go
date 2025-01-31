// Copyright 2022 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package metadata

import (
	"context"

	"github.com/coreos/go-semver/semver"
	"github.com/pingcap/tiflow/engine/pkg/adapter"
	metaModel "github.com/pingcap/tiflow/engine/pkg/meta/model"
)

// ClusterInfo represents the cluster info.
type ClusterInfo struct {
	State

	Version semver.Version
}

// NewClusterInfo creates a new ClusterInfo instance.
func NewClusterInfo(version semver.Version) *ClusterInfo {
	return &ClusterInfo{
		Version: version,
	}
}

// ClusterInfoStore manages the state of ClusterInfo.
type ClusterInfoStore struct {
	*TomlStore
}

// NewClusterInfoStore returns a new ClusterInfoStore instance
func NewClusterInfoStore(kvClient metaModel.KVClient) *ClusterInfoStore {
	clusterInfoStore := &ClusterInfoStore{
		TomlStore: NewTomlStore(kvClient),
	}
	clusterInfoStore.TomlStore.Store = clusterInfoStore
	return clusterInfoStore
}

// CreateState creates an empty ClusterInfo object
func (clusterInfoStore *ClusterInfoStore) CreateState() State {
	return &ClusterInfo{}
}

// Key returns encoded key of ClusterInfo state store
func (clusterInfoStore *ClusterInfoStore) Key() string {
	return adapter.DMInfoKeyAdapter.Encode()
}

// UpdateVersion updates the version of ClusterInfo.
func (clusterInfoStore *ClusterInfoStore) UpdateVersion(ctx context.Context, newVer semver.Version) error {
	state, err := clusterInfoStore.Get(ctx)
	if err != nil {
		return err
	}
	clusterInfo := state.(*ClusterInfo)
	clusterInfo.Version = newVer
	return clusterInfoStore.Put(ctx, clusterInfo)
}
