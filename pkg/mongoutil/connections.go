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

package mongoutil

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"
)

// IAdminConnections interface for managing multiple MongoDB connections
type IAdminConnections interface {
	// Get returns a client for the given address
	Get(addr string) (IMongoClient, error)

	// GetAll returns all clients
	GetAll() map[string]IMongoClient

	// Add adds a new client for the given address
	Add(addr string) error

	// Remove removes the client for the given address
	Remove(addr string) error

	// Reset closes and removes all connections
	Reset()
}

// AdminConnections manages multiple MongoDB connections
type AdminConnections struct {
	clients map[string]IMongoClient
	options *AdminOptions
	mutex   sync.RWMutex
	log     logr.Logger
}

// NewAdminConnections creates a new AdminConnections instance
func NewAdminConnections(addrs []string, options *AdminOptions, log logr.Logger) IAdminConnections {
	cnx := &AdminConnections{
		clients: make(map[string]IMongoClient),
		options: options,
		log:     log.WithName("mongo_connections"),
	}

	for _, addr := range addrs {
		if err := cnx.Add(addr); err != nil {
			log.Error(err, "failed to add MongoDB connection", "address", addr)
		}
	}

	return cnx
}

// Get returns a client for the given address
func (c *AdminConnections) Get(addr string) (IMongoClient, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	client, exists := c.clients[addr]
	if !exists {
		return nil, fmt.Errorf("no client found for address: %s", addr)
	}

	return client, nil
}

// GetAll returns all clients
func (c *AdminConnections) GetAll() map[string]IMongoClient {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	result := make(map[string]IMongoClient)
	for addr, client := range c.clients {
		result[addr] = client
	}

	return result
}

// Add adds a new client for the given address
func (c *AdminConnections) Add(addr string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.clients[addr]; exists {
		c.log.Info("client already exists for address", "address", addr)
		return nil
	}

	client, err := NewMongoClient(addr, c.options)
	if err != nil {
		return err
	}
	c.clients[addr] = client

	c.log.V(4).Info("added MongoDB client", "address", addr)
	return nil
}

// Remove removes the client for the given address
func (c *AdminConnections) Remove(addr string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	client, exists := c.clients[addr]
	if !exists {
		return fmt.Errorf("no client found for address: %s", addr)
	}

	// Disconnect the client
	ctx := context.Background()
	if err := client.Disconnect(ctx); err != nil {
		c.log.Error(err, "failed to disconnect MongoDB client", "address", addr)
	}

	delete(c.clients, addr)
	c.log.V(4).Info("removed MongoDB client", "address", addr)
	return nil
}

// Reset closes and removes all connections
func (c *AdminConnections) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctx := context.Background()
	for addr, client := range c.clients {
		if err := client.Disconnect(ctx); err != nil {
			c.log.Error(err, "failed to disconnect MongoDB client during reset", "address", addr)
		}
	}

	c.clients = make(map[string]IMongoClient)
	c.log.Info("reset all MongoDB connections")
}
