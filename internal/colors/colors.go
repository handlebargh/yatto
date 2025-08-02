// Copyright 2025 handlebargh
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

// package colors defines all color values
// used by the application.
package colors

import "github.com/charmbracelet/lipgloss"

var (
	Red      = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	VividRed = lipgloss.AdaptiveColor{Light: "#FE134D", Dark: "#FE134D"}
	Indigo   = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	Green    = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	Orange   = lipgloss.AdaptiveColor{Light: "#FFB733", Dark: "#FFA336"}
	Blue     = lipgloss.AdaptiveColor{Light: "#1e90ff", Dark: "#1e90ff"}
	Yellow   = lipgloss.AdaptiveColor{Light: "#CCCC00", Dark: "#CCCC00"}
	Black    = lipgloss.Color("#000000")
)
