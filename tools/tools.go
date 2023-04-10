// +build tools

package tools

import (
	_ "github.com/daixiang0/gci"
	_ "gotest.tools/gotestsum"
	_ "mvdan.cc/gofumpt"
	_ "golang.org/x/tools/cmd/goimports"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
)
