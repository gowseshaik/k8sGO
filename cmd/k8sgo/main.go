package main

import (
	"fmt"
	"os"

	"k8sgo/pkg/ui"
)

const (
	Version     = "1.0.0"
	Name        = "k8sgo"
	Description = "Cross-platform terminal-based UI tool for managing Kubernetes and OpenShift clusters"
)

func main() {
	app := ui.NewApp(Name, Version, Description)
	
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}