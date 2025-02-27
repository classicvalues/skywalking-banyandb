// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	commonv1 "github.com/apache/skywalking-banyandb/api/proto/banyandb/common/v1"
	databasev1 "github.com/apache/skywalking-banyandb/api/proto/banyandb/database/v1"
	"github.com/apache/skywalking-banyandb/banyand/metadata/schema"
)

func TestStreamRegistry(t *testing.T) {
	req := require.New(t)
	gracefulStop := setup(req, testData{
		TLS:  false,
		addr: "localhost:17912",
	})
	defer gracefulStop()

	conn, err := grpc.Dial("localhost:17912", grpc.WithInsecure())
	req.NoError(err)
	req.NotNil(conn)

	client := databasev1.NewStreamRegistryServiceClient(conn)
	req.NotNil(client)

	meta := &commonv1.Metadata{
		Group: "default",
		Name:  "sw",
	}

	getResp, err := client.Get(context.TODO(), &databasev1.StreamRegistryServiceGetRequest{Metadata: meta})

	req.NoError(err)
	req.NotNil(getResp)

	// 2 - DELETE
	deleteResp, err := client.Delete(context.TODO(), &databasev1.StreamRegistryServiceDeleteRequest{
		Metadata: meta,
	})
	req.NoError(err)
	req.NotNil(deleteResp)
	req.True(deleteResp.GetDeleted())

	// 3 - GET -> Nil
	_, err = client.Get(context.TODO(), &databasev1.StreamRegistryServiceGetRequest{
		Metadata: meta,
	})
	errStatus, _ := status.FromError(err)
	req.Equal(errStatus.Message(), schema.ErrEntityNotFound.Error())

	// 4 - CREATE
	_, err = client.Create(context.TODO(), &databasev1.StreamRegistryServiceCreateRequest{Stream: getResp.GetStream()})
	req.NoError(err)

	// 5 - GET - > Not Nil
	getResp, err = client.Get(context.TODO(), &databasev1.StreamRegistryServiceGetRequest{
		Metadata: meta,
	})
	req.NoError(err)
	req.NotNil(getResp)
}

func TestIndexRuleBindingRegistry(t *testing.T) {
	req := require.New(t)
	gracefulStop := setup(req, testData{
		TLS:  false,
		addr: "localhost:17912",
	})
	defer gracefulStop()

	conn, err := grpc.Dial("localhost:17912", grpc.WithInsecure())
	req.NoError(err)
	req.NotNil(conn)

	client := databasev1.NewIndexRuleBindingRegistryServiceClient(conn)
	req.NotNil(client)

	meta := &commonv1.Metadata{
		Group: "default",
		Name:  "sw-index-rule-binding",
	}

	getResp, err := client.Get(context.TODO(), &databasev1.IndexRuleBindingRegistryServiceGetRequest{Metadata: meta})

	req.NoError(err)
	req.NotNil(getResp)

	// 2 - DELETE
	deleteResp, err := client.Delete(context.TODO(), &databasev1.IndexRuleBindingRegistryServiceDeleteRequest{
		Metadata: meta,
	})
	req.NoError(err)
	req.NotNil(deleteResp)
	req.True(deleteResp.GetDeleted())

	// 3 - GET -> Nil
	_, err = client.Get(context.TODO(), &databasev1.IndexRuleBindingRegistryServiceGetRequest{
		Metadata: meta,
	})
	errStatus, _ := status.FromError(err)
	req.Equal(errStatus.Message(), schema.ErrEntityNotFound.Error())

	// 4 - CREATE
	_, err = client.Create(context.TODO(), &databasev1.IndexRuleBindingRegistryServiceCreateRequest{IndexRuleBinding: getResp.GetIndexRuleBinding()})
	req.NoError(err)

	// 5 - GET - > Not Nil
	getResp, err = client.Get(context.TODO(), &databasev1.IndexRuleBindingRegistryServiceGetRequest{
		Metadata: meta,
	})
	req.NoError(err)
	req.NotNil(getResp)
}

func TestIndexRuleRegistry(t *testing.T) {
	req := require.New(t)
	gracefulStop := setup(req, testData{
		TLS:  false,
		addr: "localhost:17912",
	})
	defer gracefulStop()

	conn, err := grpc.Dial("localhost:17912", grpc.WithInsecure())
	req.NoError(err)
	req.NotNil(conn)

	client := databasev1.NewIndexRuleRegistryServiceClient(conn)
	req.NotNil(client)

	meta := &commonv1.Metadata{
		Group: "default",
		Name:  "db.instance",
	}

	getResp, err := client.Get(context.TODO(), &databasev1.IndexRuleRegistryServiceGetRequest{Metadata: meta})

	req.NoError(err)
	req.NotNil(getResp)

	// 2 - DELETE
	deleteResp, err := client.Delete(context.TODO(), &databasev1.IndexRuleRegistryServiceDeleteRequest{
		Metadata: meta,
	})
	req.NoError(err)
	req.NotNil(deleteResp)
	req.True(deleteResp.GetDeleted())

	// 3 - GET -> Nil
	_, err = client.Get(context.TODO(), &databasev1.IndexRuleRegistryServiceGetRequest{
		Metadata: meta,
	})
	errStatus, _ := status.FromError(err)
	req.Equal(errStatus.Message(), schema.ErrEntityNotFound.Error())

	// 4 - CREATE
	_, err = client.Create(context.TODO(), &databasev1.IndexRuleRegistryServiceCreateRequest{IndexRule: getResp.GetIndexRule()})
	req.NoError(err)

	// 5 - GET - > Not Nil
	getResp, err = client.Get(context.TODO(), &databasev1.IndexRuleRegistryServiceGetRequest{
		Metadata: meta,
	})
	req.NoError(err)
	req.NotNil(getResp)
}
