package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/handlebargh/yatto/internal/git"
	"github.com/spf13/viper"
)

type Task struct {
	TaskId          string `json:"id"`
	TaskTitle       string `json:"title"`
	TaskDescription string `json:"description"`
	TaskPriority    string `json:"priority"`
}

func (t Task) Id() string          { return t.TaskId }
func (t Task) Title() string       { return t.TaskTitle }
func (t Task) Description() string { return t.TaskDescription }
func (t Task) Priority() string    { return t.TaskPriority }
func (t Task) FilterValue() string { return t.TaskTitle }

// Function to convert priority to a numerical value for sorting.
func PriorityValue(priority string) int {
	switch priority {
	case "high":
		return 2
	case "medium":
		return 1
	case "low":
		return 0
	default:
		return -1
	}
}

// ReadTasksFromFS reads all tasks from the storage directory
// and returns them as a slice of Task.
func ReadTasksFromFS() []Task {
	taskFiles, err := os.ReadDir(viper.GetString("storage_dir"))
	if err != nil {
		panic(fmt.Errorf("fatal error reading storage directory: %w", err))
	}

	var tasks []Task
	for _, entry := range taskFiles {
		var fileContent []byte
		if !entry.IsDir() {
			fileContent, err = os.ReadFile(filepath.Join(viper.GetString("storage_dir"), entry.Name()))
			if err != nil {
				panic(err)
			}
		} else {
			continue
		}

		var task Task
		if err := json.Unmarshal(fileContent, &task); err != nil {
			panic(err)
		}
		tasks = append(tasks, task)
	}

	return tasks
}

func MarshalTask(uuid, title, description, priority string) []byte {
	var task Task
	task.TaskId = uuid
	task.TaskTitle = title
	task.TaskDescription = description
	task.TaskPriority = priority

	json, err := json.MarshalIndent(task, "", "\t")
	if err != nil {
		panic(err)
	}

	return json
}

func WriteJson(json []byte, task Task, message string) error {
	file := viper.GetString("storage_dir") + "/" + task.Id()

	err := os.WriteFile(file, json, 0600)
	if err != nil {
		panic(err)
	}

	if viper.GetBool("use_git") {
		return git.GitCommit(file, viper.GetString("storage_dir"), message, false)
	}

	return nil
}

func DeleteTaskFromFS(task *Task, message string) error {
	file := viper.GetString("storage_dir") + "/" + task.Id()

	err := os.Remove(file)
	if err != nil {
		panic(err)
	}

	if viper.GetBool("use_git") {
		return git.GitCommit(file, viper.GetString("storage_dir"), message, false)
	}

	return nil
}
