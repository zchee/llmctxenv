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

package contextmanager

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvRoot is the environment variable that specifies the root directory for llmctxenv context environments.
const EnvRoot = "LLMCTXENV_ROOT"

var (
	LLMCtxEnvRoot = cmp.Or(
		os.Getenv(EnvRoot),
		filepath.Join(os.Getenv("HOME"), ".llmctxenv"),
	)
)

// Provider represents the type of LLM provider for which the system context is defined.
type Provider string

const (
	ProviderClaudeCode Provider = "claude"
	ProviderGeminiCLI  Provider = "gemini-cli"
	ProviderQwenCLI    Provider = "qwen-cli"
	ProviderCodex      Provider = "codex"
	ProviderOpenCode   Provider = "opencode"
	ProviderGoose      Provider = "goose"
)

var SystemContextFiles = map[Provider][]string{
	ProviderClaudeCode: {
		"CLAUDE.md",
	},
	ProviderGeminiCLI: {
		"GEMINI.md",
	},
	ProviderQwenCLI: {
		"QWEN.md",
	},
	ProviderCodex: {
		"AGENTS.md",
	},
	ProviderOpenCode: {
		"AGENTS.md",
	},
	ProviderGoose: {
		".goosehints",
	},
}

// SystemContextGlobalDir returns the directory path for the global system context of a given provider.
func SystemContextGlobalDir(provider Provider) string {
	return filepath.Join(LLMCtxEnvRoot, "global", string(provider))
}

var dirnameReplacer = strings.NewReplacer(
	".", "-",
	string(filepath.Separator), "-",
)

// SystemContextLocalDir returns the directory path for the local system context of a given provider.
func SystemContextLocalDir(provider Provider, projectDir string) (string, error) {
	path, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get current directory: %w", err)
	}
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get current user home directory: %w", err)
	}

	path = strings.TrimPrefix(path, home+string(filepath.Separator))
	sanitized := dirnameReplacer.Replace(path)

	return filepath.Join(LLMCtxEnvRoot, "local", string(provider), sanitized), nil
}
