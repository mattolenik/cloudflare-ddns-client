//go:build tools
// +build tools

package tools

// This file contains dummy imports for tools used during build.
// It allows these modules to be present in go.mod and not be removed
// during `go mod tidy`

import (
	// gotestsum is a test runner that produces test output in other formats, such as JUnit XML.
	_ "gotest.tools/gotestsum"

	// This one is actually to prevent a misleading Dependabot error. Without it,
	// Dependabot will claim that this module is using an outdated version of this
	// module even when it isn't, presumably due to it being an indirect dependency from
	// the lixiangzhong/dnsutil package. Adding it here creates a direct dependency and
	// prevents go mod tidy from removing it from this module's go.mod.
	_ "github.com/miekg/dns"

	_ "github.com/golang/mock/mockgen"
)
