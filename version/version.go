// Package version provides build version information.
package version

import (
	"fmt"
	"runtime"
)

// The following fields are populated at build time using -ldflags -X.
// Note that DATE is omitted for reproducible builds
var (
	buildVersion     = "unknown"
	buildGitRevision = "unknown"
	buildUser        = "unknown"
	buildHost        = "unknown"
	buildTime        = "unknown"
	buildStatus      = "unknown"
)

// BuildInfo describes version information about the binary build.
type BuildInfo struct {
	Version       string `json:"version"`
	GitRevision   string `json:"revision"`
	User          string `json:"user"`
	Host          string `json:"host"`
	GolangVersion string `json:"golang_version"`
	BuildTime     string `json:"build_time"`
	BuildStatus   string `json:"status"`
}

var (
	// Info exports the build version information.
	Info BuildInfo
)

// String produces a single-line version info
//
// This looks like:
//
// ```
// user@host-<build time>-<version>-<git revision>-<build status>
// ```
func (b BuildInfo) String() string {
	return fmt.Sprintf("%v@%v-%v-%v-%v-%v",
		b.User,
		b.Host,
		b.BuildTime,
		b.Version,
		b.GitRevision,
		b.BuildStatus)
}

// LongForm returns a multi-line version information
//
// This looks like:
//
// ```
// Version: <version>
// GitRevision: <git revision>
// User: user@host
// BuildTime: <build time>
// GolangVersion: go1.9.2
// BuildStatus: <build status>
// ```
func (b BuildInfo) LongForm() string {
	return fmt.Sprintf(`Version: %v
GitRevision: %v
User: %v@%v
BuildTime: %v
GolangVersion: %v
BuildStatus: %v
`,
		b.Version,
		b.GitRevision,
		b.User,
		b.Host,
		b.BuildTime,
		b.GolangVersion,
		b.BuildStatus)
}

func init() {
	Info = BuildInfo{
		Version:       buildVersion,
		GitRevision:   buildGitRevision,
		User:          buildUser,
		Host:          buildHost,
		GolangVersion: runtime.Version(),
		BuildTime:     buildTime,
		BuildStatus:   buildStatus,
	}
}
