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
	// LLMCtxEnvRoot is the root directory for llmctxenv context environments.
	LLMCtxEnvRoot = cmp.Or(
		os.Getenv(EnvRoot),
		filepath.Join(os.Getenv("HOME"), ".llmctxenv"),
	)
)

// Provider represents the type of LLM provider for which the system context is defined.
type Provider string

// String returns a string representation of the [Provider] name.
func (p Provider) String() string { return string(p) }

const (
	ProviderClaudeCode Provider = "claude"     // https://docs.anthropic.com/en/docs/claude-code/
	ProviderGeminiCLI  Provider = "gemini-cli" // https://github.com/google-gemini/gemini-cli
	ProviderQwenCLI    Provider = "qwen-cli"   // https://github.com/QwenLM/qwen-code
	ProviderCodex      Provider = "codex"      // https://github.com/openai/codex
	ProviderOpenCode   Provider = "opencode"   // https://github.com/sst/opencode
	ProviderGoose      Provider = "goose"      // https://github.com/block/goose
	ProviderCrush      Provider = "crush"      // https://github.com/charmbracelet/crush
)

// ContextFiles maps each [Provider] to a list of filenames that define the system context for that provider.
var ContextFiles = map[Provider][]string{
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
	ProviderCrush: {
		"CRUSH.md",
	},
}

// GlobalDir returns the directory path for the global system context of a given provider.
func GlobalDir(provider Provider) string {
	return filepath.Join(LLMCtxEnvRoot, "global", provider.String())
}

var dirnameReplacer = strings.NewReplacer(
	".", "-",
	string(filepath.Separator), "-",

	"A", "!a",
	"B", "!b",
	"C", "!c",
	"D", "!d",
	"E", "!e",
	"F", "!f",
	"G", "!g",
	"H", "!h",
	"I", "!i",
	"J", "!j",
	"K", "!k",
	"L", "!l",
	"M", "!m",
	"N", "!n",
	"O", "!o",
	"P", "!p",
	"Q", "!q",
	"R", "!r",
	"S", "!s",
	"T", "!t",
	"U", "!u",
	"V", "!v",
	"W", "!w",
	"X", "!x",
	"Y", "!y",
	"Z", "!z",
)

// LocalDir returns the directory path for the local system context of a given provider.
func LocalDir(provider Provider, projectDir string) (string, error) {
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

	return filepath.Join(LLMCtxEnvRoot, "local", provider.String(), sanitized), nil
}
