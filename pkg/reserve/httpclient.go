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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
)

// GloalReserveHTTPClient connects with http server created by GloalReserve
type GloalReserveHTTPClient struct {
	mu         sync.RWMutex
	client     *http.Client
	reserveURL string
}

var _ GlobalReserverInterface = &GloalReserveHTTPClient{}

// NewHTTPClient creats an http client
func NewHTTPClient(url string, timeout int) (GlobalReserverInterface, error) {
	transport := utilnet.SetTransportDefaults(&http.Transport{})
	c := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	return &GloalReserveHTTPClient{
		client:     c,
		reserveURL: strings.TrimRight(url, "/"),
	}, nil
}

// Reserve uses http.Client to reserver from remote GloalReserve
func (grhc *GloalReserveHTTPClient) Reserve(pod *v1.Pod, nodeName string) string {
	grhc.mu.Lock()
	defer grhc.mu.Unlock()

	podsArr := []*v1.Pod{pod}
	nodesArr := []string{nodeName}

	reqData := &PodsReserveRequest{
		Pods:          podsArr,
		Nodes:         nodesArr,
		SchedulerName: ReserveSchedulerName,
	}

	return grhc.send(reqData, ReserveHTTPPathPrefix)
}

// Unreserve uses http.Client to unreserver from remote GloalReserve
func (grhc *GloalReserveHTTPClient) Unreserve(pod *v1.Pod, nodeName string) {
	grhc.mu.Lock()
	defer grhc.mu.Unlock()

	podsArr := []*v1.Pod{pod}
	nodesArr := []string{nodeName}

	reqData := &PodsReserveRequest{
		Pods:          podsArr,
		Nodes:         nodesArr,
		SchedulerName: ReserveSchedulerName,
	}

	grhc.send(reqData, UnreserveHTTPPathPrefix)
}

func (grhc *GloalReserveHTTPClient) send(data *PodsReserveRequest, actionPath string) string {
	tmp, err := json.Marshal(data)
	if err != nil {
		return err.Error()
	}

	reqURL := grhc.reserveURL + actionPath

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(tmp))
	if err != nil {
		return err.Error()
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := grhc.client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Failed %s with URL %v, code %v", actionPath, reqURL, resp.StatusCode)
	}

	var result PodReserveResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err.Error()
	}

	return result.Error
}
