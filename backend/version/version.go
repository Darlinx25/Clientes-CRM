// Package version holds the application version. Release builds override
// Version at link time with -X meerkat/version.Version=<version>; local and
// test builds keep the fallback below.
package version

// Version is the user-visible application version.
var Version = "0.5.0"
