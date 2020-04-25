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
)

// PodResInfo save pod infomation in NodeResInfo.Pods
type PodResInfo struct {
	Name      string
	Status    v1.PodPhase
	Resources resVector
	Source    string // which scheduler create this pod
}

// NewPodInfo reates a PodResInfo by pod
func NewPodInfo(pod *v1.Pod, resIDMap map[v1.ResourceName]int, resVecLen int) *PodResInfo {
	resources := make([]int64, resVecLen)
	GetPodReq(pod, resIDMap, resources)

	schedulerName := pod.Spec.SchedulerName
	if len(schedulerName) < 1 {
		schedulerName = ReserveSchedulerName
	}
	return &PodResInfo{
		Name:      pod.Name,
		Status:    pod.Status.Phase,
		Resources: resources,
		Source:    schedulerName,
	}
}

func addResourceList(list, new v1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}

// GetPodReq translates pod resource requrments into a int64 slice
func GetPodReq(pod *v1.Pod, resIDMap map[v1.ResourceName]int, resVec []int64 /*return value*/) {
	reqs := v1.ResourceList{}

	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
	}

	for resName, resValue := range reqs {
		resIndex := resIDMap[resName]
		resVec[resIndex] = QuantityToInt(resName, resValue)
	}

	resIndex := resIDMap[v1.ResourcePods]
	resVec[resIndex] = 1
}

// Dump for debugging
func (pr *PodResInfo) Dump() {
	klog.Infof("        %s, %s, %s  : %v", pr.Name, pr.Source, pr.Status, pr.Resources)
}
