module github.com/tschroed/spellingbee

go 1.21

toolchain go1.23.1

require (
	github.com/google/go-cmp v0.6.0
	golang.org/x/exp v0.0.0-20240823005443-9b4947da3948
	google.golang.org/grpc v1.66.2
	google.golang.org/protobuf v1.34.2
)

require (
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117 // indirect
)

replace github.com/tschroed/spellingbee => /home/trevors/src/spellingbee

replace github.com/tschroed/spellingbee/server => /home/trevors/src/spellingbee/server

replace github.com/tschroed/spellingbee/server/proto => /home/trevors/src/spellingbee/server/proto
