package suite

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	evidencetypes "cosmossdk.io/x/evidence/types"

	"cosmossdk.io/x/feegrant"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	comettypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1types "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	nft "cosmossdk.io/x/nft"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	clocktypes "github.com/CosmosContracts/juno/v30/x/clock/types"
	cwhooktypes "github.com/CosmosContracts/juno/v30/x/cw-hooks/types"
	driptypes "github.com/CosmosContracts/juno/v30/x/drip/types"
	feepaytypes "github.com/CosmosContracts/juno/v30/x/feepay/types"
	feesharetypes "github.com/CosmosContracts/juno/v30/x/feeshare/types"
	tokenfactorytypes "github.com/CosmosContracts/juno/v30/x/tokenfactory/types"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	feemarkettypes "github.com/CosmosContracts/juno/v30/x/feemarket/types"
	minttypes "github.com/CosmosContracts/juno/v30/x/mint/types"
)

// E2ETestSuite runs the feemarket e2e test-suite against a given interchaintest specification
type E2ETestSuite struct {
	suite.Suite
	QueryClients
	// our chain spec
	Spec *interchaintest.ChainSpec
	// our main chain
	Chain *cosmos.CosmosChain
	// pregenerated and funded users
	User1, User2, User3 ibc.Wallet
	// app codec
	Cdc codec.Codec
	// app context
	Ctx context.Context
	// default token denom
	Denom string
	// default gas prices
	GasPrices string
	// authority address (gov module address)
	Authority sdk.AccAddress
	// block time
	BlockTime time.Duration
	// overrides for key-ring configuration of the broadcaster
	BroadcasterOverrides *KeyringOverride
	// default broadcaster
	Bc *cosmos.Broadcaster
	// interchain constructor
	Icc InterchainConstructor
	// interchain
	Ic interchaintest.Interchain
	// chain constructor
	Cc ChainConstructor
	// txConfig controls the tx configuration for each test
	TxConfig TestTxConfig
	// grpc client
	GrpcClient *grpc.ClientConn
}

func NewE2ETestSuite(spec *interchaintest.ChainSpec, txCfg TestTxConfig, opts ...Option) *E2ETestSuite {
	if err := txCfg.Validate(); err != nil {
		panic(err)
	}

	ctx := context.Background()

	suite := &E2ETestSuite{
		Spec:      spec,
		Ctx:       ctx,
		Denom:     Denom,
		GasPrices: "",
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName),
		Icc:       DefaultInterchainConstructor,
		Cc:        DefaultChainConstructor,
		TxConfig:  txCfg,
	}

	for _, opt := range opts {
		opt(suite)
	}

	// get grpc address
	grpcAddr := suite.Chain.GetHostGRPCAddress()
	grpcClient, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	suite.Require().NoError(err)

	// set grpc client
	suite.GrpcClient = grpcClient
	suite.setupQueryClients()

	return suite
}

type QueryClients struct {
	AuthClient         authtypes.QueryClient
	AuthzClient        authz.QueryClient
	BankClient         banktypes.QueryClient
	ConsensusClient    consensustypes.QueryClient
	DistributionClient distributiontypes.QueryClient
	EvidenceClient     evidencetypes.QueryClient
	FeegrantClient     feegrant.QueryClient
	GovClient          govv1types.QueryClient
	NftClient          nft.QueryClient
	SlashingClient     slashingtypes.QueryClient
	StakingClient      stakingtypes.QueryClient
	UpgradeClient      upgradetypes.QueryClient
	MintClient         minttypes.QueryClient
	ClockClient        clocktypes.QueryClient
	CwhooksClient      cwhooktypes.QueryClient
	DripClient         driptypes.QueryClient
	FeemarketClient    feemarkettypes.QueryClient
	FeepayClient       feepaytypes.QueryClient
	FeeShareClient     feesharetypes.QueryClient
	TokenfactoryClient tokenfactorytypes.QueryClient
}

func (s *E2ETestSuite) setupQueryClients() {
	authClient := authtypes.NewQueryClient(s.GrpcClient)
	s.AuthClient = authClient
	authzClient := authz.NewQueryClient(s.GrpcClient)
	s.AuthzClient = authzClient
	bankClient := banktypes.NewQueryClient(s.GrpcClient)
	s.BankClient = bankClient
	consensusClient := consensustypes.NewQueryClient(s.GrpcClient)
	s.ConsensusClient = consensusClient
	distributionClient := distributiontypes.NewQueryClient(s.GrpcClient)
	s.DistributionClient = distributionClient
	evidenceClient := evidencetypes.NewQueryClient(s.GrpcClient)
	s.EvidenceClient = evidenceClient
	feegrantClient := feegrant.NewQueryClient(s.GrpcClient)
	s.FeegrantClient = feegrantClient
	govClient := govv1types.NewQueryClient(s.GrpcClient)
	s.GovClient = govClient
	nftClient := nft.NewQueryClient(s.GrpcClient)
	s.NftClient = nftClient
	slashingClient := slashingtypes.NewQueryClient(s.GrpcClient)
	s.SlashingClient = slashingClient
	stakingClient := stakingtypes.NewQueryClient(s.GrpcClient)
	s.StakingClient = stakingClient
	upgradeClient := upgradetypes.NewQueryClient(s.GrpcClient)
	s.UpgradeClient = upgradeClient
	mintClient := minttypes.NewQueryClient(s.GrpcClient)
	s.MintClient = mintClient
	clockClient := clocktypes.NewQueryClient(s.GrpcClient)
	s.ClockClient = clockClient
	cwhooksClient := cwhooktypes.NewQueryClient(s.GrpcClient)
	s.CwhooksClient = cwhooksClient
	dripClient := driptypes.NewQueryClient(s.GrpcClient)
	s.DripClient = dripClient
	feemarketClient := feemarkettypes.NewQueryClient(s.GrpcClient)
	s.FeemarketClient = feemarketClient
	feepayClient := feepaytypes.NewQueryClient(s.GrpcClient)
	s.FeepayClient = feepayClient
	feeShareClient := feesharetypes.NewQueryClient(s.GrpcClient)
	s.FeeShareClient = feeShareClient
	tokenfactoryClient := tokenfactorytypes.NewQueryClient(s.GrpcClient)
	s.TokenfactoryClient = tokenfactoryClient
}

// Option is a function that modifies the E2ETestSuite
type Option func(*E2ETestSuite)

// WithDenom sets the token denom
func WithDenom(denom string) Option {
	return func(s *E2ETestSuite) {
		s.Denom = denom
	}
}

// WithGasPrices sets gas prices.
func WithGasPrices(gasPrices string) Option {
	return func(s *E2ETestSuite) {
		s.GasPrices = gasPrices
	}
}

// WithAuthority sets the authority address
func WithAuthority(addr sdk.AccAddress) Option {
	return func(s *E2ETestSuite) {
		s.Authority = addr
	}
}

// WithBlockTime sets the block time
func WithBlockTime(t time.Duration) Option {
	return func(s *E2ETestSuite) {
		s.BlockTime = t
	}
}

// WithInterchainConstructor sets the interchain constructor
func WithInterchainConstructor(ic InterchainConstructor) Option {
	return func(s *E2ETestSuite) {
		s.Icc = ic
	}
}

// WithChainConstructor sets the chain constructor
func WithChainConstructor(cc ChainConstructor) Option {
	return func(s *E2ETestSuite) {
		s.Cc = cc
	}
}

func (s *E2ETestSuite) WithKeyringOptions(cdc codec.Codec, opts keyring.Option) {
	s.BroadcasterOverrides = &KeyringOverride{
		cdc:            cdc,
		keyringOptions: opts,
	}
}

func (s *E2ETestSuite) TearDownSuite() {
	defer s.Teardown()
	if ok := os.Getenv(EnvKeepAlive); ok == "" {
		return
	}

	// await on a signal to keep the chain running
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s.T().Log("Keeping the chain running")
	<-sig
}

func (s *E2ETestSuite) Teardown() {
	// stop all nodes + sidecars in the chain
	ctx := context.Background()
	if s.Chain == nil {
		return
	}

	_ = s.Chain.StopAllNodes(ctx)
	_ = s.Chain.StopAllSidecars(ctx)
}

// WaitForHeight waits for the chain to reach the given height
func (s *E2ETestSuite) WaitForHeight(chain *cosmos.CosmosChain, height int64) {
	s.T().Helper()

	// wait for next height
	err := testutil.WaitForCondition(30*time.Second, 100*time.Millisecond, func() (bool, error) {
		pollHeight, err := chain.Height(context.Background())
		if err != nil {
			return false, err
		}
		return pollHeight >= height, nil
	})
	s.Require().NoError(err)
}

// VerifyBlock takes a Block and verifies that it contains the given bid at the 0-th index, and the bundled txs immediately after
func (s *E2ETestSuite) VerifyBlock(block *coretypes.ResultBlock, offset int, bidTxHash string, txs [][]byte) {
	s.T().Helper()

	// verify the block
	if bidTxHash != "" {
		s.Require().Equal(bidTxHash, TxHash(block.Block.Data.Txs[offset+1]))
		offset += 1
	}

	// verify the txs in sequence
	for i, tx := range txs {
		s.Require().Equal(TxHash(tx), TxHash(block.Block.Data.Txs[i+offset+1]))
	}
}

// VerifyBlockWithExpectedBlock takes in a list of raw tx bytes and compares each tx hash to the tx hashes in the block.
// The expected block is the block that should be returned by the chain at the given height.
func (s *E2ETestSuite) VerifyBlockWithExpectedBlock(chain *cosmos.CosmosChain, height uint64, txs [][]byte) {
	s.T().Helper()

	block := s.QueryBlock(chain, int64(height))
	blockTxs := block.Block.Data.Txs[1:]

	s.T().Logf("verifying block %d", height)
	s.Require().Equal(len(txs), len(blockTxs))
	for i, tx := range txs {
		s.T().Logf("verifying tx %d; expected %s, got %s", i, TxHash(tx), TxHash(blockTxs[i]))
		s.Require().Equal(TxHash(tx), TxHash(blockTxs[i]))
	}
}

func TxHash(tx []byte) string {
	return strings.ToUpper(hex.EncodeToString(comettypes.Tx(tx).Hash()))
}

func (s *E2ETestSuite) setupBroadcaster() {
	s.T().Helper()

	bc := cosmos.NewBroadcaster(s.T(), s.Chain)

	if s.BroadcasterOverrides == nil {
		s.Bc = bc
		return
	}

	// get the key-ring-dir from the node locally
	keyringDir := s.keyringDirFromNode()

	// create a new keyring
	kr, err := keyring.New("", keyring.BackendTest, keyringDir, os.Stdin, s.BroadcasterOverrides.cdc, s.BroadcasterOverrides.keyringOptions)
	s.Require().NoError(err)

	// override factory + client context keyrings
	bc.ConfigureFactoryOptions(
		func(factory tx.Factory) tx.Factory {
			return factory.WithKeybase(kr)
		},
	)
	bc.ConfigureClientContextOptions(
		func(cc client.Context) client.Context {
			return cc.WithKeyring(kr)
		},
	)

	s.Bc = bc
}

// sniped from here: https://github.com/strangelove-ventures/interchaintest ref: 9341b001214d26be420f1ca1ab0f15bad17faee6
func (s *E2ETestSuite) keyringDirFromNode() string {
	node := s.Chain.Nodes()[0]

	localDir := s.T().TempDir()

	containerKeyringDir := path.Join(node.HomeDir(), "keyring-test")
	reader, _, err := node.DockerClient.CopyFromContainer(context.Background(), node.ContainerID(), containerKeyringDir)
	s.Require().NoError(err)

	s.Require().NoError(os.Mkdir(path.Join(localDir, "keyring-test"), os.ModePerm))

	tr := tar.NewReader(reader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		s.Require().NoError(err)

		var fileBuff bytes.Buffer
		_, err = io.Copy(&fileBuff, tr)
		s.Require().NoError(err)

		name := hdr.Name
		extractedFileName := path.Base(name)
		isDirectory := extractedFileName == ""
		if isDirectory {
			continue
		}

		filePath := path.Join(localDir, "keyring-test", extractedFileName)
		s.Require().NoError(os.WriteFile(filePath, fileBuff.Bytes(), os.ModePerm))
	}

	return localDir
}
