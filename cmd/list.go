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

package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/zchee/llmctxenv/contextmanager"
	"github.com/zchee/llmctxenv/fileio"
)

type listCmd struct {
	logger   *slog.Logger
	provider contextmanager.Provider
}

// NewListCmd returns the `list` subcommand that lists managed system context files.
func NewListCmd() *cobra.Command {
	l := &listCmd{
		logger: slog.Default().WithGroup("list"),
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List managed system context files",
	}
	cmd.RunE = l.RunList

	f := cmd.Flags()
	f.StringVarP((*string)(&l.provider), "provider", "p", "", "manages system context provider name")

	return cmd
}

// RunList runs the `list` subcommand which lists managed system context files.
//
// TODO(zchee): fix documentations.
func (c *listCmd) RunList(cmd *cobra.Command, args []string) error {
	globalDir := contextmanager.SystemContextGlobalDir(c.provider)

	c.logger.DebugContext(cmd.Context(), "RunList",
		slog.Any("args", args),
		slog.String("global_directory", globalDir),
		slog.String("provider", c.provider.String()),
	)

	if c.provider == "" {
		return fmt.Errorf("--provider flag must be not empty")
	}

	if !fileio.IsExist(globalDir) {
		// Create instructionsDir if not exist
		if err := os.MkdirAll(globalDir, 0o700); err != nil {
			return fmt.Errorf("mkdir all %s path: %w", globalDir, err)
		}
		// Early return if not found instructionsDir
		return nil
	}

	ents, err := os.ReadDir(globalDir)
	if err != nil {
		return fmt.Errorf("ReadDir %s: %w", globalDir, err)
	}

	files := make([]string, 0, len(ents))
	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}
		files = append(files, ent.Name())
	}
	slices.Sort(files)

	cmd.Printf("files:\n%s", strings.Join(files, "\n"))

	return nil
}
