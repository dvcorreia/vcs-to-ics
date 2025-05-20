// Copyright (c) Diogo Correia
// SPDX-License-Identifier: MIT

package main

import "fmt"

// Provisioned by ldflags
var (
	version    string
	commitHash string
)

func main() {
	fmt.Printf("hello, world! version=%s commitHash=%s", version, commitHash)
}
