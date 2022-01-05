package main

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/spf13/cobra"
	"github.com/travigd/raft-grpc-kv/api/v1"
	"github.com/travigd/raft-grpc-kv/pkg/fsm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"os"
	"time"
)

var (
	grpcAddress string
	raftAddress string
	joinAddress string
)

var cmd = &cobra.Command{
	Use:          "kvd",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		log := hclog.New(&hclog.LoggerOptions{
			Name:   "raft",
			Output: os.Stderr,
		})

		raftConfig := raft.DefaultConfig()
		raftConfig.LocalID = raft.ServerID(raftAddress)
		raftConfig.Logger = log
		transport, err := raft.NewTCPTransport(raftAddress, nil, 3, 10*time.Second, os.Stderr)
		if err != nil {
			return err
		}

		fsm := fsm.New()
		store := raft.NewInmemStore()
		r, err := raft.NewRaft(raftConfig, fsm, store, store, raft.NewInmemSnapshotStore(), transport)

		if joinAddress == "" {
			r.BootstrapCluster(raft.Configuration{
				Servers: []raft.Server{
					{
						ID:      raftConfig.LocalID,
						Address: transport.LocalAddr(),
					},
				},
			})
		} else {
			if err := joinRaft(joinAddress); err != nil {
				return err
			}
		}

		server := Server{
			raft: r,
			fsm:  fsm,
		}
		grpcServer := grpc.NewServer()
		api.RegisterKVServer(grpcServer, server)
		lis, err := net.Listen("tcp", grpcAddress)
		if err != nil {
			return err
		}
		if err := grpcServer.Serve(lis); err != nil {
			return err
		}

		return nil
	},
}

func joinRaft(addr string) error {
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := api.NewKVClient(cc)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		hclog.Default().Info("Attempting to join cluster")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err = client.Join(ctx, &api.JoinRequest{
			Id:      raftAddress,
			Address: raftAddress,
		})
		if err == nil {
			cancel()
			return nil
		}
		cancel()
	}

	return err
}

func main() {
	cmd.Flags().StringVar(&grpcAddress, "grpc-address", "127.0.0.1:8080", "gRPC bind address")
	cmd.Flags().StringVar(&raftAddress, "raft-address", "127.0.0.1:7080", "raft bind address")
	cmd.Flags().StringVar(&joinAddress, "join-address", "", "address of raft peer to join")
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

type Server struct {
	api.UnsafeKVServer
	raft *raft.Raft
	fsm  *fsm.FSM
}

var _ api.KVServer = &Server{}

func (s Server) Join(_ context.Context, request *api.JoinRequest) (*api.JoinResponse, error) {
	idxFuture := s.raft.AddVoter(raft.ServerID(request.Id), raft.ServerAddress(request.Address), 0, 0)
	if err := idxFuture.Error(); err != nil {
		return nil, err
	}
	return &api.JoinResponse{}, nil
}

func (s Server) Get(_ context.Context, request *api.GetRequest) (*api.GetResponse, error) {
	value, ok := s.fsm.Get(request.Key)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "key %q not found", request.Key)
	}
	return &api.GetResponse{
		Value: value,
	}, nil
}

func (s Server) Set(_ context.Context, request *api.SetRequest) (*api.SetResponse, error) {
	event, err := fsm.SetEvent(request.Key, request.Value)
	if err != nil {
		return nil, err
	}
	apply := s.raft.Apply(event, 5*time.Second)
	if err := apply.Error(); err != nil {
		return nil, err
	}
	return &api.SetResponse{}, nil
}
