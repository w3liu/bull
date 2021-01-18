module github.com/w3liu/bull

go 1.13

replace (
	github.com/imdario/mergo => github.com/imdario/mergo v0.3.8
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/mitchellh/hashstructure v1.1.0
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20200625001655-4c5254603344
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/grpc v1.31.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)
