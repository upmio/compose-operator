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

package webhook

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/upmio/compose-operator/webhook/mongodbreplicaset"
	"github.com/upmio/compose-operator/webhook/mysqlgroupreplication"
	"github.com/upmio/compose-operator/webhook/mysqlreplication"
	"github.com/upmio/compose-operator/webhook/postgresreplication"
	"github.com/upmio/compose-operator/webhook/proxysqlsync"
	"github.com/upmio/compose-operator/webhook/rediscluster"
	"github.com/upmio/compose-operator/webhook/redisreplication"
)

// Setup creates all controllers with the supplied logger and adds
// them to the supplied manager.
func Setup(mgr ctrl.Manager) error {
	for _, setup := range []func(ctrl.Manager) error{
		mysqlgroupreplication.Setup,
		mysqlreplication.Setup,
		postgresreplication.Setup,
		proxysqlsync.Setup,
		rediscluster.Setup,
		redisreplication.Setup,
		mongodbreplicaset.Setup,
	} {
		if err := setup(mgr); err != nil {
			return err
		}
	}
	return nil
}
