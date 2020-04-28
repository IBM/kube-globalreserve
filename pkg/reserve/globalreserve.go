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
	"log"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	schedulerlisters "k8s.io/kubernetes/pkg/scheduler/listers"
)

//GloalReserve used for saving all infomation
type GloalReserve struct {
	mu             sync.RWMutex
	ResTypeToID    map[v1.ResourceName]int         //key: resource type, value: index
	ResTypeMaxKind int                             //the max number of resource types
	NextResourceID int                             //index when new resource type is added
	NodeCache      map[string]*NodeResInfo         //global cache, key: node name, value: NodeResInfo
	PodToNode      map[types.UID]string            //key: pod uid, value: binding host
	NodeLister     schedulerlisters.NodeInfoLister //listing all pods when starting
	PodLister      schedulerlisters.PodLister      //listing all nodes when starting
}

var _ GlobalReserverInterface = &GloalReserve{}

//NewLocalReserve return a GloalReserve with can works with k8s default scheduler plugin
func NewLocalReserve(handler framework.FrameworkHandle, listeningPort string) (GlobalReserverInterface, error) {
	gr := &GloalReserve{
		ResTypeToID:    make(map[v1.ResourceName]int),
		ResTypeMaxKind: 0,
		NextResourceID: 0,
		NodeCache:      make(map[string]*NodeResInfo),
		PodToNode:      make(map[types.UID]string),
		NodeLister:     handler.SnapshotSharedLister().NodeInfos(),
		PodLister:      handler.SnapshotSharedLister().Pods(),
	}

	// add node event handlers, add/delete node into/from cache
	handler.SharedInformerFactory().Core().V1().Nodes().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    gr.AddNode,
			UpdateFunc: gr.UpdateNode,
			DeleteFunc: gr.DeleteNode,
		},
	)

	// add pod event handlers, add/delete pod into/from cache
	handler.SharedInformerFactory().Core().V1().Pods().Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			// does not handle non-binding pod
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					return assignedPod(t)
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*v1.Pod); ok {
						return assignedPod(pod)
					}
					klog.Errorf("unable to convert object %T to *v1.Pod in %T", obj, gr)
					return false
				default:
					klog.Errorf("unable to handle object %T to *v1.Pod in %T", obj, gr)
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    gr.AddPod,
				UpdateFunc: gr.UpdatePod,
				DeleteFunc: gr.DeletePod,
			},
		},
	)

	// start the http server to receive 3rd party reserve request
	router := httprouter.New()
	router.POST(ReserveHTTPPathPrefix, AddReserveRoute(gr))
	router.POST(UnreserveHTTPPathPrefix, AddUnreserveRoute(gr))

	klog.V(3).Infof("GloalReserve is listening %s", listeningPort)

	go func() {
		log.Println(http.ListenAndServe(":"+listeningPort, router))
	}()

	return gr, nil
}

// CollectFromLister collects all nodes and pods in the cluster, and saves in the cache
func (gr *GloalReserve) CollectFromLister() error {
	nodes, err := gr.NodeLister.List()
	if err != nil {
		klog.Error("Can not get node lists from Informer.")
		return err
	}

	/* collect all resource types and create a map key is resource type value is integer
	   res1 -> 0
	   res2 -> 1
	   ...
	   res7 -> 6
	*/
	for _, nodeInfo := range nodes {
		node := nodeInfo.Node()
		for resName := range node.Status.Allocatable {
			if _, ok := gr.ResTypeToID[resName]; !ok {
				gr.ResTypeToID[resName] = gr.NextResourceID
				gr.NextResourceID++
			}
		}
	}

	gr.ResTypeMaxKind = gr.NextResourceID + ResourceTypeBuffer

	// collect all nodes
	for _, nodeInfo := range nodes {
		node := nodeInfo.Node()
		gr.NodeCache[node.Name] = NewNodeResInfo(node, gr.ResTypeToID, gr.ResTypeMaxKind)
	}
	// list all pods
	pods, err := gr.PodLister.List(labels.Everything())
	if err != nil {
		klog.Error("Can not get pod lists from Informer.")
		return err
	}

	// adding all pods into NodeCache
	for _, pod := range pods {
		hostname := pod.Spec.NodeName
		if len(hostname) > 0 {
			if podsOnHost, ok := gr.NodeCache[hostname]; ok {
				podsOnHost.AddPodToCache(pod, gr.ResTypeToID)
				gr.PodToNode[pod.UID] = hostname
			} else {
				klog.Warningf("Pod %s is binded on %s, but the node does not exist in Node list", pod.Name, hostname)
			}
		} else {
			klog.V(3).Infof("Pod %s is not binded, ignore it.", pod.Name)
		}
	}

	gr.Dump()

	return nil
}

// Dump for debugging
func (gr *GloalReserve) Dump() {
	if klog.V(3) {
		klog.Infoln("Dump resources with index:")
		for name, index := range gr.ResTypeToID {
			klog.Infof("     %d : %s", index, name)
		}

		klog.Infoln("Dump all collected nodes and pods")
		for _, nodeI := range gr.NodeCache {
			nodeI.Dump()
		}
	}
}

// Reserve pod resource from the specified nodename
func (gr *GloalReserve) Reserve(pod *v1.Pod, nodeName string) string {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	retStr := ""
	// nothing in the cache, collect all pods and nodes
	if gr.NextResourceID == 0 {
		gr.CollectFromLister()
	}

	if nodeInfo, ok := gr.NodeCache[nodeName]; ok {
		checkResult := nodeInfo.CheckPod(pod, gr.ResTypeToID)
		if checkResult {
			nodeInfo.AddPodToCache(pod, gr.ResTypeToID)
			gr.PodToNode[pod.UID] = nodeName
		} else {
			retStr = "Resource is not enough."
		}
	} else {
		// the host does not exist
		retStr = "NodeName does not exist"
	}

	return retStr
}

// Unreserve pod resources from the specified nodename
func (gr *GloalReserve) Unreserve(pod *v1.Pod, nodeName string) {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	delete(gr.PodToNode, pod.UID)
	if nodeInfo, ok := gr.NodeCache[nodeName]; ok {
		nodeInfo.DeletePod(pod)
	} else {
		// the host does not exist
		klog.V(3).Infof("NodeName does not exist")
	}
}

// ReservePods works for pods
func (gr *GloalReserve) ReservePods(pods []*v1.Pod, nodeNames []string) *PodReserveResult {
	// group pods by hostname
	hostToAvailabel := make(map[string][]int64)

	failed := make([]string, 0, len(pods))
	var errorReason string

	gr.mu.Lock()
	defer gr.mu.Unlock()

	// nothing in the cache, collect all pods and nodes
	if gr.NextResourceID == 0 {
		gr.CollectFromLister()
	}

	for i, p := range pods {
		nodeName := nodeNames[i]
		if nodeInfo, ok := gr.NodeCache[nodeName]; !ok {
			failed = append(failed, p.Name)
			errorReason = "Node does not exist"
		} else {
			if _, ok1 := hostToAvailabel[nodeName]; !ok1 {
				hostToAvailabel[nodeName] = nodeInfo.GetAvailable()
			}
		}
	}

	if len(errorReason) > 0 {
		return &PodReserveResult{
			FailedPods: failed,
			Error:      errorReason,
		}
	}

	// check pods one by one
	for i, p := range pods {
		nodeName := nodeNames[i]
		podReq := make([]int64, gr.ResTypeMaxKind)
		GetPodReq(p, gr.ResTypeToID, podReq)
		if VectorCompare(hostToAvailabel[nodeName], podReq) {
			VectorMinus(hostToAvailabel[nodeName], podReq)
		} else {
			failed = append(failed, p.Name)
			errorReason = "Node does not have enough resource"
		}
	}

	if len(errorReason) > 0 {
		return &PodReserveResult{
			FailedPods: failed,
			Error:      errorReason,
		}
	}

	for i, p := range pods {
		nodeName := nodeNames[i]
		gr.NodeCache[nodeName].AddPodToCache(p, gr.ResTypeToID)
		gr.PodToNode[p.UID] = nodeName
	}

	// add pods by hostname
	return &PodReserveResult{
		FailedPods: failed,
		Error:      "",
	}
}

//UnreservePods works for pods
func (gr *GloalReserve) UnreservePods(pods []*v1.Pod, nodeNames []string) {
	for i := range pods {
		gr.Unreserve(pods[i], nodeNames[i])
	}
}
