package common

import "regexp"

var VersionRegexp = regexp.MustCompile(`\b\d+(\.\d+){2}\b`)
