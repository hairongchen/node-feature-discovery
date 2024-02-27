/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nfdmaster

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	fakek8sclient "k8s.io/client-go/kubernetes/fake"
	fakenfdclient "sigs.k8s.io/node-feature-discovery/api/generated/clientset/versioned/fake"
)

func newFakeNodeUpdaterPool(nfdMaster *nfdMaster) *nodeUpdaterPool {
	return &nodeUpdaterPool{
		nfdMaster: nfdMaster,
		wg:        sync.WaitGroup{},
	}
}

func TestNodeUpdaterStart(t *testing.T) {
	fakeMaster := newFakeMaster(nil)
	nodeUpdaterPool := newFakeNodeUpdaterPool(fakeMaster)

	Convey("When starting the node updater pool", t, func() {
		nodeUpdaterPool.start(10)
		q := nodeUpdaterPool.queue
		Convey("Node updater pool queue properties should change", func() {
			So(q, ShouldNotBeNil)
			So(q.ShuttingDown(), ShouldBeFalse)
		})

		nodeUpdaterPool.start(10)
		Convey("Node updater pool queue should not change", func() {
			So(nodeUpdaterPool.queue, ShouldEqual, q)
		})
	})
}

func TestNodeUpdaterStop(t *testing.T) {
	fakeMaster := newFakeMaster(nil)
	nodeUpdaterPool := newFakeNodeUpdaterPool(fakeMaster)

	nodeUpdaterPool.start(10)

	Convey("When stoping the node updater pool", t, func() {
		nodeUpdaterPool.stop()
		Convey("Node updater pool queue should be removed", func() {
			// Wait for the wg.Done()
			So(func() interface{} {
				return nodeUpdaterPool.queue.ShuttingDown()
			}, withTimeout, 2*time.Second, ShouldBeTrue)
		})
	})
}

func TestRunNodeUpdater(t *testing.T) {
	fakeMaster := newFakeMaster(fakek8sclient.NewSimpleClientset())
	fakeMaster.nfdController = newFakeNfdAPIController(fakenfdclient.NewSimpleClientset())
	nodeUpdaterPool := newFakeNodeUpdaterPool(fakeMaster)

	nodeUpdaterPool.start(10)
	Convey("Queue has no element", t, func() {
		So(nodeUpdaterPool.queue.Len(), ShouldEqual, 0)
	})
	nodeUpdaterPool.queue.Add(testNodeName)
	Convey("Added element to the queue should be removed", t, func() {
		So(func() interface{} { return nodeUpdaterPool.queue.Len() },
			withTimeout, 2*time.Second, ShouldEqual, 0)
	})
}
