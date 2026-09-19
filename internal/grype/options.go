package grype

import "time"

// Options is a single request. AsOf is supplied by the CLI, never read here.
type Options struct {
	Repository         string
	RepositoryExplicit bool
	Scan               string
	Inventory          string
	Allowlist          string
	CoverageOnly       bool
	AsOf               time.Time
}
