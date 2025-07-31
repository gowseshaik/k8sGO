package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"k8sgo-web/pkg/server"
)

var (
	Version = "1.0.0-web"
)

func main() {
	fmt.Printf("🌐 K8sGO Web GUI v%s\n", Version)
	fmt.Println("============================")
	
	// Create server
	srv := server.NewServer(Version)
	
	// Start server in background
	go func() {
		fmt.Println("🚀 Starting web server...")
		if err := srv.Start(":8080"); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()
	
	// Wait a moment for server to start
	time.Sleep(1 * time.Second)
	
	// Open browser automatically
	url := "http://localhost:8080"
	fmt.Printf("🌐 Web GUI available at: %s\n", url)
	fmt.Println("📋 Features:")
	fmt.Println("   • Native browser copy/paste (Ctrl+C/Ctrl+V)")
	fmt.Println("   • Text selection with mouse")
	fmt.Println("   • Top command integration")
	fmt.Println("   • Resource browsing")
	fmt.Println("   • Context switching")
	fmt.Println("")
	
	if err := openBrowser(url); err != nil {
		fmt.Printf("ℹ️  Please open your browser and go to: %s\n", url)
	} else {
		fmt.Println("✅ Browser opened automatically")
	}
	
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop the server")
	
	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	<-c
	fmt.Println("\n🛑 Shutting down server...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	
	fmt.Println("✅ Server stopped")
}

func openBrowser(url string) error {
	var cmd string
	var args []string
	
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}