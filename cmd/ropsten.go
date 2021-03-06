package main

import (
	"log"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/metric"
	"github.com/KyberNetwork/reserve-data/signer"
	ethereum "github.com/ethereum/go-ethereum/common"
)

func GetConfigForRopsten() *Config {
	settingPath := "/go/src/github.com/KyberNetwork/reserve-data/cmd/ropsten_setting.json"
	addressConfig, err := common.GetAddressConfigFromFile(settingPath)
	if err != nil {
		log.Fatalf("Config file %s is not found. Error: %s", settingPath, err)
	}
	wrapperAddr := ethereum.HexToAddress(addressConfig.Wrapper)
	pricingAddr := ethereum.HexToAddress(addressConfig.Pricing)
	reserveAddr := ethereum.HexToAddress(addressConfig.Reserve)

	common.SupportedTokens = map[string]common.Token{}
	tokens := []common.Token{}
	for id, t := range addressConfig.Tokens {
		tok := common.Token{
			id, t.Address, t.Decimals,
		}
		common.SupportedTokens[id] = tok
		tokens = append(tokens, tok)
	}

	storage := storage.NewRamStorage()
	metricStorage := metric.NewRamMetricStorage()

	fetcherRunner := fetcher.NewTickerRunner(3*time.Second, 2*time.Second, 3*time.Second, 5*time.Second)

	fileSigner := signer.NewFileSigner("/go/src/github.com/KyberNetwork/reserve-data/cmd/config.json")

	exchangePool := NewRopstenExchangePool(
		addressConfig, fileSigner, storage,
	)

	// endpoint := "http://localhost:8545"
	// endpoint := "https://ropsten.kyber.network"
	endpoint := "https://ropsten.infura.io"

	return &Config{
		ActivityStorage:  storage,
		DataStorage:      storage,
		FetcherStorage:   storage,
		MetricStorage:    metricStorage,
		FetcherRunner:    fetcherRunner,
		FetcherExchanges: exchangePool.FetcherExchanges(),
		Exchanges:        exchangePool.CoreExchanges(),
		BlockchainSigner: fileSigner,
		EthereumEndpoint: endpoint,
		SupportedTokens:  tokens,
		WrapperAddress:   wrapperAddr,
		PricingAddress:   pricingAddr,
		ReserveAddress:   reserveAddr,
	}
}
