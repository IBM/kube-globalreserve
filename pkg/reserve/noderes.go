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
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/types"
)

// NodeResInfo saves node data in GloalReserve.NodeCache
type NodeResInfo struct {
	Name string
	Capa resVector
	Pods map[types.UID]*PodResInfo
}

// NewNodeResInfo create a NodeResInfo by v1.Node
func NewNodeResInfo(node *v1.Node, resIDMap map[v1.ResourceName]int, resVecLen int) *NodeResInfo {
	// Nodes' res map to vectors [0 0 0 100 300 100 0]
	resources := make([]int64, resVecLen)
	for resName, resValue := range node.Status.Allocatable {
		resIndex := resIDMap[resName]
		resources[resIndex] = QuantityToInt(resName, resValue)
	}

	return &NodeResInfo{
		Name: node.Name,
		Capa: resources,
		Pods: make(map[types.UID]*PodResInfo),
	}
}

// Dump for debugging
func (nr *NodeResInfo) Dump() {
	klog.Infof("    %s : %v", nr.Name, nr.Capa)
	for _, podR := range nr.Pods {
		podR.Dump()
	}
}

// AddPodToCache adds one pod into this NodeResInfo
func (nr *NodeResInfo) AddPodToCache(pod *v1.Pod, resIDMap map[v1.ResourceName]int) {
	// AddPod does not check node's avilable resources, assue the pod can be binded to this node
	podKey := pod.UID

	nr.Pods[podKey] = NewPodInfo(pod, resIDMap, cap(nr.Capa))
	klog.V(3).Infof("Add pod %s on node %s", podKey, nr.Name)
	klog.V(4).Infof("Available: %v", nr.GetAvailable())
}

// UpdatePod updates one pod in this NodeResInfo
func (nr *NodeResInfo) UpdatePod(pod *v1.Pod, resIDMap map[v1.ResourceName]int) {
	podKey := pod.UID
	if podInfo, ok := nr.Pods[pod.UID]; ok {
		podInfo.Status = pod.Status.Phase
		klog.V(3).Infof("Update pod %s on node %s", podKey, nr.Name)
	} else {
		klog.V(3).Infof("Pod %s can not be found on host %s", pod.Name, nr.Name)
	}

	klog.V(4).Infof("Available: %v", nr.GetAvailable())
}

// DeletePod deletes the pod from this NodeResInfo
func (nr *NodeResInfo) DeletePod(pod *v1.Pod) {
	podKey := pod.UID
	if _, ok := nr.Pods[podKey]; ok {
		klog.V(3).Infof("Delete pod %s from node %s", podKey, nr.Name)
		delete(nr.Pods, podKey)
	}

	klog.V(4).Infof("Available: %v", nr.GetAvailable())
}

// CheckPod checks there is enough resources in this NodeResInfo
func (nr *NodeResInfo) CheckPod(pod *v1.Pod, resIDMap map[v1.ResourceName]int) bool {
	nodeAvailable := nr.GetAvailable()
	klog.V(3).Infof("CheckPod availabe: %v", nodeAvailable)
	podReq := make([]int64, len(nr.Capa))
	GetPodReq(pod, resIDMap, podReq)
	klog.V(3).Infof("pod request: %v", podReq)

	return VectorCompare(nodeAvailable, podReq)
}

// GetAvailable caculate the free resources in this NodeResInfo, succeeded and failed pods are ignored
func (nr *NodeResInfo) GetAvailable() []int64 {
	avaiRes := make([]int64, cap(nr.Capa))
	copy(avaiRes, nr.Capa)
	for _, pod := range nr.Pods {
		if pod.Status != v1.PodSucceeded && pod.Status != v1.PodFailed {
			for i, v := range pod.Resources {
				avaiRes[i] -= v
			}
		}
	}

	return avaiRes
}
