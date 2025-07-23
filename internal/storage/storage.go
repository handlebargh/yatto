package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func CreateStorageDir() {
	storageDir := viper.GetString("storage.path")

	// Ask if storage directory should be created if it does not exist.
	_, err := os.Stat(storageDir)
	if os.IsNotExist(err) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Create storage directory at %s? [y|N]: ", storageDir)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("An error occurred while reading input. Please try again", err)
			return
		}

		input = strings.TrimSpace(input)

		if input == "yes" || input == "y" {
			// Create storage directory.
			err := os.MkdirAll(storageDir, 0700)
			if err != nil {
				panic(fmt.Errorf("fatal error creating storage directory: %w", err))
			}
		} else {
			os.Exit(0)
		}
	}
}

func FileExists(file string) bool {
	fullPath := filepath.Join(viper.GetString("storage.path"), file)
	_, err := os.Stat(fullPath)
	return !os.IsNotExist(err)
}
