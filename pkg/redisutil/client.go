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
	"strings"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
)

// IClient redis client interface
type IClient interface {
	// Close closes the connection.
	Close() error

	// Cmd calls the given Redis command.
	Cmd(cmd string, args ...interface{}) *redis.Resp

	// PipeAppend adds the given call to the pipeline queue.
	// Use PipeResp() to read the response.
	PipeAppend(cmd string, args ...interface{})

	// PipeResp returns the reply for the next request in the pipeline queue. Err
	// with ErrPipelineEmpty is returned if the pipeline queue is empty.
	PipeResp() *redis.Resp

	// PipeClear clears the contents of the current pipeline queue, both commands
	// queued by PipeAppend which have yet to be sent and responses which have yet
	// to be retrieved through PipeResp. The first returned int will be the number
	// of pending commands dropped, the second will be the number of pending
	// responses dropped
	PipeClear() (int, int)
}

// Client structure representing a client connection to redis
type Client struct {
	client *redis.Client
}

// NewClient build a client connection and connect to a redis address
func NewClient(addr, password string, cnxTimeout time.Duration) (IClient, error) {
	var err error
	c := &Client{}

	c.client, err = redis.DialTimeout("tcp", addr, cnxTimeout)
	if err != nil {
		return c, err
	}
	if password != "" {
		err = c.client.Cmd("AUTH", password).Err
	}
	return c, err
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// Cmd calls the given Redis command.
func (c *Client) Cmd(cmd string, args ...interface{}) *redis.Resp {
	return c.client.Cmd(c.getCommand(cmd), args)
}

// getCommand return the command name after applying rename-command
func (c *Client) getCommand(cmd string) string {
	upperCmd := strings.ToUpper(cmd)

	return upperCmd
}

// PipeAppend adds the given call to the pipeline queue.
func (c *Client) PipeAppend(cmd string, args ...interface{}) {
	c.client.PipeAppend(c.getCommand(cmd), args)
}

// PipeResp returns the reply for the next request in the pipeline queue. Err
func (c *Client) PipeResp() *redis.Resp {
	return c.client.PipeResp()
}

// PipeClear clears the contents of the current pipeline queue
func (c *Client) PipeClear() (int, int) {
	return c.client.PipeClear()
}
