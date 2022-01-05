.PHONY: bin
bin:
	go build -o ./bin/kv ./cmd/kv
	go build -o ./bin/kvd ./cmd/kvd

protoc_go_flags = --go_out=. --go_opt=paths=source_relative
protoc_go_grpc_flags = --go-grpc_out=. --go-grpc_opt=paths=source_relative

.PHONY: protoc
protoc:
	protoc $(protoc_go_flags) $(protoc_go_grpc_flags) ./api/v1/*.proto