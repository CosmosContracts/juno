# run test with a longer timeout

export JUNO_E2E_SKIP_UPGRADE=false
export JUNO_E2E_SKIP_IBC=false
export JUNO_E2E_SKIP_STATE_SYNC=true

docker rm -f $(docker ps -a -q)
/usr/bin/go test -timeout 6000s -run ^TestIntegrationTestSuite$ github.com/CosmosContracts/juno/v12/tests/e2e -v