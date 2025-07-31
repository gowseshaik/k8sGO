package main

import (
	"fmt"
	"os"

	"k8sgo/pkg/ui"
)

const (
	AppName    = "k8sgo"
	AppVersion = "1.0.0"
	AppDesc    = "Kubernetes & OpenShift CLI Tool"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	app := ui.NewApp(AppName, AppVersion, AppDesc)
	return app.Run()
}
