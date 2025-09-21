// Copyright 2025 handlebargh and contributors
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

package vcs

import "errors"

// ErrorNoInit is returned when a (jj) git pull command is executed
// in a non-initialized storage repository.
// TODO: account for jj
var ErrorNoInit = errors.New(
	"trying to pull but local repo is not initialized.\nPlease disable git.remote and try again",
)

type (
	// InitDoneMsg is returned when repo initialization completes successfully.
	InitDoneMsg struct{}

	// InitErrorMsg is returned when repo initialization fails.
	InitErrorMsg struct{ Err error }

	// CommitDoneMsg is returned when a commit completes successfully.
	CommitDoneMsg struct{}

	// CommitErrorMsg is returned when a commit fails.
	CommitErrorMsg struct{ Err error }

	// PullDoneMsg is returned when a pull/fetch operation completes successfully.
	PullDoneMsg struct{}

	// PullErrorMsg is returned when a pull/fetch operation fails.
	PullErrorMsg struct{ Err error }

	// PullNoInitMsg is returned when a pull/fetch operation didn't run
	// because the repository's INIT file is missing.
	PullNoInitMsg struct{}

	// PushErrorMsg is returned when a push operation fails.
	PushErrorMsg struct{ Err error }
)

// Error implements the error interface for InitErrorMsg.
func (e InitErrorMsg) Error() string { return e.Err.Error() }

// Error implements the error interface for CommitErrorMsg.
func (e CommitErrorMsg) Error() string { return e.Err.Error() }

// Error implements the error interface for PullErrorMsg.
func (e PullErrorMsg) Error() string { return e.Err.Error() }

// Error implements the error interface for PushErrorMsg.
func (e PushErrorMsg) Error() string { return e.Err.Error() }
