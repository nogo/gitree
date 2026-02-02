package version

// Version and GitCommit are set via ldflags at build time.
// Example: go build -ldflags "-X github.com/nogo/gitree/internal/version.Version=v0.4.0 -X github.com/nogo/gitree/internal/version.GitCommit=$(git rev-parse --short HEAD)"
var (
	Version   = "dev"
	GitCommit = "unknown"
)

// String returns a formatted version string.
func String() string {
	return Version + " (" + GitCommit + ")"
}
