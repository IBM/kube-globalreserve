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
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

// AddNode into GlobalReserve, called by Node watcher
func (gr *GloalReserve) AddNode(obj interface{}) {
	if gr.NextResourceID == 0 {
		klog.V(3).Infof("CollectFromLister is not called.")
		return
	}

	node, ok := obj.(*v1.Node)
	if !ok {
		klog.Errorf("cannot convert to *v1.Node: %v", obj)
		return
	}

	if _, ok := gr.NodeCache[node.Name]; ok {
		klog.V(3).Infof("Node %s is already exist.", node.Name)
		return
	}

	klog.V(3).Infof("add event for node %s ", node.Name)

	gr.mu.Lock()
	defer gr.mu.Unlock()

	//check the resouce name, ensure the name is already in ResTypeToID
	for resName := range node.Status.Allocatable {
		if _, ok := gr.ResTypeToID[resName]; !ok {
			// new resource name is imported, need to enlarge the resource slice capability
			if gr.NextResourceID >= gr.ResTypeMaxKind {
				// the resource slice length is not enough, must enlarge all nodes and pods resource slices
				klog.Errorf("The number of Resource Types is larger than slices capability, ignore %s.", resName)
				continue
			}
			gr.ResTypeToID[resName] = gr.NextResourceID
			gr.NextResourceID++
		}
	}

	gr.NodeCache[node.Name] = NewNodeResInfo(node, gr.ResTypeToID, gr.ResTypeMaxKind)
}

// UpdateNode in GlobalReserve, called by Node watcher, this function ignores the events currently
func (gr *GloalReserve) UpdateNode(oldObj, newObj interface{}) {
	//global reserve does not care about node updating event, ignore it.
	return
}

// DeleteNode from GlobalReserve, called by Node watcher
func (gr *GloalReserve) DeleteNode(obj interface{}) {
	var node *v1.Node
	switch t := obj.(type) {
	case *v1.Node:
		node = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		node, ok = t.Obj.(*v1.Node)
		if !ok {
			klog.Errorf("cannot convert to *v1.Node: %v", t.Obj)
			return
		}
	default:
		klog.Errorf("cannot convert to *v1.Node: %v", t)
		return
	}
	klog.V(3).Infof("delete event for node %q", node.Name)

	gr.mu.Lock()
	defer gr.mu.Unlock()
	if nodeInfo, ok := gr.NodeCache[node.Name]; ok {
		for podUID := range nodeInfo.Pods {
			delete(gr.PodToNode, podUID)
		}
		delete(gr.NodeCache, node.Name)
	}
}

// AddPod into GloalReserve, most of the time pod is added without binding and
// this funcion ignores them
func (gr *GloalReserve) AddPod(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		klog.Errorf("cannot convert to *v1.Pod: %v", obj)
		return
	}
	klog.V(3).Infof("add event for scheduled pod %s/%s UID: %s ", pod.Namespace, pod.Name, pod.UID)

	if gr.NextResourceID == 0 {
		klog.V(3).Infof("CollectFromLister is not called.")
		return
	}

	hostname := pod.Spec.NodeName
	if len(hostname) > 0 {
		gr.mu.Lock()
		defer gr.mu.Unlock()

		if nodeInfo, ok := gr.NodeCache[hostname]; !ok {
			// normally, it is impossible to enter this part
			klog.Warningf("Pod %s Namespace %s host %s, but host does not exist in cache", pod.Name, pod.Namespace, hostname)
			/*
				nodecache := NewEmptyNodeResInfo(hostname, gr.ResTypeMaxKind)
				nodecache.AddPodToCache(pod, gr.ResTypeToID, ReserveSchedulerName)
				gr.NodeCache[hostname] = nodecache
			*/
		} else {
			podKey := pod.UID
			if _, ok := nodeInfo.Pods[podKey]; !ok {
				nodeInfo.AddPodToCache(pod, gr.ResTypeToID)
			}
		}
	}
}

//UpdatePod from GloalReserve, the pod binding host may be changed at some extreme cases
func (gr *GloalReserve) UpdatePod(oldObj, newObj interface{}) {
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		klog.Errorf("cannot convert newObj to *v1.Pod: %v", newObj)
		return
	}

	newHostname := newPod.Spec.NodeName
	if len(newHostname) > 0 {
		addFlag := false
		gr.mu.Lock()
		defer gr.mu.Unlock()

		if oldHostName, ok := gr.PodToNode[newPod.UID]; !ok {
			klog.Warningf("Pod %s namespace %s is not reserved by scheduler %s before binding",
				newPod.Name, newPod.Namespace, newPod.Spec.SchedulerName)
			gr.PodToNode[newPod.UID] = newHostname
			addFlag = true
		} else {
			if oldHostName != newHostname {
				// hostname is changed, move the old pod data into new host
				addFlag = true
				delete(gr.NodeCache[oldHostName].Pods, newPod.UID)
			}
		}

		//update state
		if nodeCache, ok := gr.NodeCache[newHostname]; ok {
			if addFlag {
				nodeCache.AddPodToCache(newPod, gr.ResTypeToID)
			} else {
				nodeCache.UpdatePod(newPod, gr.ResTypeToID)
			}
		}
	}
}

// DeletePod from GloalReserve
func (gr *GloalReserve) DeletePod(obj interface{}) {
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			klog.Errorf("cannot convert to *v1.Pod: %v", t.Obj)
			return
		}
	default:
		klog.Errorf("cannot convert to *v1.Pod: %v", t)
		return
	}
	klog.V(3).Infof("delete event for scheduled pod %s/%s ", pod.Namespace, pod.Name)
	// NOTE: Updates must be written to scheduler cache before invalidating
	// equivalence cache, because we could snapshot equivalence cache after the
	// invalidation and then snapshot the cache itself. If the cache is
	// snapshotted before updates are written, we would update equivalence
	// cache with stale information which is based on snapshot of old cache.
	delete(gr.PodToNode, pod.UID)
	hostname := pod.Spec.NodeName
	if len(hostname) > 0 {
		gr.mu.Lock()
		defer gr.mu.Unlock()
		if nodecache, ok := gr.NodeCache[hostname]; ok {
			nodecache.DeletePod(pod)
		}
	}
}
