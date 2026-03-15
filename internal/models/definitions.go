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

// Package models defines the Bubble Tea-based
// TUI models for managing and interacting with
// task and project lists.
package models

import (
	"charm.land/lipgloss/v2"
)

type (
	// mode defines the state of the TUI, used for contextual behavior (e.g., normal, confirm delete, error).
	mode int

	// doneWaitingMsg signals that the spinner has finished its post-completion delay.
	doneWaitingMsg struct{}
)

const (
	// modeNormal indicates the default UI mode.
	modeNormal mode = iota

	// modeConfirmDelete indicates the UI is prompting for delete confirmation.
	modeConfirmDelete

	// modeBackendError indicates a backend-related error has occurred and should be displayed.
	modeBackendError
)

// appStyle defines the base padding for the entire application.
var appStyle = lipgloss.NewStyle().Padding(1, 2)
