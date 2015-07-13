// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package tools

import (
	"github.com/juju/loggo"

	"github.com/juju/juju/version"
)

var logger = loggo.GetLogger("juju.tools")

// These should be used as FileType in the struct below.
// TODO(ZIP): consider making a separate type for them
const (
	Tgz = "tgz"
	Zip = "zip"
)

// Tools represents the location and version of a tools tarball.
type Tools struct {
	Version  version.Binary `json:"version"`
	URL      string         `json:"url"`
	SHA256   string         `json:"sha256,omitempty"`
	Size     int64          `json:"size"`
	FileType string         `json:"filetype"`
}

// UseZipToolsWindows returns whether we should use zip tools on windows based on version.Binary.
// From version 1.26 we're switching to providing tools archived in zip on Windows.
func UseZipToolsWindows(vers version.Binary) bool {
	if vers.OS == version.Windows && (vers.Major > 1 || (vers.Major == 1 && vers.Minor >= 26)) {
		return true
	}
	return false
}
