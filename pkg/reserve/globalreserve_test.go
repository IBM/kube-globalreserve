//
// Copyright 2020 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package reserve

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestHttpReserve(t *testing.T) {
	node0 := GetNode0()
	nodes := []*v1.Node{node0}
	gr := InitGR(nodes, nil, false)

	pod0 := GetPod("pod0", "1", "1000", "node0", v1.PodPending)
	pod1 := GetPod("pod1", "1", "1000", "node0", v1.PodPending)
	pod2 := GetPod("pod2", "1", "1000", "node0", v1.PodPending)
	pod3 := GetPod("pod3", "100m", "500", "node1", v1.PodPending)

	pods0 := []*v1.Pod{pod0, pod1, pod2, pod3}
	hosts0 := []string{"node0", "node0", "node0", "node1"}

	t.Run("HttpReserve reserve on non-exist node", func(t *testing.T) {
		ret := gr.ReservePods(pods0, hosts0)
		if len(ret.FailedPods) != 1 || ret.FailedPods[0] != "pod3" || len(ret.Error) <= 0 {
			t.Errorf("reserve on non-exist node failed")
		}
	})

	pods1 := []*v1.Pod{pod0, pod1, pod2}
	hosts1 := []string{"node0", "node0", "node0"}

	t.Run("HttpReserve reserve too much", func(t *testing.T) {
		ret := gr.ReservePods(pods1, hosts1)
		if len(ret.FailedPods) != 1 || ret.FailedPods[0] != "pod2" || len(ret.Error) <= 0 {
			t.Errorf("reserve too much failed")
		}
	})

	pods2 := []*v1.Pod{pod0, pod1}
	hosts2 := []string{"node0", "node0"}

	t.Run("HttpReserve reserve success", func(t *testing.T) {
		ret := gr.ReservePods(pods2, hosts2)
		if len(ret.FailedPods) > 0 || len(ret.Error) > 0 {
			t.Errorf("reserve success failed")
		}
	})

	t.Run("HttpReserve reserve success calculating", func(t *testing.T) {
		nodeAvai := gr.NodeCache["node0"].GetAvailable()
		cpuIndex := gr.ResTypeToID[v1.ResourceCPU]
		memIndex := gr.ResTypeToID[v1.ResourceMemory]
		podIndex := gr.ResTypeToID[v1.ResourcePods]
		if nodeAvai[cpuIndex] != 0 || nodeAvai[memIndex] != 3000 || nodeAvai[podIndex] != 8 {
			t.Errorf("reserve success calculating failed")
		}
	})
}

func TestReserve(t *testing.T) {
	node0 := GetNode0()
	nodes := []*v1.Node{node0}
	gr := InitGR(nodes, nil, false)

	pod0 := GetPod("pod0", "1", "1000", "node0", v1.PodPending)
	pod1 := GetPod("pod1", "2", "1000", "node0", v1.PodPending)

	t.Run("Reserve success", func(t *testing.T) {
		ret := gr.Reserve(pod0, "node0")
		if len(ret) > 0 || len(gr.PodToNode) != 1 || len(gr.NodeCache["node0"].Pods) != 1 {
			t.Errorf("Reserve success failed")
		}
	})

	t.Run("Reserve calculating success", func(t *testing.T) {
		nodeAvai := gr.NodeCache["node0"].GetAvailable()
		cpuIndex := gr.ResTypeToID[v1.ResourceCPU]
		memIndex := gr.ResTypeToID[v1.ResourceMemory]
		podIndex := gr.ResTypeToID[v1.ResourcePods]
		if nodeAvai[cpuIndex] != 1000 || nodeAvai[memIndex] != 4000 || nodeAvai[podIndex] != 9 {
			t.Errorf("Reserve success failed")
		}
	})

	t.Run("Reserve error", func(t *testing.T) {
		ret := gr.Reserve(pod1, "node0")
		if len(ret) == 0 || len(gr.PodToNode) != 1 || len(gr.NodeCache["node0"].Pods) != 1 {
			t.Errorf("Reserve error failed")
		}
	})

	t.Run("Reserve non-exist node", func(t *testing.T) {
		ret := gr.Reserve(pod1, "node1")
		if len(ret) == 0 || len(gr.PodToNode) != 1 || len(gr.NodeCache["node0"].Pods) != 1 {
			t.Errorf("Reserve non-exist node failed")
		}
	})
}

func TestUnreserve(t *testing.T) {
	node0 := GetNode0()
	nodes := []*v1.Node{node0}

	pod0 := GetPod("pod0", "1", "1000", "node0", v1.PodPending)
	pod1 := GetPod("pod1", "2", "1000", "node0", v1.PodPending)

	pods := []*v1.Pod{pod0, pod1}
	gr := InitGR(nodes, pods, true)

	t.Run("Unreserve success", func(t *testing.T) {
		gr.Unreserve(pod0, "node0")
		if len(gr.PodToNode) != 1 || len(gr.NodeCache["node0"].Pods) != 1 {
			t.Errorf("Unreserve success failed")
		}
	})
}
