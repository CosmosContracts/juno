package config

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/CosmosContracts/juno/price-feeder/oracle/provider"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

const (
	DenomUSD = "USD"

	defaultListenAddr      = "0.0.0.0:7171"
	defaultSrvWriteTimeout = 15 * time.Second
	defaultSrvReadTimeout  = 15 * time.Second
	defaultProviderTimeout = 100 * time.Millisecond
)

var (
	validate = validator.New()

	// ErrEmptyConfigPath defines a sentinel error for an empty config path.
	ErrEmptyConfigPath = errors.New("empty configuration file path")

	// SupportedProviders defines a lookup table of all the supported currency API
	// providers.
	SupportedProviders = map[provider.Name]struct{}{
		provider.ProviderKraken:   {},
		provider.ProviderBinance:  {},
		provider.ProviderOsmosis:  {},
		provider.ProviderJuno:     {},
		provider.ProviderOkx:      {},
		provider.ProviderHuobi:    {},
		provider.ProviderGate:     {},
		provider.ProviderCoinbase: {},
		provider.ProviderBitget:   {},
		provider.ProviderMexc:     {},
		provider.ProviderCrypto:   {},
		provider.ProviderMock:     {},
	}

	// maxDeviationThreshold is the maxmimum allowed amount of standard
	// deviations which validators are able to set for a given asset.
	maxDeviationThreshold = sdk.MustNewDecFromStr("3.0")

	// SupportedQuotes defines a lookup table for which assets we support
	// using as quotes.
	SupportedQuotes = map[string]struct{}{
		DenomUSD: {},
		"USDC":   {},
		"USDT":   {},
		"DAI":    {},
		"BTC":    {},
		"ETH":    {},
	}
)

type (
	// Config defines all necessary price-feeder configuration parameters.
	Config struct {
		Server              Server              `mapstructure:"server"`
		CurrencyPairs       []CurrencyPair      `mapstructure:"currency_pairs" validate:"required,gt=0,dive,required"`
		Deviations          []Deviation         `mapstructure:"deviation_thresholds"`
		Account             Account             `mapstructure:"account" validate:"required,gt=0,dive,required"`
		Keyring             Keyring             `mapstructure:"keyring" validate:"required,gt=0,dive,required"`
		RPC                 RPC                 `mapstructure:"rpc" validate:"required,gt=0,dive,required"`
		Telemetry           telemetry.Config    `mapstructure:"telemetry"`
		GasAdjustment       float64             `mapstructure:"gas_adjustment" validate:"required"`
		ProviderTimeout     string              `mapstructure:"provider_timeout"`
		ProviderMinOverride bool                `mapstructure:"provider_min_override"`
		ProviderEndpoints   []provider.Endpoint `mapstructure:"provider_endpoints" validate:"dive"`
		GasPrice            string              `mapstructure:"gas_price"`
		FeeGrantAddress     string              `mapstructure:"fee_grant_address"`
	}

	// Server defines the API server configuration.
	Server struct {
		ListenAddr     string   `mapstructure:"listen_addr"`
		WriteTimeout   string   `mapstructure:"write_timeout"`
		ReadTimeout    string   `mapstructure:"read_timeout"`
		VerboseCORS    bool     `mapstructure:"verbose_cors"`
		AllowedOrigins []string `mapstructure:"allowed_origins"`
	}

	// CurrencyPair defines a price quote of the exchange rate for two different
	// currencies and the supported providers for getting the exchange rate.
	CurrencyPair struct {
		Base      string          `mapstructure:"base" validate:"required"`
		Quote     string          `mapstructure:"quote" validate:"required"`
		Providers []provider.Name `mapstructure:"providers" validate:"required,gt=0,dive,required"`
	}

	// Deviation defines a maximum amount of standard deviations that a given asset can
	// be from the median without being filtered out before voting.
	Deviation struct {
		Base      string `mapstructure:"base" validate:"required"`
		Threshold string `mapstructure:"threshold" validate:"required"`
	}

	// Account defines account related configuration that is related to the JUNO
	// network and transaction signing functionality.
	Account struct {
		ChainID   string `mapstructure:"chain_id" validate:"required"`
		Address   string `mapstructure:"address" validate:"required"`
		Validator string `mapstructure:"validator" validate:"required"`
	}

	// Keyring defines the required JUNO keyring configuration.
	Keyring struct {
		Backend string `mapstructure:"backend" validate:"required"`
		Dir     string `mapstructure:"dir" validate:"required"`
	}

	// RPC defines RPC configuration of both the JUNO gRPC and Tendermint nodes.
	RPC struct {
		TMRPCEndpoint string `mapstructure:"tmrpc_endpoint" validate:"required"`
		GRPCEndpoint  string `mapstructure:"grpc_endpoint" validate:"required"`
		RPCTimeout    string `mapstructure:"rpc_timeout" validate:"required"`
	}
)

// telemetryValidation is custom validation for the Telemetry struct.
func telemetryValidation(sl validator.StructLevel) {
	tel := sl.Current().Interface().(telemetry.Config)

	if tel.Enabled && (len(tel.GlobalLabels) == 0 || len(tel.ServiceName) == 0) {
		sl.ReportError(tel.Enabled, "enabled", "Enabled", "enabledNoOptions", "")
	}
}

// endpointValidation is custom validation for the ProviderEndpoint struct.
func endpointValidation(sl validator.StructLevel) {
	endpoint := sl.Current().Interface().(provider.Endpoint)

	if len(endpoint.Name) < 1 || len(endpoint.Rest) < 1 || len(endpoint.Websocket) < 1 {
		sl.ReportError(endpoint, "endpoint", "Endpoint", "unsupportedEndpointType", "")
	}
	if _, ok := SupportedProviders[endpoint.Name]; !ok {
		sl.ReportError(endpoint.Name, "name", "Name", "unsupportedEndpointProvider", "")
	}
}

// Validate returns an error if the Config object is invalid.
func (c Config) Validate() error {
	validate.RegisterStructValidation(telemetryValidation, telemetry.Config{})
	validate.RegisterStructValidation(endpointValidation, provider.Endpoint{})
	return validate.Struct(c)
}

// ParseConfig attempts to read and parse configuration from the given file path.
// An error is returned if reading or parsing the config fails.
func ParseConfig(configPath string) (Config, error) {
	var cfg Config

	if configPath == "" {
		return cfg, ErrEmptyConfigPath
	}

	viper.AutomaticEnv()
	viper.SetConfigFile(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config: %w", err)
	}

	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = defaultListenAddr
	}
	if len(cfg.Server.WriteTimeout) == 0 {
		cfg.Server.WriteTimeout = defaultSrvWriteTimeout.String()
	}
	if len(cfg.Server.ReadTimeout) == 0 {
		cfg.Server.ReadTimeout = defaultSrvReadTimeout.String()
	}
	if len(cfg.ProviderTimeout) == 0 {
		cfg.ProviderTimeout = defaultProviderTimeout.String()
	}

	pairs := make(map[string]map[provider.Name]struct{})
	coinQuotes := make(map[string]struct{})
	for _, cp := range cfg.CurrencyPairs {
		if _, ok := pairs[cp.Base]; !ok {
			pairs[cp.Base] = make(map[provider.Name]struct{})
		}
		if strings.ToUpper(cp.Quote) != DenomUSD {
			coinQuotes[cp.Quote] = struct{}{}
		}
		if _, ok := SupportedQuotes[strings.ToUpper(cp.Quote)]; !ok {
			return cfg, fmt.Errorf("unsupported quote: %s", cp.Quote)
		}

		for _, provider := range cp.Providers {
			if _, ok := SupportedProviders[provider]; !ok {
				return cfg, fmt.Errorf("unsupported provider: %s", provider)
			}
			pairs[cp.Base][provider] = struct{}{}
		}
	}

	// Use coinQuotes to ensure that any quotes can be converted to USD.
	for quote := range coinQuotes {
		for index, pair := range cfg.CurrencyPairs {
			if pair.Base == quote && pair.Quote == DenomUSD {
				break
			}
			if index == len(cfg.CurrencyPairs)-1 {
				return cfg, fmt.Errorf("all non-usd quotes require a conversion rate feed")
			}
		}
	}

	gatePairs := []string{}
	for base, providers := range pairs {
		if _, ok := providers[provider.ProviderGate]; ok {
			gatePairs = append(gatePairs, base)
		}
	}
	if len(gatePairs) > 1 {
		return cfg, fmt.Errorf("gate provider does not support multiple pairs: %v", gatePairs)
	}

	for _, deviation := range cfg.Deviations {
		threshold, err := sdk.NewDecFromStr(deviation.Threshold)
		if err != nil {
			return cfg, fmt.Errorf("deviation thresholds must be numeric: %w", err)
		}

		if threshold.GT(maxDeviationThreshold) {
			return cfg, fmt.Errorf("deviation thresholds must not exceed 3.0")
		}
	}

	return cfg, cfg.Validate()
}

// CheckProviderMins starts the currency provider tracker to check the amount of
// providers available for a currency by querying CoinGecko's API. It will enforce
// a provider minimum for a given currency based on its available providers.
func CheckProviderMins(ctx context.Context, logger zerolog.Logger, cfg Config) error {
	currencyProviderTracker, err := NewCurrencyProviderTracker(ctx, logger, cfg.CurrencyPairs...)
	if err != nil {
		logger.Error().Err(err).Msg("failed to start currency provider tracker")
		// If currency tracker errors out and override flag is set, the price-feeder
		// will run without enforcing provider minimums.
		if cfg.ProviderMinOverride {
			return nil
		}
	}

	pairs := make(map[string]map[provider.Name]struct{})
	for _, cp := range cfg.CurrencyPairs {
		if _, ok := pairs[cp.Base]; !ok {
			pairs[cp.Base] = make(map[provider.Name]struct{})
		}
		for _, provider := range cp.Providers {

			pairs[cp.Base][provider] = struct{}{}
		}
	}

	for base, providers := range pairs {
		// If currency provider tracker errored, default to three providers as
		// the minimum.
		var minProviders int
		if currencyProviderTracker != nil {
			minProviders = currencyProviderTracker.CurrencyProviderMin[base]
		} else {
			minProviders = 2
		}
		if _, ok := pairs[base][provider.ProviderMock]; !ok && len(providers) < minProviders {
			return fmt.Errorf("must have at least %d providers for %s", minProviders, base)
		}
	}

	return nil
}
