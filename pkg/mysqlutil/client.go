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
	// this import needs to be done otherwise the mysql driver don't work
	_ "github.com/go-sql-driver/mysql"
)

// IClient mysql client interface
type IClient interface {
	// Close closes the connection.
	Close() error

	// Exec calls the given query like
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(ctx context.Context, input interface{}, query string, args ...any) error
}

// Client structure representing a client connection to mysql
type Client struct {
	db *sql.DB
}

// NewClient build a client connection and connect to a mysql
func NewClient(addr, username, password string, timeout int) (IClient, error) {
	var err error
	c := &Client{}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/?timeout=%ds&multiStatements=true&interpolateParams=true", username, password, addr, timeout)

	c.db, err = sql.Open("mysql", dsn)
	if err != nil {
		return c, fmt.Errorf("open MySQL connection failed (dsn=%s): %w", dsn, err)

	}

	err = c.db.PingContext(context.TODO())
	if err != nil {
		return c, fmt.Errorf("ping MySQL failed (dsn=%s, timeout=%ds): %w", dsn, timeout, err)
	}

	return c, err
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.db.Close()
}

func (c *Client) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

func (c *Client) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

func (c *Client) QueryRow(ctx context.Context, input interface{}, query string, args ...any) error {
	return c.db.QueryRowContext(ctx, query, args...).Scan(input)
}
