# raft-grpc-kv

**IMPORTANT:** Do not use any part of this project for production. It is not
meant to be representative of how resilient distributed raft-based systems are
built. It's simply an experiment by the sleep-deprived author to see how it
works.

With that out of the way, this is a simple experiment to see how to use gRPC and
Raft together to implement an (INCREDIBLY BASIC) in-memory key-value store.

## Related reading/watching

- The
  [Practical Distributed Consensus using HashiCorp/raft](https://www.youtube.com/watch?v=EGRmmxVFOfE)
  was the inspiration for this project. The code linked in the video is
  [available on GitHub](https://github.com/jen20/hashiconf-raft) (though it's a
  few years old at this point).
- [Another toy KV store implementation using Raft](https://github.com/otoolep/hraftd)

## Structure

- `./cmd/kvd` -- the server implementation (exposes a GRPC interface)
- `./cmd/kv` -- the client implementation
- `./pkg/fsm` -- the implementation of the Raft FSM
- `./api` -- definition of the gRPC API

## Usage

Build the binaries:

```sh
make
```

Start each of the servers in separate terminals:

```sh
# Initial node (bootstraps due to empty --join-address flag)
./bin/kvd --grpc-address 127.0.0.1:8080 --raft-address 127.0.0.1:7080
```

```sh
# Second node
./bin/kvd --grpc-address 127.0.0.1:8081 --raft-address 127.0.0.1:7081 --join-address 127.0.0.1:8080
```

```sh
# Third node
./bin/kvd --grpc-address 127.0.0.1:8082 --raft-address 127.0.0.1:7082 --join-address 127.0.0.1:8080
```

Use the client to interact:

```sh
./bin/kv get foo
# Error: rpc error: code = NotFound desc = key "foo" not found

./bin/kv set foo bar

# Query the default node (:8080)
./bin/kv get foo
# bar

# Query a follower node (:8081)
./bin/kv get --address 127.0.0.1:8081 foo
# bar
```

## Next steps

- Currently, `kv set` requests must be directed to the leader. I'd like to
  automagically forward these to the leader if the gRPC request is handled by a
  follower.
