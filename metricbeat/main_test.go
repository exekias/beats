package main

// This file is mandatory as otherwise the packetbeat.test binary is not generated correctly.

import (
	"flag"
	"testing"

	"github.com/elastic/beats/metricbeat/cmd"
)

var systemTest *bool

func init() {
	systemTest = flag.Bool("systemTest", false, "Set to true when running system tests")
	cmd.RootCmd.Flags().AddGoFlag(flag.CommandLine.Lookup("systemTest"))
}

// Test started when the test binary is started. Only calls main.
func TestSystem(t *testing.T) {

	if *systemTest {
		main()
	}
}
