package app

import (
	"fmt"
	"regexp"
)

// These constants define the application version, following
// the semantic versioning 2.0.0 specification (http://semver.org).
const (
	appMajor = 0
	appMinor = 2
	appPatch = 0

	// appPreRelease must contain only characters from the
	// semantic alphabet, as per https://semver.org/#spec-item-9.
	appPreRelease = "alpha"
)

var (
	// commit stores the current commit of this build, which includes
	// the most recent tag, the number of commits since that tag,
	// the commit hash and a dirty marker.
	// It should be set using -ldflags during compilation.
	commit string

	// commitHash stores the commit hash of this build.
	// It should be set using -ldflags during compilation.
	commitHash string
)

func init() {
	semverRe := regexp.MustCompile(`^([0-9A-Za-z\-]+(?:\.[0-9A-Za-z\-]+)*)$`)
	conformsSemVer := func(s string) bool {
		return semverRe.MatchString(s)
	}

	// Check that appPreRelease conforms to semantic versioning spec.
	if appPreRelease != "" && !conformsSemVer(appPreRelease) {
		panic(fmt.Errorf("PreRelease does not conform"+
			" to semantic versioning: %s", appPreRelease))
	}
}

// Version returns the application version as a properly formed string,
// as per the semantic versioning specification 2.0.0.
func Version() string {
	ver := fmt.Sprintf("%d.%d.%d", appMajor, appMinor, appPatch)

	// Append pre-release if there is one.
	if appPreRelease != "" {
		ver = fmt.Sprintf("%s-%s", ver, appPreRelease)
	}

	return ver
}

// BuildInfo returns the commit descriptor as well as
// the commit hash used in the current c13n build.
func BuildInfo() (commitDesc string, hash string) {
	return commit, commitHash
}
