package helpers

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/handlebargh/yatto/internal/colors"
	"github.com/stretchr/testify/assert"
)

func TestLabelsStringToSlice(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string(nil),
		},
		{
			name:     "single label",
			input:    "work",
			expected: []string{"work"},
		},
		{
			name:     "multiple labels",
			input:    "work,urgent,home",
			expected: []string{"work", "urgent", "home"},
		},
		{
			name:     "labels with spaces",
			input:    " work , urgent,  home  ",
			expected: []string{"work", "urgent", "home"},
		},
		{
			name:     "labels with empty entries",
			input:    "work,,home",
			expected: []string{"work", "home"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := LabelsStringToSlice(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetColorCode(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected lipgloss.AdaptiveColor
	}{
		{"green", "green", colors.Green()},
		{"orange", "orange", colors.Orange()},
		{"red", "red", colors.Red()},
		{"blue", "blue", colors.Blue()},
		{"indigo", "indigo", colors.Indigo()},
		{"unknown", "unknown", colors.Blue()},
		{"empty", "", colors.Blue()},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetColorCode(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestPromptUser(t *testing.T) {
	t.Run("returns user input when no expected values are given", func(t *testing.T) {
		input := bytes.NewBufferString("  user input  \n")
		output := &bytes.Buffer{}
		message := "Enter something: "

		result, err := PromptUser(input, output, message)

		assert.NoError(t, err)
		assert.Equal(t, "user input", result)
		assert.Equal(t, message, output.String())
	})

	t.Run("returns user input when it matches an expected value", func(t *testing.T) {
		input := bytes.NewBufferString("yes\n")
		output := &bytes.Buffer{}
		message := "Are you sure? (yes/no): "
		expected := []string{"yes", "no"}

		result, err := PromptUser(input, output, message, expected...)

		assert.NoError(t, err)
		assert.Equal(t, "yes", result)
	})

	t.Run("returns error when user input does not match any expected value", func(t *testing.T) {
		input := bytes.NewBufferString("maybe\n")
		output := &bytes.Buffer{}
		message := "Are you sure? (yes/no): "
		expected := []string{"yes", "no"}

		_, err := PromptUser(input, output, message, expected...)

		assert.ErrorIs(t, err, ErrUnexpectedInput)
	})

	t.Run("handles input with extra whitespace", func(t *testing.T) {
		input := bytes.NewBufferString("  no  \n")
		output := &bytes.Buffer{}
		message := "Are you sure? (yes/no): "
		expected := []string{"yes", "no"}

		result, err := PromptUser(input, output, message, expected...)

		assert.NoError(t, err)
		assert.Equal(t, "no", result)
	})

	t.Run("handles case-insensitive input", func(t *testing.T) {
		input := bytes.NewBufferString("YES\n")
		output := &bytes.Buffer{}
		message := "Are you sure? (yes/no): "
		expected := []string{"yes", "no"}

		result, err := PromptUser(input, output, message, expected...)

		assert.NoError(t, err)
		assert.Equal(t, "yes", result)
	})
}

func TestUniqueNonEmptyStrings(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "all unique",
			input:    []string{"one", "two", "three"},
			expected: []string{"one", "two", "three"},
		},
		{
			name:     "with duplicates",
			input:    []string{"one", "two", "one"},
			expected: []string{"one", "two"},
		},
		{
			name:     "with empty strings",
			input:    []string{"one", "", "two", " "},
			expected: []string{"one", "two"},
		},
		{
			name:     "with emails",
			input:    []string{"user1@example.com", "User 2 <user2@example.com>", "user1@example.com"},
			expected: []string{"<user1@example.com>", "User 2 <user2@example.com>"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := UniqueNonEmptyStrings(tc.input)
			assert.ElementsMatch(t, tc.expected, result)
		})
	}
}

func TestAddAngleBracketsToEmail(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple email",
			input:    "user@example.com",
			expected: "<user@example.com>",
		},
		{
			name:     "email with name",
			input:    "John Doe <user@example.com>",
			expected: "John Doe <user@example.com>",
		},
		{
			name:     "email already bracketed",
			input:    "<user@example.com>",
			expected: "<user@example.com>",
		},
		{
			name:     "string with no email",
			input:    "just a string",
			expected: "just a string",
		},
		{
			name:     "name and email no brackets",
			input:    "Jane Doe user@example.com",
			expected: "Jane Doe <user@example.com>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := AddAngleBracketsToEmail(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
