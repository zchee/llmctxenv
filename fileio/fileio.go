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

package fileio

import (
	"encoding/hex"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/bytedance/gg/gstd/gsync"
	sha256simd "github.com/minio/sha256-simd"
)

var hashPool = gsync.Pool[hash.Hash]{
	New: func() hash.Hash { return sha256simd.New() },
}

// digestPool provides fixed-size buffers (32 bytes) for SHA-256 sums to avoid
// per-call allocations from [hash.Hash.Sum].
var digestPool = gsync.Pool[*[]byte]{
	New: func() *[]byte { b := make([]byte, 0, sha256simd.Size); return &b },
}

// hexPool provides fixed-size buffers (64 bytes) for hex-encoded output.
var hexPool = gsync.Pool[*[]byte]{
	New: func() *[]byte { b := make([]byte, sha256simd.Size*2); return &b },
}

// copyBufPool provides reusable buffers for [io.CopyBuffer] to minimize
// allocations during file hashing. 32 KiB matches [io.Copy] default size.
var copyBufPool = gsync.Pool[*[]byte]{
	New: func() *[]byte { b := make([]byte, 32*1024); return &b },
}

// HashFile computes the SHA-256 hash of the file at the given path.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := hashPool.Get()
	defer hashPool.Put(h)
	h.Reset()

	// Use a pooled copy buffer to reduce allocations.
	buf := copyBufPool.Get()
	if _, err := io.CopyBuffer(h, f, *buf); err != nil {
		copyBufPool.Put(buf)
		return "", err
	}
	copyBufPool.Put(buf)

	// Compute the sum into a pooled 32-byte buffer to avoid allocation.
	digest := digestPool.Get()
	sum := h.Sum((*digest)[:0])

	// Hex-encode into a pooled 64-byte buffer.
	hexbuf := hexPool.Get()
	n := hex.Encode((*hexbuf)[:], sum)

	// Create an immutable string by copying the bytes.
	out := string((*hexbuf)[:n])

	// Return buffers to pools.
	hexPool.Put(hexbuf)
	digestPool.Put(digest)

	return out, nil
}

// IsExist reports whether the given path exists.
func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// CopyFile copies a file from source to destination with the specified permissions.
func CopyFile(dest, source string, perm os.FileMode) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	if closeErr := dst.Close(); err == nil {
		err = closeErr
	}
	return err
}

// CopyDir recursively copies a directory from srcDir to destDir.
func CopyDir(srcDir, destDir string) error {
	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		src := filepath.Join(srcDir, entry.Name())
		dest := filepath.Join(destDir, entry.Name())

		fileInfo, err := os.Stat(src)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := os.MkdirAll(dest, 0755); err != nil {
				return err
			}
			if err := CopyDir(src, dest); err != nil {
				return err
			}
		default:
			// Ensure parent directory exists for the destination file
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				return err
			}
			if err := CopyFile(dest, src, fileInfo.Mode().Perm()); err != nil {
				return err
			}
		}
	}
	return nil
}
