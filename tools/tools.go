//go:build tools
// +build tools

package tools

//go:generate go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
//go:generate go install github.com/bflad/tfproviderlint/cmd/tfproviderlintx
//go:generate go install github.com/go-delve/delve/cmd/dlv
//go:generate go install github.com/golangci/golangci-lint/cmd/golangci-lint
//go:generate go install github.com/google/go-github/github
//go:generate go install golang.org/x/tools/gopls@latest

import (
	_ "github.com/bflad/tfproviderlint/cmd/tfproviderlintx"
	_ "github.com/go-delve/delve/cmd/dlv"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/google/go-github/github"
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	_ "golang.org/x/tools/gopls"
)
