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

package items

import (
	"testing"
	"time"
)

func TestIsToday(t *testing.T) {
	t.Run("returns true for today's date", func(t *testing.T) {
		today := time.Now()
		if !IsToday(&today) {
			t.Error("expected IsToday to return true for today's date")
		}
	})

	t.Run("returns false for a nil time", func(t *testing.T) {
		if IsToday(nil) {
			t.Error("expected IsToday to return false for a nil time")
		}
	})

	t.Run("returns false for a different date", func(t *testing.T) {
		yesterday := time.Now().AddDate(0, 0, -1)
		if IsToday(&yesterday) {
			t.Error("expected IsToday to return false for a different date")
		}
	})
}

func TestTaskFilterFunc(t *testing.T) {
	targets := []string{
		"first task with high priority",
		"second task with low priority",
		"third task with medium priority",
	}

	t.Run("returns all items for an empty search term", func(t *testing.T) {
		ranks := TaskFilterFunc("", targets)
		if len(ranks) != len(targets) {
			t.Errorf("expected %d ranks, but got %d", len(targets), len(ranks))
		}
	})

	t.Run("returns only matching items", func(t *testing.T) {
		ranks := TaskFilterFunc("high", targets)
		if len(ranks) != 1 {
			t.Errorf("expected 1 rank, but got %d", len(ranks))
		}
		if ranks[0].Index != 0 {
			t.Errorf("expected index 0, but got %d", ranks[0].Index)
		}
	})

	t.Run("returns items matching all tokens", func(t *testing.T) {
		ranks := TaskFilterFunc("task high", targets)
		if len(ranks) != 1 {
			t.Errorf("expected 1 rank, but got %d", len(ranks))
		}
		if ranks[0].Index != 0 {
			t.Errorf("expected index 0, but got %d", ranks[0].Index)
		}
	})

	t.Run("returns no items if not all tokens match", func(t *testing.T) {
		ranks := TaskFilterFunc("task unknown", targets)
		if len(ranks) != 0 {
			t.Errorf("expected 0 ranks, but got %d", len(ranks))
		}
	})
}
