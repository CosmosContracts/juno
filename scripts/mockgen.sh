set -eo pipefail

install_mockgen() {
   go install go.uber.org/mock/mockgen@latest
}

install_mockgen

mockgen_cmd="mockgen"
$mockgen_cmd -source=x/clock/types/expected_keepers.go -package mock -destination x/clock/testutil/expected_keepers_mocks.go