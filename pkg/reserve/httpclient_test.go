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
)

func TestHttpClient(t *testing.T) {
	node0 := GetNode0()
	nodes := []*v1.Node{node0}
	gr := InitGR(nodes, nil, false)
	ser := InitHTTPServer(gr)

	pod0 := GetPod("pod0", "1000m", "1000", "node0", v1.PodPending)

	ghc, _ := NewHTTPClient("http://127.0.0.1:23456", 5)

	t.Run("HttpClient reserve success", func(t *testing.T) {
		result := ghc.Reserve(pod0, "node0")
		if len(result) > 0 || len(gr.NodeCache["node0"].Pods) != 1 {
			t.Errorf("reserve success")
		}
	})

	t.Run("HttpClient reserve error", func(t *testing.T) {
		pod1 := GetPod("pod1", "2000m", "1000", "node0", v1.PodPending)
		result := ghc.Reserve(pod1, "node0")
		if len(result) == 0 || len(gr.NodeCache["node0"].Pods) != 1 {
			t.Errorf("reserve error failed")
		}
	})

	t.Run("HttpClient unreserve", func(t *testing.T) {
		ghc.Unreserve(pod0, "node0")
		if len(gr.NodeCache["node0"].Pods) != 0 {
			t.Errorf("unreserve failed")
		}
	})

	ser.Close()
}
