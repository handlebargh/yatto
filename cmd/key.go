// Copyright 2025-2026 handlebargh and contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/handlebargh/yatto/internal/encryption"
	"github.com/spf13/cobra"
)

var (
	keyOutputPath string
	keyForce      bool
)

// keyCmd represents the key command
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage encryption keys",
}

// keyGenerateCmd generates a new encryption key
var keyGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new AES-256 encryption key",
	Long: `Generate a new 32-byte AES-256 encryption key and save it to a file.

The key is base64-encoded for easy storage. Keep this file secure!

Example usage:
  yatto key generate --output ~/.config/yatto/encryption.key

Then update your config.toml:
  [encryption]
  enable = true
  key_path = "/home/<you>/.config/yatto/encryption.key"
`,
	RunE: func(_ *cobra.Command, _ []string) error {
		key, err := encryption.GenerateKey()
		if err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}

		// Default output path if not specified
		if keyOutputPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			keyOutputPath = filepath.Join(home, ".config", "yatto", "encryption.key")

			// Create directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(keyOutputPath), 0o700); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
		}

		// Check if key file already exists
		if _, err := os.Stat(keyOutputPath); err == nil {
			if !keyForce {
				return fmt.Errorf(
					"key file already exists at %s. Use --force to overwrite, or remove the existing file first",
					keyOutputPath,
				)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check key file: %w", err)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(keyOutputPath), 0o700); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Write key file with restrictive permissions
		if err := os.WriteFile(keyOutputPath, []byte(key), 0o600); err != nil {
			return fmt.Errorf("failed to write key file: %w", err)
		}

		fmt.Printf("Encryption key generated and saved to: %s\n", keyOutputPath)
		fmt.Printf("\nIMPORTANT: Keep this file secure! Anyone with access to this key can decrypt your tasks.\n")
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("1. Update your config.toml:\n")
		fmt.Printf("   [encryption]\n")
		fmt.Printf("   enable = true\n")
		fmt.Printf("   key_path = \"%s\"\n", keyOutputPath)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(keyCmd)
	keyCmd.AddCommand(keyGenerateCmd)

	keyGenerateCmd.Flags().StringVarP(&keyOutputPath, "output", "o", "",
		"Output path for the encryption key file (default: ~/.config/yatto/encryption.key)")
	keyGenerateCmd.Flags().BoolVarP(&keyForce, "force", "f", false,
		"Overwrite existing key file if it exists")
}
