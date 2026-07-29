module github.com/unionai/cloud/gen/pb-go

go 1.24.6

replace github.com/flyteorg/flyte/v2/gen/go/flyteidl2 => ../flyteidl2

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.9-20250912141014-52f32327d4b0.1
	connectrpc.com/connect v1.19.1
	github.com/flyteorg/flyte/flyteidl v1.16.1
	github.com/flyteorg/flyte/v2/gen/go/flyteidl2 v0.0.0-00010101000000-000000000000
	google.golang.org/genproto/googleapis/api v0.0.0-20250908214217-97024824d090
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
)

require (
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250826171959-ef028d996bc1 // indirect
)
