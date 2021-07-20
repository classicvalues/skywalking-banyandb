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
	"github.com/apache/skywalking-banyandb/api/common"
	apischema "github.com/apache/skywalking-banyandb/api/schema"
	logical "github.com/apache/skywalking-banyandb/pkg/query/logical"
	"io"
	"log"
	"net"

	flatbuffers "github.com/google/flatbuffers/go"
	grpclib "google.golang.org/grpc"

	"github.com/apache/skywalking-banyandb/api/event"
	v1 "github.com/apache/skywalking-banyandb/api/fbs/v1"
	"github.com/apache/skywalking-banyandb/banyand/discovery"
	"github.com/apache/skywalking-banyandb/banyand/queue"
	"github.com/apache/skywalking-banyandb/pkg/bus"
	"github.com/apache/skywalking-banyandb/pkg/logger"
	"github.com/apache/skywalking-banyandb/pkg/run"
)

type Server struct {
	addr       string
	log        *logger.Logger
	ser        *grpclib.Server
	pipeline   queue.Queue
	repo       discovery.ServiceRepo
	shardInfo  *shardInfo
	seriesInfo *seriesInfo
}

type shardInfo struct {
	log *logger.Logger
}

func (s *shardInfo) Rev(message bus.Message) (resp bus.Message) {
	data, ok := message.Data().([]byte)
	if !ok {
		s.log.Warn().Msg("invalid event data type")
		return
	}
	shardEvent := v1.GetRootAsShardEvent(data, 0)
	s.log.Info().
		Str("action", shardEvent.Action().String()).
		Uint64("shardID", shardEvent.Shard(nil).Id()).
		Msg("received a shard event")
	return
}

type seriesInfo struct {
	log *logger.Logger
}

var seriesEventData *v1.SeriesEvent
func (s *seriesInfo) Rev(message bus.Message) (resp bus.Message) {
	data, ok := message.Data().([]byte)
	if !ok {
		s.log.Warn().Msg("invalid event data type")
		return
	}
	seriesEvent := v1.GetRootAsSeriesEvent(data, 0)
	seriesEventData = seriesEvent
	s.log.Info().
		Str("action", seriesEvent.Action().String()).
		Str("name", string(seriesEvent.Series(nil).Name())).
		Str("group", string(seriesEvent.Series(nil).Group())).
		Msg("received a shard event")
	return
}

func (s *Server) PreRun() error {
	s.log = logger.GetLogger("liaison-grpc")
	s.shardInfo.log = s.log
	s.seriesInfo.log = s.log
	err := s.repo.Subscribe(event.TopicShardEvent, s.shardInfo)
	if err != nil {
		return err
	}
	return s.repo.Subscribe(event.TopicSeriesEvent, s.seriesInfo)
}

func NewServer(ctx context.Context, pipeline queue.Queue, repo discovery.ServiceRepo) *Server {
	return &Server{
		pipeline:   pipeline,
		repo:       repo,
		shardInfo:  &shardInfo{},
		seriesInfo: &seriesInfo{},
	}
}

func (s *Server) Name() string {
	return "grpc"
}

func (s *Server) FlagSet() *run.FlagSet {
	fs := run.NewFlagSet("grpc")
	fs.StringVarP(&s.addr, "addr", "", ":17912", "the address of banyand listens")
	return fs
}

func (s *Server) Validate() error {
	return nil
}

func (s *Server) Serve() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.log.Fatal().Err(err).Msg("Failed to listen")
	}

	s.ser = grpclib.NewServer(grpclib.CustomCodec(flatbuffers.FlatbuffersCodec{}))
	//s.ser = grpclib.NewServer()

	v1.RegisterTraceServer(s.ser, &TraceServer{})

	return s.ser.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.log.Info().Msg("stopping")
	s.ser.GracefulStop()
}

type TraceServer struct {
	v1.UnimplementedTraceServer
}

func (t *TraceServer) Write(TraceWriteServer v1.Trace_WriteServer) error {
	for {
		writeEntity, err := TraceWriteServer.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		//log.Println("writeEntity:", writeEntity)
		ana := logical.DefaultAnalyzer()
		metadata := common.Metadata{
			KindVersion: apischema.SeriesKindVersion,
			Spec:        writeEntity.MetaData(nil),
		}
		schema, ruleError := ana.BuildTraceSchema(context.TODO(), metadata)
		if ruleError != nil {
			return  ruleError
		}
		seriesIdLen := seriesEventData.FieldNamesCompositeSeriesIdLength()
		for i := 0; i < seriesIdLen; i++ {
			id := seriesEventData.FieldNamesCompositeSeriesId(i)
			if defined, sub := schema.FieldSubscript(string(id)); defined {
				log.Println("FieldSubscript", sub)
				var field v1.Field
				if ok := writeEntity.Entity(nil).Fields(&field, sub); !ok {
					return nil
				}

			}
		}
		builder := flatbuffers.NewBuilder(0)
		v1.WriteResponseStart(builder)
		builder.Finish(v1.WriteResponseEnd(builder))
		if errSend := TraceWriteServer.Send(builder); errSend != nil {
			return errSend
		}
		//writeEntity.Entity().Fields()
		//writeEntity.MetaData(nil).Group()
		//serviceID+instanceID
		//seriesID := hash(fieds, f1, f2)
		//shardID := shardingFunc(seriesID, shardNum)
		//queue
	}
}

func (t *TraceServer) Query(ctx context.Context, entityCriteria *v1.EntityCriteria) (*flatbuffers.Builder, error) {
	log.Println("entityCriteria:", entityCriteria)

	// receive entity, then serialize entity
	b := flatbuffers.NewBuilder(0)
	v1.EntityStart(b)
	b.Finish(v1.EntityEnd(b))

	return b, nil
}
