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

package testutil

import (
	"context"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

func CreateMysqlContainer(ctx context.Context, username, password string) (*mysql.MySQLContainer, error) {
	return mysql.Run(ctx,
		"mysql:8.0.36",
		mysql.WithConfigFile(filepath.Join("./", "testdata", "my.cnf")),
		mysql.WithUsername(username),
		mysql.WithPassword(password),
		mysql.WithScripts(filepath.Join("./", "testdata", "init.sql")),
	)
}
