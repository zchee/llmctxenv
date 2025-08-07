// Copyright 2025 The llmctxenv Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package cmd provides the llmctxenv command-line interface.
package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

type llmCLIEnvCmd struct {
	cmd         *cobra.Command
	leveler     *slog.LevelVar
	logger      *slog.Logger
	verbose     bool
	showVersion bool
}

// New setup commands of [llmCLIEnvCmd].
func New() *llmCLIEnvCmd {
	leveler := &slog.LevelVar{}
	leveler.Set(slog.LevelInfo)

	opts := &slog.HandlerOptions{
		Level: leveler,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	llmCLIEnv := &llmCLIEnvCmd{
		leveler: leveler,
		logger:  logger,
	}

	cmd := &cobra.Command{
		Use:     "llmctxenv <subcommand>",
		Version: "v0.0.0",
		Short:   "Manages the LLM CLIs context environment.",
		Args:    cobra.MaximumNArgs(1),
	}
	// Handle "--verbose" flag
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if llmCLIEnv.verbose {
			llmCLIEnv.leveler.Set(slog.LevelDebug)
		}
	}

	// Set version flag to only root command
	cmd.Flags().Bool("version", false, "Show "+cmd.Name()+" version.") // version flag is root only
	cmd.SetVersionTemplate("{{ with .Name }}{{ printf \"%s \" . }}{{ end }}{{ printf \"version: %s \" .Version }}\n")

	// Set persistent flags
	fs := cmd.PersistentFlags()
	fs.BoolVar(&llmCLIEnv.verbose, "verbose", false, "Set verbose mode")

	cmd.AddCommand(NewListCmd())

	llmCLIEnv.cmd = cmd

	return llmCLIEnv
}

// Execute executes the [llmCLIEnvCmd] root command.
func (c *llmCLIEnvCmd) Execute() error {
	ctx, cancel := signal.NotifyContext(context.Background(), unix.SIGINT, unix.SIGTERM)
	defer cancel()

	return c.cmd.ExecuteContext(ctx)
}
