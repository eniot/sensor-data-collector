package main

import (
	"fmt"
	"os"
)

func main() {
	collectorCmd := collectorCmd()
	collectorCmd.AddCommand(apiCmd())
	if err := collectorCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
