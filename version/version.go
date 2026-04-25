package version

// These get overwritten at link time via -ldflags "-X ..."
var (
	Tag    = "dev"
	Commit = "unknown"
	Date   = "unknown"
)
