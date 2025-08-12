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

package contextmanager_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/zchee/llmctxenv/contextmanager"
)

// Test helper functions
func setupTestEnv(t *testing.T) (cleanup func()) {
	t.Helper()

	originalRoot := os.Getenv(contextmanager.EnvRoot)
	return func() {
		if originalRoot != "" {
			os.Setenv(contextmanager.EnvRoot, originalRoot)
		} else {
			os.Unsetenv(contextmanager.EnvRoot)
		}
	}
}

func TestProvider_String(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		provider contextmanager.Provider
		want     string
	}{
		"claude provider": {
			provider: contextmanager.ProviderClaudeCode,
			want:     "claude",
		},
		"gemini-cli provider": {
			provider: contextmanager.ProviderGeminiCLI,
			want:     "gemini-cli",
		},
		"qwen-cli provider": {
			provider: contextmanager.ProviderQwenCLI,
			want:     "qwen-cli",
		},
		"codex provider": {
			provider: contextmanager.ProviderCodex,
			want:     "codex",
		},
		"opencode provider": {
			provider: contextmanager.ProviderOpenCode,
			want:     "opencode",
		},
		"goose provider": {
			provider: contextmanager.ProviderGoose,
			want:     "goose",
		},
		"crush provider": {
			provider: contextmanager.ProviderCrush,
			want:     "crush",
		},
		"empty provider": {
			provider: contextmanager.Provider(""),
			want:     "",
		},
		"custom provider": {
			provider: contextmanager.Provider("custom"),
			want:     "custom",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tt.provider.String()
			if got != tt.want {
				t.Errorf("Provider.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderConstants(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		provider contextmanager.Provider
		want     string
	}{
		"ProviderClaudeCode": {contextmanager.ProviderClaudeCode, "claude"},
		"ProviderGeminiCLI":  {contextmanager.ProviderGeminiCLI, "gemini-cli"},
		"ProviderQwenCLI":    {contextmanager.ProviderQwenCLI, "qwen-cli"},
		"ProviderCodex":      {contextmanager.ProviderCodex, "codex"},
		"ProviderOpenCode":   {contextmanager.ProviderOpenCode, "opencode"},
		"ProviderGoose":      {contextmanager.ProviderGoose, "goose"},
		"ProviderCrush":      {contextmanager.ProviderCrush, "crush"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if string(tt.provider) != tt.want {
				t.Errorf("Provider constant %s = %v, want %v", name, string(tt.provider), tt.want)
			}
		})
	}
}

func TestContextFiles(t *testing.T) {
	t.Parallel()

	expectedFiles := map[contextmanager.Provider][]string{
		contextmanager.ProviderClaudeCode: {"CLAUDE.md"},
		contextmanager.ProviderGeminiCLI:  {"GEMINI.md"},
		contextmanager.ProviderQwenCLI:    {"QWEN.md"},
		contextmanager.ProviderCodex:      {"AGENTS.md"},
		contextmanager.ProviderOpenCode:   {"AGENTS.md"},
		contextmanager.ProviderGoose:      {".goosehints"},
		contextmanager.ProviderCrush:      {"CRUSH.md"},
	}

	if !reflect.DeepEqual(contextmanager.ContextFiles, expectedFiles) {
		t.Errorf("ContextFiles mapping incorrect.\nGot: %+v\nWant: %+v", contextmanager.ContextFiles, expectedFiles)
	}

	// Test that all providers have corresponding context files
	allProviders := []contextmanager.Provider{
		contextmanager.ProviderClaudeCode,
		contextmanager.ProviderGeminiCLI,
		contextmanager.ProviderQwenCLI,
		contextmanager.ProviderCodex,
		contextmanager.ProviderOpenCode,
		contextmanager.ProviderGoose,
		contextmanager.ProviderCrush,
	}

	for _, provider := range allProviders {
		t.Run(fmt.Sprintf("provider_%s_has_context_files", provider), func(t *testing.T) {
			t.Parallel()

			files, exists := contextmanager.ContextFiles[provider]
			if !exists {
				t.Errorf("Provider %s missing from ContextFiles mapping", provider)
			}
			if len(files) == 0 {
				t.Errorf("Provider %s has no context files defined", provider)
			}
		})
	}
}

func TestGlobalDir(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tests := map[string]struct {
		envRoot  string // value to set for LLMCTXENV_ROOT
		provider contextmanager.Provider
		wantPath string // relative to the root
	}{
		"default root with claude": {
			envRoot:  "", // use default
			provider: contextmanager.ProviderClaudeCode,
			wantPath: "global/claude",
		},
		"custom root with claude": {
			envRoot:  "/tmp/custom-llmctx",
			provider: contextmanager.ProviderClaudeCode,
			wantPath: "global/claude",
		},
		"custom root with gemini-cli": {
			envRoot:  "/tmp/custom-llmctx",
			provider: contextmanager.ProviderGeminiCLI,
			wantPath: "global/gemini-cli",
		},
		"custom root with goose": {
			envRoot:  "/opt/llmctx",
			provider: contextmanager.ProviderGoose,
			wantPath: "global/goose",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.envRoot != "" {
				os.Setenv(contextmanager.EnvRoot, tt.envRoot)
				// Need to update the global variable since it's initialized at package load
				contextmanager.LLMCtxEnvRoot = tt.envRoot
			} else {
				os.Unsetenv(contextmanager.EnvRoot)
				home := os.Getenv("HOME")
				contextmanager.LLMCtxEnvRoot = filepath.Join(home, ".llmctxenv")
			}

			got := contextmanager.GlobalDir(tt.provider)
			expected := filepath.Join(contextmanager.LLMCtxEnvRoot, tt.wantPath)

			if got != expected {
				t.Errorf("GlobalDir(%v) = %v, want %v", tt.provider, got, expected)
			}
		})
	}
}

func TestLocalDir(t *testing.T) {
	// Setup test environment
	cleanup := setupTestEnv(t)
	defer cleanup()
	testRoot := "/tmp/test-llmctx"
	os.Setenv(contextmanager.EnvRoot, testRoot)
	contextmanager.LLMCtxEnvRoot = testRoot

	tests := map[string]struct {
		provider      contextmanager.Provider
		projectDir    string // unused in current implementation
		currentDir    string // what os.Getwd should return
		userHomeDir   string // what os.UserHomeDir should return
		wantRelPath   string // expected path relative to testRoot/local/provider/
		wantErr       bool
		skipOnWindows bool
	}{
		"simple path under home": {
			provider:    contextmanager.ProviderClaudeCode,
			projectDir:  "",
			currentDir:  "/home/user/projects/myapp",
			userHomeDir: "/home/user",
			wantRelPath: "projects-myapp",
			wantErr:     false,
		},
		"path with dots": {
			provider:    contextmanager.ProviderClaudeCode,
			projectDir:  "",
			currentDir:  "/home/user/project.name/sub.dir",
			userHomeDir: "/home/user",
			wantRelPath: "project-name-sub-dir",
			wantErr:     false,
		},
		"path with uppercase letters": {
			provider:    contextmanager.ProviderGeminiCLI,
			projectDir:  "",
			currentDir:  "/home/user/MyProject/SubDir",
			userHomeDir: "/home/user",
			wantRelPath: "!my!project-!sub!dir",
			wantErr:     false,
		},
		"path with mixed special chars": {
			provider:    contextmanager.ProviderCodex,
			projectDir:  "",
			currentDir:  "/home/user/My.Project/sub/Dir.V2",
			userHomeDir: "/home/user",
			wantRelPath: "!my-!project-sub-!dir-!v2",
			wantErr:     false,
		},
		"current dir is home": {
			provider:    contextmanager.ProviderClaudeCode,
			projectDir:  "",
			currentDir:  "/home/user",
			userHomeDir: "/home/user",
			wantRelPath: "", // empty after stripping home prefix
			wantErr:     false,
		},
		"path outside home directory": {
			provider:    contextmanager.ProviderClaudeCode,
			projectDir:  "",
			currentDir:  "/tmp/project",
			userHomeDir: "/home/user",
			wantRelPath: "-tmp-project", // leading separator becomes dash
			wantErr:     false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.skipOnWindows && strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
				t.Skip("Skipping on Windows")
			}

			// For this test, we'll use the actual function but verify the logic
			// Since we can't easily mock os.Getwd and os.UserHomeDir, we'll test with current environment
			// and focus on the path sanitization logic separately
			got, err := contextmanager.LocalDir(tt.provider, tt.projectDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LocalDir() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("LocalDir() unexpected error: %v", err)
				return
			}

			// Verify the result has the expected structure
			expectedPrefix := filepath.Join(testRoot, "local", tt.provider.String())
			if !strings.HasPrefix(got, expectedPrefix) {
				t.Errorf("LocalDir() = %v, expected to start with %v", got, expectedPrefix)
			}
		})
	}
}

func TestPathSanitization(t *testing.T) {
	// Test the dirnameReplacer logic by testing what LocalDir would do
	// This helps us test the sanitization logic in isolation
	tests := map[string]struct {
		input    string
		expected string
	}{
		"real path": {
			input:    "go/src/github.com/zchee/llmctxenv",
			expected: "go-src-github-com-zchee-llmctxenv",
		},
		"dots to dashes": {
			input:    "project.name",
			expected: "project-name",
		},
		"separators to dashes": {
			input:    "project/sub/dir",
			expected: "project-sub-dir",
		},
		"uppercase A": {
			input:    "ProjectA",
			expected: "!project!a",
		},
		"uppercase B": {
			input:    "ProjectB",
			expected: "!project!b",
		},
		"uppercase Z": {
			input:    "ProjectZ",
			expected: "!project!z",
		},
		"all uppercase": {
			input:    "MYPROJECT",
			expected: "!m!y!p!r!o!j!e!c!t",
		},
		"mixed case with dots": {
			input:    "My.Project.V2",
			expected: "!my-!project-!v2",
		},
		"complex path": {
			input:    "My.App/src/Main.go",
			expected: "!my-!app-src-!main-go",
		},
		"empty string": {
			input:    "",
			expected: "",
		},
		"only separators": {
			input:    "//",
			expected: "--",
		},
		"only dots": {
			input:    "...",
			expected: "---",
		},
		"numbers unchanged": {
			input:    "project123",
			expected: "project123",
		},
		"lowercase unchanged":  {"myproject", "myproject"},
		"underscore unchanged": {"my_project", "my_project"},
		"hyphen unchanged":     {"my-project", "my-project"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := contextmanager.DirnameReplacer.Replace(tt.input)
			if got != tt.expected {
				t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLocalDir_EdgeCases(t *testing.T) {
	// These tests focus on edge cases and error conditions
	// Since we can't easily mock os.Getwd and os.UserHomeDir in Go without
	// dependency injection, we'll test what we can with the current setup

	cleanup := setupTestEnv(t)
	defer cleanup()
	testRoot := "/tmp/test-llmctx-edge"
	os.Setenv(contextmanager.EnvRoot, testRoot)
	contextmanager.LLMCtxEnvRoot = testRoot

	// Test with various providers
	providers := []contextmanager.Provider{
		contextmanager.ProviderClaudeCode,
		contextmanager.ProviderGeminiCLI,
		contextmanager.ProviderQwenCLI,
		contextmanager.ProviderCodex,
		contextmanager.ProviderOpenCode,
		contextmanager.ProviderGoose,
		contextmanager.ProviderCrush,
	}
	for _, provider := range providers {
		t.Run(fmt.Sprintf("provider_%s", provider), func(t *testing.T) {
			got, err := contextmanager.LocalDir(provider, "")
			if err != nil {
				t.Errorf("LocalDir(%v, \"\") unexpected error: %v", provider, err)
				return
			}

			expectedPrefix := filepath.Join(testRoot, "local", provider.String())
			if !strings.HasPrefix(got, expectedPrefix) {
				t.Errorf("LocalDir(%v) = %v, expected to start with %v", provider, got, expectedPrefix)
			}
		})
	}
}

func BenchmarkGlobalDir(b *testing.B) {
	provider := contextmanager.ProviderClaudeCode
	for b.Loop() {
		_ = contextmanager.GlobalDir(provider)
	}
}

func BenchmarkLocalDir(b *testing.B) {
	provider := contextmanager.ProviderClaudeCode
	for b.Loop() {
		_, _ = contextmanager.LocalDir(provider, "")
	}
}

func BenchmarkPathSanitization(b *testing.B) {
	testPath := "My.Complex/Project.Name/With.Many/Components.V2"

	for b.Loop() {
		_ = contextmanager.DirnameReplacer.Replace(testPath)
	}
}
