//go:build tools
// +build tools

package tools

import (
	_ "github.com/amacneil/dbmate"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
	_ "github.com/riverqueue/river/cmd/river"
)
