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
	"fmt"

	"github.com/go-logr/logr"
)

const (
	defaultClientTimeout = 2

	// ErrNotFound cannot find a node to connect to
	ErrNotFound = "unable to find a node to connect"
)

// IAdminConnections interface representing the map of admin connections to mysql nodes
type IAdminConnections interface {
	// AddAll connect to the given list of addresses and
	// register them in the map
	// fail silently
	AddAll(addrs []string)

	// Add connect to the given address and
	// register the client connection to the pool
	Add(addr string) error

	// Reconnect force a reconnection on the given address
	// if the adress is not part of the map, act like Add
	Reconnect(addr string) error

	// Remove disconnect and remove the client connection from the map
	Remove(addr string)

	// ReplaceAll clear the map and re-populate it with new connections
	// fail silently
	ReplaceAll(addrs []string)

	// Reset close all connections and clear the connection map
	Reset()

	// Get returns a client connection for the given address,
	// connects if the connection is not in the map yet
	Get(addr string) (IClient, error)

	// GetAll returns a map of all clients per address
	GetAll() map[string]IClient

	//GetSelected returns a map of clients based on the input addresses
	GetSelected(addrs []string) map[string]IClient
}

// AdminConnections connection map for mysql replication cluster
// currently the admin connection is not threadSafe since it is only use in the Events thread.
type AdminConnections struct {
	clients           map[string]IClient
	connectionTimeout int

	username string
	password string
	log      logr.Logger
}

// NewAdminConnections returns and instance of AdminConnectionsInterface
func NewAdminConnections(addrs []string, options *AdminOptions, log logr.Logger) IAdminConnections {
	cnx := &AdminConnections{
		clients:           make(map[string]IClient),
		connectionTimeout: defaultClientTimeout,
		log:               log,
	}
	if options != nil {
		if options.ConnectionTimeout != 0 {
			cnx.connectionTimeout = options.ConnectionTimeout
		}

		cnx.username = options.Username
		cnx.password = options.Password
	}
	cnx.AddAll(addrs)
	return cnx
}

// AddAll connect to the given list of addresses and
// register them in the map
// fail silently
func (cnx *AdminConnections) AddAll(addrs []string) {
	for _, addr := range addrs {
		_ = cnx.Add(addr)
	}
}

// Add connect to the given address and
// register the client connection to the map
func (cnx *AdminConnections) Add(addr string) error {
	_, err := cnx.update(addr)
	return err
}

// Update returns a client connection for the given adress,
// connects if the connection is not in the map yet
func (cnx *AdminConnections) update(addr string) (IClient, error) {
	// if already exist close the current connection
	if c, ok := cnx.clients[addr]; ok {
		_ = c.Close()
	}

	c, err := cnx.connect(addr)
	if err == nil && c != nil {
		cnx.clients[addr] = c
	} else {
		cnx.log.Info(fmt.Sprintf("cannot connect to %s ", addr))
	}
	return c, err
}

func (cnx *AdminConnections) connect(addr string) (IClient, error) {
	c, err := NewClient(addr, cnx.username, cnx.password, cnx.connectionTimeout)

	if err != nil {
		return nil, err
	}

	return c, nil
}

// Reconnect force a reconnection on the given address
// is the adress is not part of the map, act like Add
func (cnx *AdminConnections) Reconnect(addr string) error {
	cnx.log.Info(fmt.Sprintf("reconnecting to %s", addr))
	cnx.Remove(addr)
	return cnx.Add(addr)
}

// Remove disconnect and remove the client connection from the map
func (cnx *AdminConnections) Remove(addr string) {
	if c, ok := cnx.clients[addr]; ok {
		_ = c.Close()
		delete(cnx.clients, addr)
	}
}

// ReplaceAll clear the pool and re-populate it with new connections
// fail silently
func (cnx *AdminConnections) ReplaceAll(addrs []string) {
	cnx.Reset()
	cnx.AddAll(addrs)
}

// Reset close all connections and clear the connection map
func (cnx *AdminConnections) Reset() {
	for _, c := range cnx.clients {
		_ = c.Close()
	}
	cnx.clients = map[string]IClient{}
}

// Get returns a client connection for the given adress,
// connects if the connection is not in the map yet
func (cnx *AdminConnections) Get(addr string) (IClient, error) {
	if c, ok := cnx.clients[addr]; ok {
		return c, nil
	}
	c, err := cnx.connect(addr)
	if err == nil && c != nil {
		cnx.clients[addr] = c
	}
	return c, err
}

// GetAll returns a map of all clients per address
func (cnx *AdminConnections) GetAll() map[string]IClient {
	return cnx.clients
}

// GetSelected returns a map of clients based on the input addresses
func (cnx *AdminConnections) GetSelected(addrs []string) map[string]IClient {
	clientsSelected := make(map[string]IClient)
	for _, addr := range addrs {
		if client, ok := cnx.clients[addr]; ok {
			clientsSelected[addr] = client
		}
	}
	return clientsSelected
}
