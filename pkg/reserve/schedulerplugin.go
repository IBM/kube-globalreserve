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
	"context"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	framework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

// GlobalReservePlugin is an centralized resources approver by hooking for reserve.
type GlobalReservePlugin struct {
	ReserveImpl GlobalReserverInterface
}

// GRConf scheduler plugin configuration
type GRConf struct {
	// if globalreserve works in another pod, this field must be specified
	RemoteURL string `json:"remoteURL,omitempty"`
	//globalreserve http server listening port
	Port int `json:"port,omitempty"`
}

var _ framework.ReservePlugin = &GlobalReservePlugin{}
var _ framework.UnreservePlugin = &GlobalReservePlugin{}

// Name is the name of the plugin used in Registry and configurations.
const Name = "global-resource-reserve-plugin"

// Name returns name of the plugin. It is used in logs, etc.
func (rp *GlobalReservePlugin) Name() string {
	return Name
}

// Reserve is the functions invoked by the framework at "reserve" extension point.
func (rp *GlobalReservePlugin) Reserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	klog.V(3).Infof("Reserve Pod %s on node %s", pod.Name, nodeName)

	reserveResult := rp.ReserveImpl.Reserve(pod, nodeName)
	if len(reserveResult) > 0 {
		return framework.NewStatus(framework.Error, reserveResult)
	}

	return framework.NewStatus(framework.Success, "")
}

// Unreserve is the functions invoked by the framework at "unreserve" extension point.
func (rp *GlobalReservePlugin) Unreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	klog.V(3).Infof("Unreserve Pod %s on node %s", pod.Name, nodeName)

	rp.ReserveImpl.Unreserve(pod, nodeName)
}

// New initializes a new plugin and returns it.
func New(config *runtime.Unknown, handler framework.FrameworkHandle) (framework.Plugin, error) {
	klog.V(3).Infof("global reserve plugin is created")

	conf := &GRConf{}
	if err := framework.DecodeInto(config, conf); err != nil {
		klog.Errorf("Decoding configuration failed with: %s", err.Error())
		return nil, err
	}

	var impl GlobalReserverInterface
	var err error

	if len(conf.RemoteURL) > 0 {
		remote := conf.RemoteURL
		klog.Infof("Remote Global Reserve URL is %s", remote)

		if impl, err = NewHTTPClient(remote, 5); err != nil {
			klog.Errorf("Creating Http client failed with: %s", err.Error())
			return nil, err
		}
	} else {
		port := "23456"

		if conf.Port > 1024 && conf.Port < 65535 {
			port = strconv.Itoa(conf.Port)
		}

		impl, err = NewLocalReserve(handler, port)
		if err != nil {
			klog.Errorf("Creating GlobalReserve failed with: %s", err.Error())
			return nil, err
		}
	}

	return &GlobalReservePlugin{
		ReserveImpl: impl,
	}, nil
}
