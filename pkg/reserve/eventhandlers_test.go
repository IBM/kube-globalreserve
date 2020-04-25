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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAddNode(t *testing.T) {
	node0 := GetNode0()
	node1 := GetNode1()
	node2 := GetNode2()

	gr := InitGR(nil, nil, false)

	t.Run("GloalReserve empty", func(t *testing.T) {
		gr.AddNode(node0)
		if gr.NextResourceID != 0 || len(gr.NodeCache) != 0 {
			t.Errorf("the adding must be ignored")
		}
	})

	// insert 2 nodes in GloalReserve and test
	nodes := []*v1.Node{
		node0,
		node1,
	}

	gr = InitGR(nodes, nil, false)

	t.Run("GloalReserve Init", func(t *testing.T) {
		gr.CollectFromLister()
		if gr.NextResourceID != 4 || len(gr.NodeCache) != 2 {
			t.Errorf("GR Init failed.")
		}
	})

	t.Run("GloalReserve ignoring duplicated Node", func(t *testing.T) {
		gr.AddNode(node1)
		if gr.NextResourceID != 4 || len(gr.NodeCache) != 2 {
			t.Errorf("the adding duplicated node into GR failed")
		}
	})

	t.Run("GloalReserve adding new node Successful", func(t *testing.T) {
		gr.AddNode(node2)
		if gr.NextResourceID != 5 && len(gr.NodeCache) != 3 {
			t.Errorf("Adding new node into GR failed")
		}
	})
}

func TestDeleteNode(t *testing.T) {
	node0 := GetNode0()
	node1 := GetNode1()
	node2 := GetNode2()

	// insert 3 nodes in GloalReserve and test
	nodes := []*v1.Node{node0, node1, node2}

	pod0 := GetPod("pod0", "500m", "500", "node2", v1.PodPending)

	pods := []*v1.Pod{pod0}

	gr := InitGR(nodes, pods, true)

	node3 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node3",
			UID:  "node3",
		},
	}

	t.Run("GloalReserve Deleting non-existing node", func(t *testing.T) {
		gr.DeleteNode(node3)
		if gr.NextResourceID != 5 || len(gr.NodeCache) != 3 || len(gr.PodToNode) != 1 {
			t.Errorf("Deleting non-existing node from GR failed.")
		}
	})

	t.Run("GloalReserve Deleting one node", func(t *testing.T) {
		gr.DeleteNode(node2)
		if gr.NextResourceID != 5 || len(gr.NodeCache) != 2 || len(gr.PodToNode) != 0 {
			t.Errorf("Deleting node from GR failed.")
		}
	})
}

func TestAddPod(t *testing.T) {
	gr := InitGR(nil, nil, false)
	pod0 := GetPod("pod0", "1000m", "1000", "node0", v1.PodPending)

	t.Run("GloalReserve Adding pod into empty GR", func(t *testing.T) {
		gr.AddPod(pod0)
		if gr.NextResourceID != 0 || len(gr.NodeCache) != 0 {
			t.Errorf("Adding pod into empty GR failed.")
		}
	})

	node0 := GetNode0()

	//adding
	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "NS1",
			UID:       "NS1-pod1",
		},
		Spec: v1.PodSpec{
			NodeName: "node1",
		},
	}

	nodes := []*v1.Node{node0}
	gr = InitGR(nodes, nil, true)

	t.Run("GloalReserve Adding a pod into an non-exists node", func(t *testing.T) {
		gr.AddPod(pod1)
		if gr.NextResourceID != 3 || len(gr.NodeCache) != 1 || len(gr.NodeCache["node0"].Pods) != 0 {
			t.Errorf("Adding a pod into an non-exists node failed.")
		}
	})

	//adding pod successful
	t.Run("GloalReserve Adding a pod into node", func(t *testing.T) {
		gr.AddPod(pod0)
		if gr.NextResourceID != 3 || len(gr.NodeCache) != 1 || len(gr.NodeCache["node0"].Pods) != 1 {
			t.Errorf("Adding a pod into node failed.")
		}
	})
}

func TestUpdatePod(t *testing.T) {
	pod0 := GetPod("pod0", "1000m", "1000", "node0", v1.PodPending)
	node0 := GetNode0()

	nodes := []*v1.Node{node0}
	pods := []*v1.Pod{pod0}

	gr := InitGR(nodes, pods, true)

	pod1 := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "NS1",
			UID:       "NS1-pod1",
		},
		Spec: v1.PodSpec{
			NodeName:      "node0",
			SchedulerName: "non-binding",
		},
	}

	t.Run("GloalReserve update a non-exist pod", func(t *testing.T) {
		gr.UpdatePod(nil, pod1)
		if gr.NextResourceID != 3 || len(gr.NodeCache) != 1 || len(gr.NodeCache["node0"].Pods) != 2 {
			t.Errorf("Update a non-exists pod failed.")
		}
	})

	t.Run("GloalReserve update a pod into running", func(t *testing.T) {
		pod0.Status.Phase = v1.PodRunning
		gr.UpdatePod(nil, pod0)
		newStatus := gr.NodeCache["node0"].Pods["NS1-pod0"].Status
		if gr.NextResourceID != 3 || len(gr.NodeCache) != 1 || len(gr.NodeCache["node0"].Pods) != 2 ||
			newStatus != v1.PodRunning {
			t.Errorf("Update a pod into running failed.")
		}
	})

	t.Run("GloalReserve update a pod into succeed", func(t *testing.T) {
		pod0.Status.Phase = v1.PodSucceeded
		gr.UpdatePod(nil, pod0)
		newStatus := gr.NodeCache["node0"].Pods["NS1-pod0"].Status
		if gr.NextResourceID != 3 || len(gr.NodeCache) != 1 || len(gr.NodeCache["node0"].Pods) != 2 ||
			newStatus != v1.PodSucceeded {
			t.Errorf("Update a pod into succeed failed.")
		}
	})
}

func TestDeletePod(t *testing.T) {
	pod0 := GetPod("pod0", "1000m", "1000", "node0", v1.PodPending)
	node0 := GetNode0()

	nodes := []*v1.Node{node0}
	pods := []*v1.Pod{pod0}

	gr := InitGR(nodes, pods, true)

	t.Run("GloalReserve delete a pod from cache", func(t *testing.T) {
		gr.DeletePod(pod0)
		if gr.NextResourceID != 3 || len(gr.NodeCache) != 1 || len(gr.NodeCache["node0"].Pods) != 0 {
			t.Errorf("Delete a pod failed.")
		}
	})
}
