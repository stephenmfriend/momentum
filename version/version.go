// Package version provides version information for the momentum CLI.
// These values are set at build time using ldflags.
package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// Build-time variables set via ldflags
var (
	// Version is the semantic version (e.g., "1.2.3")
	Version = "dev"

	// Commit is the git commit SHA
	Commit = "none"

	// Date is the build date
	Date = "unknown"
)

// Info returns a formatted version string
func Info() string {
	return fmt.Sprintf("momentum %s (commit: %s, built: %s, go: %s)",
		Version, Commit, Date, runtime.Version())
}

// Short returns just the version number
func Short() string {
	return Version
}

// githubRelease represents a GitHub release response
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckForUpdate checks GitHub for the latest release version.
// Returns the latest version string and true if an update is available.
func CheckForUpdate() (latestVersion string, updateAvailable bool) {
	if Version == "dev" {
		return "", false
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/sirsjg/momentum/releases/latest")
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", false
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(Version, "v")

	if latest != current && compareVersions(latest, current) > 0 {
		return latest, true
	}

	return latest, false
}

// compareVersions compares two semver strings.
// Returns 1 if a > b, -1 if a < b, 0 if equal.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		var aNum, bNum int
		fmt.Sscanf(aParts[i], "%d", &aNum)
		fmt.Sscanf(bParts[i], "%d", &bNum)

		if aNum > bNum {
			return 1
		} else if aNum < bNum {
			return -1
		}
	}

	if len(aParts) > len(bParts) {
		return 1
	} else if len(aParts) < len(bParts) {
		return -1
	}

	return 0
}
