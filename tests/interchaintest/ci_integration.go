package interchaintest

import (
	"os"
)

// GetDockerImageInfo returns the appropriate repo and branch version string for integration with the CI pipeline.
// The remote runner sets the BRANCH_CI env var. If present, interchaintest will use the docker image pushed up to the repo.
// If testing locally, user should run `make local-image` and interchaintest will use the local image.
func GetDockerImageInfo() (repo, version string) {
	branchVersion, found := os.LookupEnv("BRANCH_CI")
	repo = "ghcr.io/CosmosContracts/juno"
	if !found {
		repo = "juno"
		branchVersion = "local"
	}
	return repo, branchVersion
}
