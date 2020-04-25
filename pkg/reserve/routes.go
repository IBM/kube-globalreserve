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
	"io"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"k8s.io/klog"
)

func checkBody(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
}

// AddReserveRoute handles reservice requests
func AddReserveRoute(gr *GloalReserve) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		checkBody(w, r)

		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)

		klog.V(3).Infof("Receie reserve request %v", body)

		var request PodsReserveRequest
		var result *PodReserveResult

		if err := json.NewDecoder(body).Decode(&request); err != nil {
			result = &PodReserveResult{
				FailedPods: nil,
				Error:      err.Error(),
			}
		} else {
			if len(request.SchedulerName) <= 0 {
				result = &PodReserveResult{
					FailedPods: nil,
					Error:      "Scheduler name is not specified",
				}
			} else {
				result = gr.ReservePods(request.Pods, request.Nodes)
			}
		}

		if resultBody, err := json.Marshal(result); err != nil {
			klog.Errorf("Failed to marshal reserveResult: %+v, %+v",
				err, result)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			klog.V(3).Infof("reserve successed")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(resultBody)
		}
	}
}

// AddUnreserveRoute handles unreserve requests
func AddUnreserveRoute(gr *GloalReserve) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		checkBody(w, r)

		var buf bytes.Buffer
		body := io.TeeReader(r.Body, &buf)

		klog.V(3).Infof("Receie reserve request %v", body)

		var request PodsReserveRequest
		var result string
		status := http.StatusOK

		if err := json.NewDecoder(body).Decode(&request); err != nil {
			result = err.Error()
		} else {
			if len(request.SchedulerName) <= 0 {
				klog.V(3).Infof("unreserve failed")
				result = "Scheduler name is not specified"
				status = http.StatusBadRequest
			} else {
				gr.UnreservePods(request.Pods, request.Nodes)
				klog.V(3).Infof("unreserve succeed")
				result = ""
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(result))
	}
}
