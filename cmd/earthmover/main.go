// Command earthmover starts the Gomoku AI HTTP server.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/todd/earthmover/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: earthmover <port> [static-dir]")
		fmt.Println("  port:       port number to listen on")
		fmt.Println("  static-dir: path to frontend files (default: ./static)")
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid port: %s\n", os.Args[1])
		os.Exit(1)
	}

	staticDir := "./static"
	if len(os.Args) >= 3 {
		staticDir = os.Args[2]
	}

	srv := server.New(staticDir)
	if err := srv.Run(port); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
