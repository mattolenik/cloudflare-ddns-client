// +build tools

package tools

// This file contains dummy imports for tools used during build.
// It allows these modules to be present in go.mod and not be removed
// during `go mod tidy`

import (
	_ "gotest.tools/gotestsum"
)
