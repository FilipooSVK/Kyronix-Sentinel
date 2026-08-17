package version

// Build information.
// These values can be overridden during build using -ldflags.
var (
	Version   = "0.1.0"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Info returns formatted version information.
type Info struct {
	Version   string
	Commit    string
	BuildDate string
}

// Current returns current build information.
func Current() Info {

	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
}
