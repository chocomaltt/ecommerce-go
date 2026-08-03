module github.com/chocomaltt/ecommerce-go/apps/order-service

go 1.26.4

replace github.com/chocomaltt/ecommerce-go/common-rpc => ../../common-rpc

require (
	github.com/chocomaltt/ecommerce-go/common-rpc v0.0.0
	google.golang.org/grpc v1.83.0
)

require (
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
