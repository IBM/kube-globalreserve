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
	v1resource "k8s.io/apimachinery/pkg/api/resource"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

// ResourceTypeBuffer is used for adding new resource types
const ResourceTypeBuffer int = 3

// ReserveHTTPPathPrefix reserve url prefix
const ReserveHTTPPathPrefix string = "/reserve"

// UnreserveHTTPPathPrefix unreserve url prefix
const UnreserveHTTPPathPrefix string = "/unreserve"

// ReserveSchedulerName defines the empty scheduler name
const ReserveSchedulerName string = "schedulername_is_empty"

type resVector []int64

// PodsReserveRequest reserve http request data struct
type PodsReserveRequest struct {
	Pods          []*v1.Pod
	Nodes         []string
	SchedulerName string
}

// PodReserveResult reserve http return data struct
type PodReserveResult struct {
	FailedPods []string
	Error      string
}

// QuantityToInt translate k8s resource quantity into an integer
func QuantityToInt(name v1.ResourceName, res v1resource.Quantity) int64 {
	var ret int64 = 0
	if !v1helper.IsExtendedResourceName(name) {
		if name == v1.ResourceCPU {
			ret = res.MilliValue()
		} else {
			ret = res.Value()
		}
	} else {
		ret = res.Value()
	}
	return ret
}

// assignedPod selects pods that are assigned (scheduled and running).
func assignedPod(pod *v1.Pod) bool {
	return len(pod.Spec.NodeName) != 0
}

// VectorCompare compares 2 vectors, if a is larger than or equal to b, return true
func VectorCompare(a []int64, b []int64) bool {
	if len(a) < len(b) {
		return false
	}

	for i, v := range b {
		if v > a[i] {
			return false
		}
	}

	return true
}

// VectorMinus a minus b
func VectorMinus(a []int64, b []int64) {
	for i, v := range b {
		a[i] = a[i] - v
	}
}
