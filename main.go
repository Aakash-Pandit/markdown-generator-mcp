package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/Aakash-Pandit/markdown-generator-mcp/internal/mcpserver"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--install" {
		self, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		cmd := exec.Command("claude", "mcp", "add", "--scope", "user", "markdown-generator", "--", self)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Registration failed: %v\nMake sure Claude Code CLI is installed.", err)
		}
		fmt.Println("\nDone! Restart Claude Code, then say 'make a markdown'.")
		return
	}
	if err := mcpserver.Start(); err != nil {
		log.Fatal(err)
	}
}
