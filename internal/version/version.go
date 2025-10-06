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

package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	// revision saves the git commit the application is built from.
	revision = "unknown"

	// revisionDate saves the commit's date.
	revisionDate = "unknown"

	// goVersion saves the Go version the application is built with.
	goVersion = runtime.Version()
)

// Info returns the version, commit sha hash and commit date
// from which the application is built.
func Info() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "Unable to read version information."
	}

	if buildInfo.Main.Version != "" {
		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				revision = setting.Value
			case "vcs.time":
				revisionDate = setting.Value
			case "vcs.modified":
				if setting.Value == "true" {
					revision += "+dirty"
				}
			}
		}

		return fmt.Sprintf("Version:\t%s\nRevision:\t%s\nRevisionDate:\t%s\nGoVersion:\t%s\n",
			buildInfo.Main.Version, revision, revisionDate, goVersion)
	}

	return fmt.Sprintf(
		"Version:\tunknown\nRevision:\tunknown\nRevisionDate:\tunknown\nGoVersion:\t%s\n",
		goVersion,
	)
}

// Header returns the stylized application name
// and project URL.
func Header() string {
	return `
 ____ ____ ____ ____ ____ 
||y |||a |||t |||t |||o ||
||__|||__|||__|||__|||__||
|/__\|/__\|/__\|/__\|/__\|

https://github.com/handlebargh/yatto
`
}
