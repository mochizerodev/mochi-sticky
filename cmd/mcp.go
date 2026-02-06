package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mochi-sticky/internal/mcp"

	"github.com/spf13/cobra"
)

var mcpRootFlag string
var mcpTimeoutFlag string

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine the root directory: --root flag > MOCHI_STICKY_ROOT env > current directory
		root := mcpRootFlag
		if root == "" {
			root = os.Getenv("MOCHI_STICKY_ROOT")
		}

		// Change to the specified root directory if provided
		if root != "" {
			if err := os.Chdir(root); err != nil {
				return fmt.Errorf("failed to change directory to %s: %w", root, err)
			}
		}

		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		// Allow missing storage root - it will be set dynamically from workspace roots in initialize request
		storageRoot, err := resolveStorageRoot(workingDir, true)
		if err != nil {
			return err
		}
		server, err := mcp.NewServer(workingDir, storageRoot)
		if err != nil {
			return err
		}

		// Parse timeout duration
		var timeout time.Duration
		if mcpTimeoutFlag != "" {
			timeout, err = time.ParseDuration(mcpTimeoutFlag)
			if err != nil {
				return fmt.Errorf("invalid timeout value: %w", err)
			}
		} else if envTimeout := os.Getenv("MOCHI_MCP_TIMEOUT"); envTimeout != "" {
			timeout, err = time.ParseDuration(envTimeout)
			if err != nil {
				return fmt.Errorf("invalid MOCHI_MCP_TIMEOUT value: %w", err)
			}
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		return server.ServeContextWithTimeout(ctx, os.Stdin, os.Stdout, timeout)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().StringVar(&mcpRootFlag, "root", "", "Storage root directory (overrides MOCHI_STICKY_ROOT env var)")
	mcpCmd.Flags().StringVar(&mcpTimeoutFlag, "timeout", "", "Idle timeout duration (e.g., '30m', '1h') - server exits if no requests received within this period. Defaults to no timeout. Can also be set via MOCHI_MCP_TIMEOUT env var.")
}
