package main

import (
	"github.com/stretchr/testify/suite"

	configurer "github.com/CosmosContracts/juno/v12/tests/e2e/configurer"
)

type IntegrationTestSuite struct {
	suite.Suite

	configurer    configurer.Configurer
	skipUpgrade   bool
	skipIBC       bool
	skipStateSync bool
	// forkHeight    int
}
