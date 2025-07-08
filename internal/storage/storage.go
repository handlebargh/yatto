package storage

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/handlebargh/yatto/internal/git"
	"github.com/spf13/viper"
)

func CreateStorageDir() {
	storageDir := viper.GetString("storage_dir")

	// Ask if storage directory should be created if it does not exist.
	_, err := os.Stat(storageDir)
	if os.IsNotExist(err) {
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("Create storage directory at %s? [y|N]: ", storageDir)

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("An error occured while reading input. Please try again", err)
			return
		}

		input = strings.TrimSpace(input)

		if input == "yes" || input == "y" {
			// Create storage directory.
			err := os.MkdirAll(storageDir, 0755)
			if err != nil {
				panic(fmt.Errorf("fatal error creating storage directory: %w", err))
			}

			if viper.GetBool("use_git") {
				err = git.GitInit(storageDir)
				if err != nil {
					panic(fmt.Errorf("fatal error initializing git repository: %w", err))
				}
			}
		} else {
			os.Exit(0)
		}
	}
}

func FileExists(file string) bool {
	_, err := os.Stat(viper.GetString("storage_dir") + "/" + file)
	return os.IsExist(err)
}
