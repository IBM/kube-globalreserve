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
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	schedulerlisters "k8s.io/kubernetes/pkg/scheduler/listers"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

// FakeNodeInfoLister declares a schedulernodeinfo.NodeInfo type for testing.
type FakeNodeInfoLister []*schedulernodeinfo.NodeInfo

// Get returns a fake node object in the fake nodes.
func (nodes FakeNodeInfoLister) Get(nodeName string) (*schedulernodeinfo.NodeInfo, error) {
	for _, node := range nodes {
		if node != nil && node.Node().Name == nodeName {
			return node, nil
		}
	}
	return nil, fmt.Errorf("unable to find node: %s", nodeName)
}

// List lists all nodes.
func (nodes FakeNodeInfoLister) List() ([]*schedulernodeinfo.NodeInfo, error) {
	return nodes, nil
}

// HavePodsWithAffinityList is supposed to list nodes with at least one pod with affinity. For the fake lister
// we just return everything.
func (nodes FakeNodeInfoLister) HavePodsWithAffinityList() ([]*schedulernodeinfo.NodeInfo, error) {
	return nodes, nil
}

// StubNodeInfoLister this file is only used for running go test
func StubNodeInfoLister(nodes []*v1.Node) schedulerlisters.NodeInfoLister {
	nodeInfoList := make([]*schedulernodeinfo.NodeInfo, 0, len(nodes))
	for _, node := range nodes {
		nodeInfo := schedulernodeinfo.NewNodeInfo()
		nodeInfo.SetNode(node)
		nodeInfoList = append(nodeInfoList, nodeInfo)
	}

	return FakeNodeInfoLister(nodeInfoList)
}

// FakePodInfoLister uses for simulating pod list
type FakePodInfoLister []*v1.Pod

// List returns []*v1.Pod matching a query.
func (f FakePodInfoLister) List(s labels.Selector) (selected []*v1.Pod, err error) {
	for _, pod := range f {
		if s.Matches(labels.Set(pod.Labels)) {
			selected = append(selected, pod)
		}
	}
	return selected, nil
}

// FilteredList returns pods matching a pod filter and a label selector.
func (f FakePodInfoLister) FilteredList(podFilter schedulerlisters.PodFilter, s labels.Selector) (selected []*v1.Pod, err error) {
	for _, pod := range f {
		if podFilter(pod) && s.Matches(labels.Set(pod.Labels)) {
			selected = append(selected, pod)
		}
	}
	return selected, nil
}

// StubPodInfoLister return a lister
func StubPodInfoLister(pods []*v1.Pod) FakePodInfoLister {
	return FakePodInfoLister(pods)
}

// InitGR creates a GloalReserve for UT.
func InitGR(nodes []*v1.Node, pods []*v1.Pod, collected bool) *GloalReserve {
	gr := &GloalReserve{
		ResTypeToID:    make(map[v1.ResourceName]int),
		ResTypeMaxKind: 0,
		NextResourceID: 0,
		NodeCache:      make(map[string]*NodeResInfo),
		PodToNode:      make(map[types.UID]string),
		NodeLister:     nil,
		PodLister:      nil,
	}

	gr.NodeLister = StubNodeInfoLister(nodes)
	gr.PodLister = StubPodInfoLister(pods)

	if collected {
		gr.CollectFromLister()
	}

	return gr
}

// InitHTTPServer creates GlobalReserve http server for UT
func InitHTTPServer(gr *GloalReserve) *http.Server {
	// start the http server to receive 3rd party reserve request
	router := httprouter.New()
	router.POST(ReserveHTTPPathPrefix, AddReserveRoute(gr))
	router.POST(UnreserveHTTPPathPrefix, AddUnreserveRoute(gr))

	server := &http.Server{Addr: ":23456", Handler: router}

	go func() {
		server.ListenAndServe()
	}()

	return server
}

// GetResIDMap return a map between resource type and index for UT
func GetResIDMap() map[v1.ResourceName]int {
	tt := map[v1.ResourceName]int{
		v1.ResourceCPU:     0,
		v1.ResourceMemory:  1,
		v1.ResourcePods:    2,
		"ExtendedResType0": 3,
		"ExtendedResType1": 4,
	}
	return tt
}

// GetNode0 returs a node for UT
func GetNode0() *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node0",
			UID:  "node0",
		},
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    *(resource.NewQuantity(2, resource.DecimalSI)),
				v1.ResourceMemory: *(resource.NewQuantity(5000, resource.DecimalSI)),
				v1.ResourcePods:   *(resource.NewQuantity(10, resource.DecimalSI)),
			},
		},
	}
}

// GetNode1 returs a node for UT
func GetNode1() *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			UID:  "node1",
		},
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:     *(resource.NewQuantity(2, resource.DecimalSI)),
				v1.ResourceMemory:  *(resource.NewQuantity(5000, resource.DecimalSI)),
				v1.ResourcePods:    *(resource.NewQuantity(10, resource.DecimalSI)),
				"ExtendedResType0": *(resource.NewQuantity(10, resource.DecimalSI)),
			},
		},
	}
}

// GetNode2 returs a node for UT
func GetNode2() *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node2",
			UID:  "node2",
		},
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:     *(resource.NewQuantity(2, resource.DecimalSI)),
				v1.ResourceMemory:  *(resource.NewQuantity(5000, resource.DecimalSI)),
				v1.ResourcePods:    *(resource.NewQuantity(10, resource.DecimalSI)),
				"ExtendedResType1": *(resource.NewQuantity(10, resource.DecimalSI)),
			},
		},
	}
}

// GetPod returns a pod
func GetPod(name string, cpuStr string, memStr string, nodename string, sta v1.PodPhase) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "NS1",
			UID:       types.UID("NS1-" + name),
		},
		Spec: v1.PodSpec{
			NodeName: nodename,
			Containers: []v1.Container{
				{
					Name: "Container0",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse(cpuStr),
							v1.ResourceMemory: resource.MustParse(memStr),
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: sta,
		},
	}
}
