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

package fileio_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zchee/llmctxenv/internal/fileio"
)

// createFile creates a temporary file with the specified content and permissions.
func createFile(tb testing.TB, dir, name, content string, perm os.FileMode) string {
	tb.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		tb.Fatalf("failed to create test file %s: %v", path, err)
	}
	return path
}

// createDir creates a temporary directory with the specified permissions.
func createDir(tb testing.TB, dir, name string, perm os.FileMode) string {
	tb.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(path, perm); err != nil {
		tb.Fatalf("failed to create test directory %s: %v", path, err)
	}
	return path
}

// createNestedTestStructure creates a complex nested directory structure for testing.
func createNestedTestStructure(t *testing.T, baseDir string) map[string]string {
	t.Helper()
	structure := make(map[string]string)

	// Create directories
	structure["dir1"] = createDir(t, baseDir, "dir1", 0755)
	structure["dir1/subdir1"] = createDir(t, structure["dir1"], "subdir1", 0755)
	structure["dir1/subdir2"] = createDir(t, structure["dir1"], "subdir2", 0700)
	structure["dir2"] = createDir(t, baseDir, "dir2", 0755)

	// Create files
	structure["file1.txt"] = createFile(t, baseDir, "file1.txt", "Hello, World!", 0644)
	structure["file2.txt"] = createFile(t, baseDir, "file2.txt", "Test content", 0600)
	structure["dir1/nested_file.txt"] = createFile(t, structure["dir1"], "nested_file.txt", "Nested content", 0755)
	structure["dir1/subdir1/deep_file.txt"] = createFile(t, structure["dir1/subdir1"], "deep_file.txt", "Deep content", 0644)
	structure["dir1/subdir2/private_file.txt"] = createFile(t, structure["dir1/subdir2"], "private_file.txt", "Private content", 0600)

	return structure
}

// compareFileContent compares the content of two files.
func compareFileContent(t *testing.T, file1, file2 string) {
	t.Helper()

	content1, err := os.ReadFile(file1)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", file1, err)
	}

	content2, err := os.ReadFile(file2)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", file2, err)
	}

	if !bytes.Equal(content1, content2) {
		t.Errorf("file contents differ:\nfile1 (%s): %q\nfile2 (%s): %q",
			file1, string(content1), file2, string(content2))
	}
}

// compareFileMode compares the file permissions of two files.
func compareFileMode(t *testing.T, file1, file2 string) {
	t.Helper()

	info1, err := os.Stat(file1)
	if err != nil {
		t.Fatalf("failed to stat file %s: %v", file1, err)
	}

	info2, err := os.Stat(file2)
	if err != nil {
		t.Fatalf("failed to stat file %s: %v", file2, err)
	}

	if info1.Mode().Perm() != info2.Mode().Perm() {
		t.Errorf("file permissions differ:\nfile1 (%s): %o\nfile2 (%s): %o",
			file1, info1.Mode().Perm(), file2, info2.Mode().Perm())
	}
}

// verifyDirectoryStructure recursively verifies that the directory structure matches expected paths.
func verifyDirectoryStructure(t *testing.T, baseDir string, expectedPaths []string) {
	t.Helper()

	var foundPaths []string
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		if relPath != "." {
			foundPaths = append(foundPaths, relPath)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("failed to walk directory %s: %v", baseDir, err)
	}

	expectedMap := make(map[string]bool)
	for _, path := range expectedPaths {
		expectedMap[path] = true
	}

	foundMap := make(map[string]bool)
	for _, path := range foundPaths {
		foundMap[path] = true
	}

	for path := range expectedMap {
		if !foundMap[path] {
			t.Errorf("expected path %s not found in directory structure", path)
		}
	}

	for path := range foundMap {
		if !expectedMap[path] {
			t.Errorf("unexpected path %s found in directory structure", path)
		}
	}
}

func TestIsExist(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name     string
		setup    func() string
		expected bool
	}{
		{
			name: "existing regular file",
			setup: func() string {
				return createFile(t, tempDir, "existing_file.txt", "test content", 0644)
			},
			expected: true,
		},
		{
			name: "existing directory",
			setup: func() string {
				return createDir(t, tempDir, "existing_dir", 0755)
			},
			expected: true,
		},
		{
			name: "non-existent path",
			setup: func() string {
				return filepath.Join(tempDir, "non_existent_file.txt")
			},
			expected: false,
		},
		{
			name: "empty file",
			setup: func() string {
				return createFile(t, tempDir, "empty_file.txt", "", 0644)
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result := fileio.IsExist(path)
			if result != tt.expected {
				t.Errorf("IsExist(%s) = %v, expected %v", path, result, tt.expected)
			}
		})
	}
}

func TestIsExistSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink tests on Windows")
	}

	tempDir := t.TempDir()

	t.Run("symlink to existing file", func(t *testing.T) {
		target := createFile(t, tempDir, "target.txt", "target content", 0644)
		symlink := filepath.Join(tempDir, "symlink.txt")

		if err := os.Symlink(target, symlink); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}

		if !fileio.IsExist(symlink) {
			t.Error("IsExist should return true for symlink to existing file")
		}
	})

	t.Run("broken symlink", func(t *testing.T) {
		nonExistent := filepath.Join(tempDir, "non_existent_target.txt")
		brokenSymlink := filepath.Join(tempDir, "broken_symlink.txt")

		if err := os.Symlink(nonExistent, brokenSymlink); err != nil {
			t.Fatalf("failed to create broken symlink: %v", err)
		}

		if fileio.IsExist(brokenSymlink) {
			t.Error("IsExist should return false for broken symlink")
		}
	})
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("successful copy of regular file", func(t *testing.T) {
		source := createFile(t, tempDir, "source.txt", "Hello, World!", 0644)
		dest := filepath.Join(tempDir, "dest.txt")

		err := fileio.CopyFile(dest, source, 0644)
		if err != nil {
			t.Fatalf("CopyFile failed: %v", err)
		}

		compareFileContent(t, source, dest)
		compareFileMode(t, source, dest)
	})

	t.Run("copy with different permissions", func(t *testing.T) {
		source := createFile(t, tempDir, "source_perm.txt", "test content", 0644)
		dest := filepath.Join(tempDir, "dest_perm.txt")

		err := fileio.CopyFile(dest, source, 0755)
		if err != nil {
			t.Fatalf("CopyFile failed: %v", err)
		}

		compareFileContent(t, source, dest)

		info, err := os.Stat(dest)
		if err != nil {
			t.Fatalf("failed to stat destination file: %v", err)
		}

		if info.Mode().Perm() != 0755 {
			t.Errorf("destination file permissions = %o, expected %o", info.Mode().Perm(), 0o755)
		}
	})

	t.Run("copy empty file", func(t *testing.T) {
		source := createFile(t, tempDir, "empty_source.txt", "", 0644)
		dest := filepath.Join(tempDir, "empty_dest.txt")

		err := fileio.CopyFile(dest, source, 0644)
		if err != nil {
			t.Fatalf("CopyFile failed: %v", err)
		}

		compareFileContent(t, source, dest)
	})

	t.Run("copy large file", func(t *testing.T) {
		largeContent := strings.Repeat("A", 10*1024*1024) // 10MB
		source := createFile(t, tempDir, "large_source.txt", largeContent, 0644)
		dest := filepath.Join(tempDir, "large_dest.txt")

		err := fileio.CopyFile(dest, source, 0644)
		if err != nil {
			t.Fatalf("CopyFile failed: %v", err)
		}

		compareFileContent(t, source, dest)
	})

	t.Run("non-existent source file", func(t *testing.T) {
		source := filepath.Join(tempDir, "non_existent_source.txt")
		dest := filepath.Join(tempDir, "dest_from_non_existent.txt")

		err := fileio.CopyFile(dest, source, 0644)
		if err == nil {
			t.Error("CopyFile should fail with non-existent source file")
		}
	})

	t.Run("destination already exists", func(t *testing.T) {
		source := createFile(t, tempDir, "source_existing.txt", "source content", 0644)
		dest := createFile(t, tempDir, "existing_dest.txt", "existing content", 0644)

		err := fileio.CopyFile(dest, source, 0644)
		if err == nil {
			t.Error("CopyFile should fail when destination already exists due to O_EXCL flag")
		}
	})
}

func TestCopyFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	tempDir := t.TempDir()

	t.Run("no read permission on source", func(t *testing.T) {
		source := createFile(t, tempDir, "no_read_source.txt", "content", 0000)
		dest := filepath.Join(tempDir, "dest_no_read.txt")

		err := fileio.CopyFile(dest, source, 0644)
		if err == nil {
			t.Error("CopyFile should fail with no read permission on source")
		}

		// Cleanup
		os.Chmod(source, 0644)
	})

	t.Run("no write permission on destination directory", func(t *testing.T) {
		source := createFile(t, tempDir, "source_no_write.txt", "content", 0644)
		noWriteDir := createDir(t, tempDir, "no_write_dir", 0555)
		dest := filepath.Join(noWriteDir, "dest_in_no_write.txt")

		err := fileio.CopyFile(dest, source, 0644)
		if err == nil {
			t.Error("CopyFile should fail with no write permission on destination directory")
		}

		// Cleanup
		os.Chmod(noWriteDir, 0755)
	})
}

func TestCopyDir(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("empty source directory", func(t *testing.T) {
		srcDir := createDir(t, tempDir, "empty_src", 0755)
		destDir := filepath.Join(tempDir, "empty_dest")

		err := fileio.CopyDir(srcDir, destDir)
		if err != nil {
			t.Fatalf("CopyDir failed: %v", err)
		}

		if !fileio.IsExist(destDir) {
			t.Error("destination directory should exist after copying empty directory")
		}

		verifyDirectoryStructure(t, destDir, []string{})
	})

	t.Run("single file in directory", func(t *testing.T) {
		srcDir := createDir(t, tempDir, "single_file_src", 0755)
		createFile(t, srcDir, "single.txt", "single file content", 0644)
		destDir := filepath.Join(tempDir, "single_file_dest")

		err := fileio.CopyDir(srcDir, destDir)
		if err != nil {
			t.Fatalf("CopyDir failed: %v", err)
		}

		expectedPaths := []string{"single.txt"}
		verifyDirectoryStructure(t, destDir, expectedPaths)

		srcFile := filepath.Join(srcDir, "single.txt")
		destFile := filepath.Join(destDir, "single.txt")
		compareFileContent(t, srcFile, destFile)
		compareFileMode(t, srcFile, destFile)
	})

	t.Run("nested directory structure", func(t *testing.T) {
		srcDir := createDir(t, tempDir, "nested_src", 0755)
		structure := createNestedTestStructure(t, srcDir)
		destDir := filepath.Join(tempDir, "nested_dest")

		err := fileio.CopyDir(srcDir, destDir)
		if err != nil {
			t.Fatalf("CopyDir failed: %v", err)
		}

		expectedPaths := []string{
			"dir1",
			"dir1/subdir1",
			"dir1/subdir2",
			"dir2",
			"file1.txt",
			"file2.txt",
			"dir1/nested_file.txt",
			"dir1/subdir1/deep_file.txt",
			"dir1/subdir2/private_file.txt",
		}
		verifyDirectoryStructure(t, destDir, expectedPaths)

		// Verify file contents
		for name, srcPath := range structure {
			if !strings.Contains(name, "/") && !strings.HasSuffix(name, ".txt") {
				continue // Skip directories in structure map
			}
			if strings.HasSuffix(name, ".txt") {
				srcFile := srcPath
				destFile := filepath.Join(destDir, name)
				compareFileContent(t, srcFile, destFile)
			}
		}
	})

	t.Run("non-existent source directory", func(t *testing.T) {
		srcDir := filepath.Join(tempDir, "non_existent_src")
		destDir := filepath.Join(tempDir, "dest_from_non_existent")

		err := fileio.CopyDir(srcDir, destDir)
		if err == nil {
			t.Error("CopyDir should fail with non-existent source directory")
		}
	})
}

func TestCopyDirPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	tempDir := t.TempDir()

	t.Run("no read permission on source directory", func(t *testing.T) {
		srcDir := createDir(t, tempDir, "no_read_src", 0000)
		destDir := filepath.Join(tempDir, "dest_from_no_read")

		err := fileio.CopyDir(srcDir, destDir)
		if err == nil {
			t.Error("CopyDir should fail with no read permission on source directory")
		}

		// Cleanup
		os.Chmod(srcDir, 0755)
	})
}

func TestCopyDirIntegration(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("CopyDir uses CopyFile correctly", func(t *testing.T) {
		srcDir := createDir(t, tempDir, "integration_src", 0755)
		createFile(t, srcDir, "test1.txt", "content1", 0644)
		createFile(t, srcDir, "test2.txt", "content2", 0755)
		destDir := filepath.Join(tempDir, "integration_dest")

		err := fileio.CopyDir(srcDir, destDir)
		if err != nil {
			t.Fatalf("CopyDir failed: %v", err)
		}

		// Verify that files were copied with correct permissions
		test1Src := filepath.Join(srcDir, "test1.txt")
		test1Dest := filepath.Join(destDir, "test1.txt")
		test2Src := filepath.Join(srcDir, "test2.txt")
		test2Dest := filepath.Join(destDir, "test2.txt")

		compareFileContent(t, test1Src, test1Dest)
		compareFileContent(t, test2Src, test2Dest)
		compareFileMode(t, test1Src, test1Dest)
		compareFileMode(t, test2Src, test2Dest)
	})
}

func TestEdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("filenames with special characters", func(t *testing.T) {
		specialChars := []string{
			"file with spaces.txt",
			"file-with-dashes.txt",
			"file_with_underscores.txt",
			"file.with.dots.txt",
		}

		for _, filename := range specialChars {
			t.Run(filename, func(t *testing.T) {
				source := createFile(t, tempDir, filename, "special content", 0644)
				dest := filepath.Join(tempDir, "copy_"+filename)

				err := fileio.CopyFile(dest, source, 0644)
				if err != nil {
					t.Fatalf("CopyFile failed with special filename %s: %v", filename, err)
				}

				compareFileContent(t, source, dest)
			})
		}
	})

	t.Run("very long file path", func(t *testing.T) {
		longName := strings.Repeat("a", 200) + ".txt"
		source := createFile(t, tempDir, longName, "long path content", 0644)
		dest := filepath.Join(tempDir, "copy_"+longName)

		err := fileio.CopyFile(dest, source, 0644)
		if err != nil {
			t.Fatalf("CopyFile failed with long filename: %v", err)
		}

		compareFileContent(t, source, dest)
	})
}

// Benchmark tests
func BenchmarkCopyFile(b *testing.B) {
	tempDir := b.TempDir()

	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
	}
	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			content := strings.Repeat("A", size.size)
			source := createFile(b, tempDir, "bench_source.txt", content, 0644)

			b.ResetTimer()
			for i := 0; b.Loop(); i++ {
				dest := filepath.Join(tempDir, "bench_dest_"+filepath.Base(tempDir)+"_"+string(rune(i))+".txt")
				err := fileio.CopyFile(dest, source, 0644)
				if err != nil {
					b.Fatalf("CopyFile failed: %v", err)
				}
				os.Remove(dest) // Clean up for next iteration
			}
		})
	}
}

func BenchmarkCopyDir(b *testing.B) {
	tempDir := b.TempDir()

	// Create a complex directory structure
	srcDir := createDir(b, tempDir, "bench_src", 0755)
	for i := range 10 {
		subDir := createDir(b, srcDir, "subdir_"+string(rune(i)), 0755)
		for j := range 5 {
			content := strings.Repeat("content", 100)
			createFile(b, subDir, "file_"+string(rune(j))+".txt", content, 0644)
		}
	}

	for i := 0; b.Loop(); i++ {
		destDir := filepath.Join(tempDir, "bench_dest_"+string(rune(i)))
		err := fileio.CopyDir(srcDir, destDir)
		if err != nil {
			b.Fatalf("CopyDir failed: %v", err)
		}
		os.RemoveAll(destDir) // Clean up for next iteration
	}
}
