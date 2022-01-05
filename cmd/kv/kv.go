package main

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/travigd/raft-grpc-kv/api/v1"
	"google.golang.org/grpc"
	"os"
)

var (
	addr string
)

var rootCmd = &cobra.Command{
	Use:          "kv",
	SilenceUsage: true,
}

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a value from a key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}
		key := args[0]
		kv, err := kvClient()
		if err != nil {
			return err
		}
		res, err := kv.Get(context.Background(), &api.GetRequest{Key: key})
		if err != nil {
			return err
		}
		println(res.Value)
		return nil
	},
}

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a value for a key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return cmd.Usage()
		}
		key := args[0]
		value := args[1]
		kv, err := kvClient()
		if err != nil {
			return err
		}
		_, err = kv.Set(context.Background(), &api.SetRequest{Key: key, Value: value})
		if err != nil {
			return err
		}
		return nil
	},
}

func main() {
	rootCmd.PersistentFlags().StringVarP(&addr, "address", "a", "127.0.0.1:8080", "address of the kv kvd")
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func kvClient() (api.KVClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return api.NewKVClient(conn), nil
}
