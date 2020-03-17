// Copyright ©2020 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build tools

package tools

import (
	_ "github.com/mattn/goveralls"       // for coverage"
	_ "golang.org/x/tools/cmd/cover"     // for coverage
	_ "golang.org/x/tools/cmd/goimports" // for format check
)
