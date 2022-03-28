package config

import "fmt"

/*
Make sure to update version numbers in these locations also:

- .github/*

*/

const (
	// Version specifies the verion of the API and its structs
	Version = "v1"

	// MajorVersion of the API
	MajorVersion = 0
	// MinorVersion of the API
	MinorVersion = 8
	// FixVersion of the API
	FixVersion = 0

	CmdLineName     = "po"
	ProjectName     = "podops"
	CopyrightString = "podops copyright 2022"
)

var (
	// VersionString is the canonical API description
	VersionString string = fmt.Sprintf("%d.%d.%d", MajorVersion, MinorVersion, FixVersion)
	// UserAgentString identifies any http request podops makes
	UserAgentString string = fmt.Sprintf("podops %d.%d.%d", MajorVersion, MinorVersion, FixVersion)
	// ServerString identifies the content server
	ServerString string = fmt.Sprintf("podops cdn %d.%d.%d", MajorVersion, MinorVersion, FixVersion)
)
