package suite

import (
    "os"
    "strings"
)

// GetDockerImageInfo returns the appropriate repo and branch version string for integration with the CI pipeline.
// The remote runner sets the BRANCH_CI env var. If present, interchaintest will use the docker image pushed up to the repo.
// If testing locally, user should run `make local-image` and interchaintest will use the local image.
func GetDockerImageInfo() (repo, version string) {
    repo = "ghcr.io/cosmoscontracts/juno"

    // If BRANCH_CI set, sanitize and use it
    if v, ok := os.LookupEnv("BRANCH_CI"); ok {
        v = strings.ReplaceAll(v, "/", "-")
        if v != "" {
            return repo, v
        }
    }

	return repo, "local"
}
