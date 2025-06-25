package main

import (
	"fmt"
	"os"

	_ "github.com/Axway/agent-sdk/pkg/traceability"

	"github.com/vivekschauhan/sample_traceability_agent/pkg/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
