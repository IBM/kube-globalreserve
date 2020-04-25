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

func TestNewNodeResInfo(t *testing.T) {
	node1 := GetNode1()

	t.Run("NodeResInfo creating new", func(t *testing.T) {
		riMap := GetResIDMap()
		ni := NewNodeResInfo(node1, riMap, len(riMap)+ResourceTypeBuffer)
		if ni.Name != "node1" || ni.Capa[0] != 2000 || ni.Capa[1] != 5000 || ni.Capa[2] != 10 ||
			ni.Capa[3] != 10 || len(ni.Pods) != 0 {
			t.Errorf("creating new failed")
		}
	})
}

func TestAddPodToCache(t *testing.T) {
	node1 := GetNode1()

	pod0 := GetPod("pod0", "1000m", "1000", "node0", v1.PodPending)
	pod1 := GetPod("pod1", "500m", "1000", "node0", v1.PodPending)

	riMap := GetResIDMap()
	ni := NewNodeResInfo(node1, riMap, len(riMap)+ResourceTypeBuffer)

	t.Run("NodeResInfo adding pod", func(t *testing.T) {
		ni.AddPodToCache(pod0, riMap)
		ni.AddPodToCache(pod1, riMap)
		if len(ni.Pods) != 2 {
			t.Errorf("adding pod failed with the number of pod")
		}

		pod0Cache := ni.Pods["NS1-pod0"]

		if pod0Cache.Source != ReserveSchedulerName || pod0Cache.Status != v1.PodPending || pod0Cache.Resources[0] != 1000 ||
			pod0Cache.Resources[1] != 1000 {
			t.Errorf("adding pod failed with pod cache")
		}
	})
}

func TestGetAvailable(t *testing.T) {
	node1 := GetNode1()

	riMap := GetResIDMap()
	ni := NewNodeResInfo(node1, riMap, len(riMap)+ResourceTypeBuffer)

	t.Run("NodeResInfo getting Availabel resources for empty node", func(t *testing.T) {
		na := ni.GetAvailable()
		if na[0] != 2000 || na[1] != 5000 || na[2] != 10 || na[3] != 10 {
			t.Errorf("getting Availabel resources for empty node failed")
		}
	})

	pod0 := GetPod("pod0", "1000m", "1000", "node0", v1.PodPending)

	ni.AddPodToCache(pod0, riMap)
	t.Run("NodeResInfo getting Availabel resources for node with 1 pending pod", func(t *testing.T) {
		na := ni.GetAvailable()
		if na[0] != 1000 || na[1] != 4000 || na[2] != 9 || na[3] != 10 {
			t.Errorf("getting Availabel resources for node with 1 pending pod failed")
		}
	})

	pod1 := GetPod("pod1", "1000m", "1000", "node0", v1.PodSucceeded)

	ni.AddPodToCache(pod1, riMap)
	t.Run("NodeResInfo getting Availabel resources for node with 1 succeeded pod", func(t *testing.T) {
		na := ni.GetAvailable()
		if na[0] != 1000 || na[1] != 4000 || na[2] != 9 || na[3] != 10 {
			t.Errorf("getting Availabel resources for node with 1 succeeded pod failed")
		}
	})
}

func TestCheckPod(t *testing.T) {
	node0 := GetNode0()

	pod0 := GetPod("pod0", "1000m", "1000", "node0", v1.PodPending)
	pod1 := GetPod("pod1", "1500m", "1000", "node0", v1.PodRunning)

	riMap := GetResIDMap()
	ni := NewNodeResInfo(node0, riMap, len(riMap)+ResourceTypeBuffer)
	ni.AddPodToCache(pod0, riMap)

	t.Run("NodeResInfo checking pod", func(t *testing.T) {

		checkRet := ni.CheckPod(pod1, riMap)

		if checkRet {
			t.Errorf("checking pod failed")
		}
	})
}
