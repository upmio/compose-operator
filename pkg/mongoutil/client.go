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
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IMongoClient interface for MongoDB client operations
type IMongoClient interface {

	// Disconnect closes the connection to MongoDB
	Disconnect(ctx context.Context) error

	// GetReplicaSetStatus returns the replica set status
	GetReplicaSetStatus(ctx context.Context) (*ReplicaSetStatus, error)

	// GetReplicaSetConfig returns the replica set configuration
	GetReplicaSetConfig(ctx context.Context) (*ReplicaSetConfig, error)

	// InitiateReplicaSet initializes a new replica set
	InitiateReplicaSet(ctx context.Context, config *ReplicaSetConfig) error

	// AddMemberToReplicaSet adds a new member to the replica set
	AddMemberToReplicaSet(ctx context.Context, host string, port int) error

	// RemoveMemberFromReplicaSet removes a member from the replica set
	RemoveMemberFromReplicaSet(ctx context.Context, host string, port int) error

	// StepDown forces the primary to step down
	StepDown(ctx context.Context, stepDownSecs int) error

	// IsMaster checks if the current node is the primary
	IsMaster(ctx context.Context) (*IsMasterResult, error)
}

// MongoClient wraps the MongoDB client
type MongoClient struct {
	client *mongo.Client
}

// IsMasterResult represents the result of isMaster command
type IsMasterResult struct {
	IsMaster                     bool      `bson:"ismaster"`
	IsSecondary                  bool      `bson:"secondary"`
	MaxBsonObjectSize            int       `bson:"maxBsonObjectSize"`
	MaxMessageSizeBytes          int       `bson:"maxMessageSizeBytes"`
	MaxWriteBatchSize            int       `bson:"maxWriteBatchSize"`
	LocalTime                    time.Time `bson:"localTime"`
	LogicalSessionTimeoutMinutes int       `bson:"logicalSessionTimeoutMinutes"`
	ConnectionId                 int       `bson:"connectionId"`
	MinWireVersion               int       `bson:"minWireVersion"`
	MaxWireVersion               int       `bson:"maxWireVersion"`
	ReadOnly                     bool      `bson:"readOnly"`
	OK                           float64   `bson:"ok"`
	SetName                      string    `bson:"setName,omitempty"`
	SetVersion                   int       `bson:"setVersion,omitempty"`
	Hosts                        []string  `bson:"hosts,omitempty"`
	Primary                      string    `bson:"primary,omitempty"`
	Me                           string    `bson:"me,omitempty"`
}

// NewMongoClient creates a new MongoDB client
func NewMongoClient(address string, opts *AdminOptions) (IMongoClient, error) {
	var err error
	c := &MongoClient{}

	uri := fmt.Sprintf("mongodb://%s", address)

	clientOptions := options.Client().ApplyURI(uri)

	if opts.ConnectionTimeout > 0 {
		clientOptions = clientOptions.SetConnectTimeout(opts.ConnectionTimeout)
		clientOptions = clientOptions.SetServerSelectionTimeout(opts.ConnectionTimeout)
	}

	// Enable direct connection to allow connecting to uninitialized replica set nodes
	clientOptions = clientOptions.SetDirect(true)

	if opts.Username != "" && opts.Password != "" {
		authDB := opts.AuthDatabase
		if authDB == "" {
			authDB = "admin"
		}

		credential := options.Credential{
			Username:   opts.Username,
			Password:   opts.Password,
			AuthSource: authDB,
		}
		clientOptions = clientOptions.SetAuth(credential)
	}

	ctx := context.Background()
	c.client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return c, fmt.Errorf("open MongoDB connection failed (uri=%s): %w", uri, err)
	}

	if err := c.client.Ping(ctx, nil); err != nil {
		return c, fmt.Errorf("ping MongoDB connection failed (uri=%s): %w", uri, err)
	}

	return c, err
}

// Disconnect closes the connection to MongoDB
func (c *MongoClient) Disconnect(ctx context.Context) error {
	if c.client != nil {
		return c.client.Disconnect(ctx)
	}
	return nil
}

// GetReplicaSetStatus returns the replica set status
func (c *MongoClient) GetReplicaSetStatus(ctx context.Context) (*ReplicaSetStatus, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not connected")
	}

	var result ReplicaSetStatus
	err := c.client.Database("admin").RunCommand(ctx, bson.D{{Key: "replSetGetStatus", Value: 1}}).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to get replica set status: %v", err)
	}

	return &result, nil
}

// GetReplicaSetConfig returns the replica set configuration
func (c *MongoClient) GetReplicaSetConfig(ctx context.Context) (*ReplicaSetConfig, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not connected")
	}

	var result struct {
		Config *ReplicaSetConfig `bson:"config"`
	}

	err := c.client.Database("admin").RunCommand(ctx, bson.D{{Key: "replSetGetConfig", Value: 1}}).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to get replica set config: %v", err)
	}

	return result.Config, nil
}

// InitiateReplicaSet initializes a new replica set
func (c *MongoClient) InitiateReplicaSet(ctx context.Context, config *ReplicaSetConfig) error {
	if c.client == nil {
		return fmt.Errorf("client not connected")
	}

	cmd := bson.D{{Key: "replSetInitiate", Value: config}}

	var result bson.M
	err := c.client.Database("admin").RunCommand(ctx, cmd).Decode(&result)
	if err != nil {
		return fmt.Errorf("failed to initiate replica set: %v", err)
	}

	return nil
}

// reconfigureReplicaSet reconfigures the replica set
func (c *MongoClient) reconfigureReplicaSet(ctx context.Context, config *ReplicaSetConfig) error {
	if c.client == nil {
		return fmt.Errorf("client not connected")
	}

	cmd := bson.D{{Key: "replSetReconfig", Value: config}}

	var result bson.M
	err := c.client.Database("admin").RunCommand(ctx, cmd).Decode(&result)
	if err != nil {
		return fmt.Errorf("failed to reconfigure replica set: %v", err)
	}

	return nil
}

// AddMemberToReplicaSet adds a new member to the replica set
func (c *MongoClient) AddMemberToReplicaSet(ctx context.Context, host string, port int) error {
	// Get current config
	config, err := c.GetReplicaSetConfig(ctx)
	if err != nil {
		return err
	}

	// Find next available member ID
	maxID := -1
	hostPort := fmt.Sprintf("%s:%d", host, port)

	// Check if member already exists
	for _, member := range config.Members {
		if member.Host == hostPort {
			return fmt.Errorf("member %s already exists in replica set", hostPort)
		}
		if member.ID > maxID {
			maxID = member.ID
		}
	}

	// Add new member
	newMember := ReplicaSetConfigMember{
		ID:       maxID + 1,
		Host:     hostPort,
		Priority: 1.0,
		Votes:    1,
	}

	config.Members = append(config.Members, newMember)
	config.Version++

	return c.reconfigureReplicaSet(ctx, config)
}

// RemoveMemberFromReplicaSet removes a member from the replica set
func (c *MongoClient) RemoveMemberFromReplicaSet(ctx context.Context, host string, port int) error {
	// Get current config
	config, err := c.GetReplicaSetConfig(ctx)
	if err != nil {
		return err
	}

	hostPort := fmt.Sprintf("%s:%d", host, port)
	var newMembers []ReplicaSetConfigMember

	found := false
	for _, member := range config.Members {
		if member.Host != hostPort {
			newMembers = append(newMembers, member)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("member %s not found in replica set", hostPort)
	}

	config.Members = newMembers
	config.Version++

	return c.reconfigureReplicaSet(ctx, config)
}

// StepDown forces the primary to step down
func (c *MongoClient) StepDown(ctx context.Context, stepDownSecs int) error {
	if c.client == nil {
		return fmt.Errorf("client not connected")
	}

	cmd := bson.D{{Key: "replSetStepDown", Value: stepDownSecs}}

	var result bson.M
	err := c.client.Database("admin").RunCommand(ctx, cmd).Decode(&result)
	if err != nil {
		return fmt.Errorf("failed to step down: %v", err)
	}

	return nil
}

// IsMaster checks if the current node is the primary
func (c *MongoClient) IsMaster(ctx context.Context) (*IsMasterResult, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not connected")
	}

	var result IsMasterResult
	err := c.client.Database("admin").RunCommand(ctx, bson.D{{Key: "isMaster", Value: 1}}).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to check isMaster: %v", err)
	}

	return &result, nil
}
