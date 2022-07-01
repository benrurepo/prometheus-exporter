// Copyright 2022 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build go1.17
// +build go1.17

// A minimal example of how to include Prometheus instrumentation.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

var version string = "local"

func main() {
	app := &cli.App{
		Name:      "spacelift-promex",
		Usage:     "Exports metrics from your Spacelift account to Prometheus",
		Commands:  []*cli.Command{serveCommand},
		Version:   version,
		Copyright: fmt.Sprintf("Copyright (c) %d spacelift-io", time.Now().Year()),
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
