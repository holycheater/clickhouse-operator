// Copyright 2019 Altinity Ltd and/or its affiliates. All rights reserved.
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

package macro

import (
	"strconv"
	"strings"

	api "github.com/altinity/clickhouse-operator/pkg/apis/clickhouse.altinity.com/v1"
	"github.com/altinity/clickhouse-operator/pkg/model/chi/namer/short"
	"github.com/altinity/clickhouse-operator/pkg/util"
)

// MacrosEngine
type MacrosEngine struct {
	namer *short.Namer
	scope any
}

// Macro
func Macro(scope any) *MacrosEngine {
	m := new(MacrosEngine)
	m.namer = short.NewNamer(short.TargetNames)
	m.scope = scope
	return m
}

// Line expands line with macros(es)
func (m *MacrosEngine) Line(line string) string {
	switch t := m.scope.(type) {
	case api.ICustomResource:
		return m.newLineMacroReplacerCR(t).Replace(line)
	case api.ICluster:
		return m.newLineMacroReplacerCluster(t).Replace(line)
	case api.IShard:
		return m.newLineMacroReplacerShard(t).Replace(line)
	case api.IHost:
		return m.newLineMacroReplacerHost(t).Replace(line)
	default:
		return "unknown scope"
	}
}

// Map expands map with macros(es)
func (m *MacrosEngine) Map(_map map[string]string) map[string]string {
	switch t := m.scope.(type) {
	case api.ICustomResource:
		return m.newMapMacroReplacerCR(t).Replace(_map)
	case api.ICluster:
		return m.newMapMacroReplacerCluster(t).Replace(_map)
	case api.IShard:
		return m.newMapMacroReplacerShard(t).Replace(_map)
	case api.IHost:
		return m.newMapMacroReplacerHost(t).Replace(_map)
	default:
		return map[string]string{
			"unknown scope": "unknown scope",
		}
	}
}

// newLineMacroReplacerCR
func (m *MacrosEngine) newLineMacroReplacerCR(cr api.ICustomResource) *strings.Replacer {
	return strings.NewReplacer(
		MacrosNamespace, m.namer.Name(short.Namespace, cr),
		MacrosChiName, m.namer.Name(short.CRName, cr),
	)
}

// newMapMacroReplacerCR
func (m *MacrosEngine) newMapMacroReplacerCR(cr api.ICustomResource) *util.MapReplacer {
	return util.NewMapReplacer(m.newLineMacroReplacerCR(cr))
}

// newLineMacroReplacerCluster
func (m *MacrosEngine) newLineMacroReplacerCluster(cluster api.ICluster) *strings.Replacer {
	return strings.NewReplacer(
		MacrosNamespace, m.namer.Name(short.Namespace, cluster),
		MacrosChiName, m.namer.Name(short.CRName, cluster),
		MacrosClusterName, m.namer.Name(short.ClusterName, cluster),
		MacrosClusterIndex, strconv.Itoa(cluster.GetRuntime().GetAddress().GetClusterIndex()),
	)
}

// newMapMacroReplacerCluster
func (m *MacrosEngine) newMapMacroReplacerCluster(cluster api.ICluster) *util.MapReplacer {
	return util.NewMapReplacer(m.newLineMacroReplacerCluster(cluster))
}

// newLineMacroReplacerShard
func (m *MacrosEngine) newLineMacroReplacerShard(shard api.IShard) *strings.Replacer {
	return strings.NewReplacer(
		MacrosNamespace, m.namer.Name(short.Namespace, shard),
		MacrosChiName, m.namer.Name(short.CRName, shard),
		MacrosClusterName, m.namer.Name(short.ClusterName, shard),
		MacrosClusterIndex, strconv.Itoa(shard.GetRuntime().GetAddress().GetClusterIndex()),
		MacrosShardName, m.namer.Name(short.ShardName, shard),
		MacrosShardIndex, strconv.Itoa(shard.GetRuntime().GetAddress().GetShardIndex()),
	)
}

// newMapMacroReplacerShard
func (m *MacrosEngine) newMapMacroReplacerShard(shard api.IShard) *util.MapReplacer {
	return util.NewMapReplacer(m.newLineMacroReplacerShard(shard))
}

// clusterScopeIndexOfPreviousCycleTail gets cluster-scope index of previous cycle tail
func clusterScopeIndexOfPreviousCycleTail(host api.IHost) int {
	if host.GetRuntime().GetAddress().GetClusterScopeCycleOffset() == 0 {
		// This is the cycle head - the first host of the cycle
		// We need to point to previous host in this cluster - which would be previous cycle tail

		if host.GetRuntime().GetAddress().GetClusterScopeIndex() == 0 {
			// This is the very first host in the cluster - head of the first cycle
			// No previous host available, so just point to the same host, mainly because label must be an empty string
			// or consist of alphanumeric characters, '-', '_' or '.', and must start and end with an alphanumeric character
			// So we can't set it to "-1"
			return host.GetRuntime().GetAddress().GetClusterScopeIndex()
		}

		// This is head of non-first cycle, point to previous host in the cluster - which would be previous cycle tail
		return host.GetRuntime().GetAddress().GetClusterScopeIndex() - 1
	}

	// This is not cycle head - just point to the same host
	return host.GetRuntime().GetAddress().GetClusterScopeIndex()
}

// newLineMacroReplacerHost
func (m *MacrosEngine) newLineMacroReplacerHost(host api.IHost) *strings.Replacer {
	return strings.NewReplacer(
		MacrosNamespace, m.namer.Name(short.Namespace, host),
		MacrosChiName, m.namer.Name(short.CRName, host),
		MacrosClusterName, m.namer.Name(short.ClusterName, host),
		MacrosClusterIndex, strconv.Itoa(host.GetRuntime().GetAddress().GetClusterIndex()),
		MacrosShardName, m.namer.Name(short.ShardName, host),
		MacrosShardIndex, strconv.Itoa(host.GetRuntime().GetAddress().GetShardIndex()),
		MacrosShardScopeIndex, strconv.Itoa(host.GetRuntime().GetAddress().GetShardScopeIndex()), // TODO use appropriate namePart function
		MacrosReplicaName, m.namer.Name(short.ReplicaName, host),
		MacrosReplicaIndex, strconv.Itoa(host.GetRuntime().GetAddress().GetReplicaIndex()),
		MacrosReplicaScopeIndex, strconv.Itoa(host.GetRuntime().GetAddress().GetReplicaScopeIndex()), // TODO use appropriate namePart function
		MacrosHostName, m.namer.Name(short.HostName, host),
		MacrosChiScopeIndex, strconv.Itoa(host.GetRuntime().GetAddress().GetCRScopeIndex()), // TODO use appropriate namePart function
		MacrosChiScopeCycleIndex, strconv.Itoa(host.GetRuntime().GetAddress().GetCRScopeCycleIndex()), // TODO use appropriate namePart function
		MacrosChiScopeCycleOffset, strconv.Itoa(host.GetRuntime().GetAddress().GetCRScopeCycleOffset()), // TODO use appropriate namePart function
		MacrosClusterScopeIndex, strconv.Itoa(host.GetRuntime().GetAddress().GetClusterScopeIndex()), // TODO use appropriate namePart function
		MacrosClusterScopeCycleIndex, strconv.Itoa(host.GetRuntime().GetAddress().GetClusterScopeCycleIndex()), // TODO use appropriate namePart function
		MacrosClusterScopeCycleOffset, strconv.Itoa(host.GetRuntime().GetAddress().GetClusterScopeCycleOffset()), // TODO use appropriate namePart function
		MacrosClusterScopeCycleHeadPointsToPreviousCycleTail, strconv.Itoa(clusterScopeIndexOfPreviousCycleTail(host)),
	)
}

// newMapMacroReplacerHost
func (m *MacrosEngine) newMapMacroReplacerHost(host api.IHost) *util.MapReplacer {
	return util.NewMapReplacer(m.newLineMacroReplacerHost(host))
}