/*
Copyright 2025 The Compose Operator Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package redisutil

import (
	"fmt"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestGetClusterInfos(t *testing.T) {
	clusterAddrs := map[string][]string{
		"shard1": {
			"192.168.26.10:16379",
			"192.168.26.10:16380",
		},
		"shard2": {
			"192.168.26.10:16381",
			"192.168.26.10:16382",
		},
		"shard3": {
			"192.168.26.10:16383",
			"192.168.26.10:16384",
		},
	}

	//for _, addr := range nodesAddrs {
	//	client, err := redis.DialTimeout("tcp", addr, 5*time.Second)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	resp := client.Cmd("CLUSTER", "NODES")
	//	t.Log(resp.String())
	//	resp = client.Cmd("CLUSTER", "INFO")
	//	t.Log(resp.String())
	//}

	//cluster := NewCluster("test", "test")
	nodesAddrs := make([]string, 0)
	for _, addrs := range clusterAddrs {
		for _, addr := range addrs {
			nodesAddrs = append(nodesAddrs, addr)
		}
	}

	opts := AdminOptions{
		ConnectionTimeout: 5 * time.Second,
		Password:          "",
	}

	zapopts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zapopts)))

	reqLogger := ctrl.Log.WithName("redis_utils").WithValues("Request.Namespace", "aaa", "Request.Name", "bbb")
	admin := NewClusterAdmin(nodesAddrs, &opts, reqLogger)

	defer admin.Close()

	infos, err := admin.GetClusterInfos()
	if err != nil {
		fmt.Println(infos.Status)

		for addr, nodeinfos := range infos.Infos {
			fmt.Println("==============================")
			fmt.Printf("address: %s\n", addr)
			fmt.Printf("nodeinfo: %#v\n", nodeinfos.Node)

			fmt.Printf("nodeinfo friends %#v\n", nodeinfos.Friends)
		}

		var firstNode *ClusterNode

		if infos.Status == ClusterInfosInconsistent {
			for _, nodeInfo := range infos.Infos {
				firstNode = nodeInfo.Node
				break
			}
			err := admin.AttachNodeToCluster(firstNode.IPPort())
			if err != nil {
				t.Errorf(err.Error())
			}
			//
			//for _, addrs := range cluster {
			//	for _, addr := range addrs {
			//
			//	}
			//	AttachingSlavesToMaster(admin)
			//}

			//if err := clusterCtx.AttachingSlavesToMaster(admin); err != nil {
			//	return Cluster.Wrap(err, "AttachingSlavesToMaster")
			//}
			//
			//if err := clusterCtx.AllocSlots(admin, newMasters); err != nil {
			//	return Cluster.Wrap(err, "AllocSlots")
		}
	} else {
		fmt.Println(infos.Status)

		fmt.Printf("%v\n", infos)
		for addr, nodeinfos := range infos.Infos {
			fmt.Println("##############################")

			fmt.Printf("address: %s\n", addr)
			fmt.Printf("nodeinfo: %#v\n", nodeinfos.Node)

			for _, frientdNodeInfo := range nodeinfos.Friends {
				fmt.Printf("nodeinfo friends %#v\n", frientdNodeInfo)
			}
		}

	}
}
