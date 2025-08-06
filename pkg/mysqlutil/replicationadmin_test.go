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

package mysqlutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-logr/logr"
)

type result struct {
	insertID     int64
	rowsAffected int64
	err          error
}

var (
	db   *sql.DB
	mock sqlmock.Sqlmock
	addr = "127.0.0.1"
)

func newMockAdmin(db *sql.DB) (IReplicationAdmin, error) {
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		return nil, err
	}

	logger := logr.Discard()
	admin := &ReplicationAdmin{
		cnx: &AdminConnections{
			clients: map[string]IClient{
				addr: &Client{
					db: db,
				},
			},
			connectionTimeout: 0,
			username:          "",
			password:          "",
			log:               logr.Logger{},
		},
		log: logger,
	}
	return admin, nil
}

func (r *result) LastInsertId() (int64, error) {
	return r.insertID, r.err
}

func (r *result) RowsAffected() (int64, error) {
	return r.rowsAffected, r.err
}

func TestSetSemiSyncSourceON(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}

	defer admin.Close()

	mock.ExpectExec(setSemiSyncSourceEnabled).WithArgs(1).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	mock.ExpectExec(setSemiSyncTimeout).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	mock.ExpectExec(setSemiSyncWaitPoint).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.SetSemiSyncSourceON(ctx, addr); err != nil {
		t.Fatal(err)
	}
}

func TestSetSemiSyncSourceOFF(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	mock.ExpectExec(setSemiSyncSourceEnabled).WithArgs(0).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.SetSemiSyncSourceOFF(ctx, addr); err != nil {
		t.Fatal(err)
	}
}

func TestSetSemiSyncReplicaON(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	mock.ExpectExec(setSemiSyncReplicaEnabled).WithArgs(1).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	mock.ExpectExec(setSemiSyncTimeout).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	mock.ExpectExec(setSemiSyncWaitPoint).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.SetSemiSyncReplicaON(ctx, addr); err != nil {
		t.Fatal(err)
	}
}

func TestSetSemiSyncReplicaOFF(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	mock.ExpectExec(setSemiSyncReplicaEnabled).WithArgs(0).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.SetSemiSyncReplicaOFF(ctx, addr); err != nil {
		t.Fatal(err)
	}
}

func TestSetReadOnly(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	mock.ExpectExec(setReadOnlySQL).WithArgs("ON").WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.SetReadOnly(ctx, addr, true); err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec(setReadOnlySQL).WithArgs("OFF").WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.SetReadOnly(ctx, addr, false); err != nil {
		t.Fatal(err)
	}
}

func TestSetSuperReadOnly(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	mock.ExpectExec(setSuperReadOnlySQL).WithArgs("ON").WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.SetSuperReadOnly(ctx, addr, true); err != nil {
		t.Fatal(err)
	}

	mock.ExpectExec(setSuperReadOnlySQL).WithArgs("OFF").WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.SetSuperReadOnly(ctx, addr, false); err != nil {
		t.Fatal(err)
	}
}

func TestResetReplica(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	mock.ExpectExec(stopReplicaSQL).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	mock.ExpectExec(resetReplicaSQL).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.ResetReplica(ctx, addr); err != nil {
		t.Fatal(err)
	}
}

func TestStartReplica(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	mock.ExpectExec(startReplicaSQL).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.StartReplica(ctx, addr); err != nil {
		t.Fatal(err)
	}
}

func TestStopReplica(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	mock.ExpectExec(stopReplicaSQL).WillReturnResult(&result{
		insertID:     0,
		rowsAffected: 0,
		err:          nil,
	})

	if err = admin.StopReplica(ctx, addr); err != nil {
		t.Fatal(err)
	}
}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	mock.ExpectQuery(getUserSql).WithArgs("192.168.1.1").WillReturnRows(
		sqlmock.NewRows([]string{"user", "authentication_string"}).
			AddRow("testuser01", "password").
			AddRow("testuser02", "password"))

	if users, err := admin.GetUser(ctx, addr, "192.168.1.1", []string{}); err != nil {
		t.Fatal(err)
	} else {
		for name, user := range users {
			fmt.Print(name, user)
		}
	}
}

func TestGetReplicationStatus(t *testing.T) {
	ctx := context.Background()
	admin, err := newMockAdmin(db)

	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()

	mock.ExpectQuery(showReplicaStatusSQL).WillReturnRows(sqlmock.NewRows([]string{"Source_Host", "Source_Port", "Replica_IO_Running", "Replica_SQL_Running", "Read_Source_Log_Pos", "Source_Log_File", "Seconds_Behind_Source"}).
		AddRow("10.1.1.1", "3306", "yes", "yes", "node01.log", "node01", "0"))
	mock.ExpectQuery(showReadOnlySQL).WillReturnRows(sqlmock.NewRows([]string{"@@READ_ONLY"}).
		AddRow("1"))
	mock.ExpectQuery(showSuperReadOnlySQL).WillReturnRows(sqlmock.NewRows([]string{"@@SUPER_READ_ONLY"}).
		AddRow("1"))
	mock.ExpectQuery(showSemiSyncSourceEnabled).WillReturnRows(sqlmock.NewRows([]string{"@@rpl_semi_sync_source_enabled"}).
		AddRow("1"))
	mock.ExpectQuery(showSemiSyncReplicaEnabled).WillReturnRows(sqlmock.NewRows([]string{"@@rpl_semi_sync_replica_enabled"}).
		AddRow("1"))

	info := admin.GetReplicationStatus(ctx)
	for a, v := range info.Nodes {
		t.Log(a, v)
	}
}
