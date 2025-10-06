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
