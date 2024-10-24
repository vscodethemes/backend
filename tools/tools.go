//go:build tools
// +build tools

package tools

import (
	_ "github.com/amacneil/dbmate"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/riverqueue/river/cmd/river"
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
)
