module github.com/unionai/terraform-provider-unionai

go 1.24.6

replace github.com/unionai/cloud/gen/pb-go => ./proto/union

replace github.com/flyteorg/flyte/v2/gen/go/flyteidl2 => ./proto/flyteidl2

require (
	github.com/coreos/go-oidc/v3 v3.15.0
	github.com/flyteorg/flyte/flyteidl v1.16.1
	github.com/flyteorg/flyte/v2/gen/go/flyteidl2 v0.0.0-00010101000000-000000000000
	github.com/hashicorp/terraform-plugin-framework v1.16.1
	github.com/hashicorp/terraform-plugin-go v0.29.0
	github.com/hashicorp/terraform-plugin-log v0.9.0
	github.com/hashicorp/terraform-plugin-testing v1.13.3
	github.com/unionai/cloud/gen/pb-go v0.0.0-00010101000000-000000000000
	golang.org/x/oauth2 v0.30.0
	google.golang.org/grpc v1.75.1
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.9-20250912141014-52f32327d4b0.1 // indirect
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	github.com/ProtonMail/go-crypto v1.1.6 // indirect
	github.com/agext/levenshtein v1.2.2 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-cty v1.5.0 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.7.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/hc-install v0.9.2 // indirect
	github.com/hashicorp/hcl/v2 v2.23.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.23.0 // indirect
	github.com/hashicorp/terraform-json v0.25.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.37.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.4.0 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/zclconf/go-cty v1.16.3 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/net v0.45.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	golang.org/x/tools v0.37.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)
