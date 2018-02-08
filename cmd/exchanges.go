package main

import (
	"os"
	"strings"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/bittrex"
	"github.com/KyberNetwork/reserve-data/signer"
)

type ExchangePool struct {
	Exchanges map[common.ExchangeID]interface{}
}

func AsyncUpdateDepositAddress(ex common.Exchange, tokenID, addr string, wait *sync.WaitGroup) {
	defer wait.Done()
	ex.UpdateDepositAddress(common.MustGetToken(tokenID), addr)
}

func NewSimulationExchangePool(
	addressConfig common.AddressConfig,
	signer *signer.FileSigner,
	bittrexStorage exchange.BittrexStorage) *ExchangePool {

	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "bittrex":
			endpoint := bittrex.NewSimulatedBittrexEndpoint(signer)
			bit := exchange.NewBittrex(addressConfig.Exchanges["bittrex"], endpoint, bittrexStorage)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bit, tokenID, addr, &wait)
			}
			wait.Wait()
			bit.UpdatePairsPrecision()
			exchanges[bit.ID()] = bit
		case "binance":
			endpoint := binance.NewSimulatedBinanceEndpoint(signer)
			bin := exchange.NewBinance(addressConfig.Exchanges["binance"], endpoint)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bin, tokenID, addr, &wait)
			}
			wait.Wait()
			bin.UpdatePairsPrecision()
			exchanges[bin.ID()] = bin
		}
	}
	return &ExchangePool{exchanges}
}

func NewDevExchangePool(addressConfig common.AddressConfig, signer *signer.FileSigner, bittrexStorage exchange.BittrexStorage) *ExchangePool {
	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "bittrex":
			endpoint := bittrex.NewDevBittrexEndpoint(signer)
			bit := exchange.NewBittrex(addressConfig.Exchanges["bittrex"], endpoint, bittrexStorage)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bit, tokenID, addr, &wait)
			}
			wait.Wait()
			bit.UpdatePairsPrecision()
			exchanges[bit.ID()] = bit
		case "binance":
			endpoint := binance.NewDevBinanceEndpoint(signer)
			bin := exchange.NewBinance(addressConfig.Exchanges["binance"], endpoint)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bin, tokenID, addr, &wait)
			}
			wait.Wait()
			bin.UpdatePairsPrecision()
			exchanges[bin.ID()] = bin
		}
	}
	return &ExchangePool{exchanges}
}

func NewKovanExchangePool(addressConfig common.AddressConfig, signer *signer.FileSigner, bittrexStorage exchange.BittrexStorage) *ExchangePool {
	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "bittrex":
			endpoint := bittrex.NewKovanBittrexEndpoint(signer)
			bit := exchange.NewBittrex(addressConfig.Exchanges["bittrex"], endpoint, bittrexStorage)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bit, tokenID, addr, &wait)
			}
			wait.Wait()
			bit.UpdatePairsPrecision()
			exchanges[bit.ID()] = bit
		case "binance":
			endpoint := binance.NewKovanBinanceEndpoint(signer)
			bin := exchange.NewBinance(addressConfig.Exchanges["binance"], endpoint)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bin, tokenID, addr, &wait)
			}
			wait.Wait()
			bin.UpdatePairsPrecision()
			exchanges[bin.ID()] = bin
		}
	}
	return &ExchangePool{exchanges}
}

func NewRopstenExchangePool(addressConfig common.AddressConfig, signer *signer.FileSigner, bittrexStorage exchange.BittrexStorage) *ExchangePool {
	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "bittrex":
			endpoint := bittrex.NewRopstenBittrexEndpoint(signer)
			bit := exchange.NewBittrex(addressConfig.Exchanges["bittrex"], endpoint, bittrexStorage)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bit, tokenID, addr, &wait)
			}
			wait.Wait()
			bit.UpdatePairsPrecision()
			exchanges[bit.ID()] = bit
		case "binance":
			endpoint := binance.NewRopstenBinanceEndpoint(signer)
			bin := exchange.NewBinance(addressConfig.Exchanges["binance"], endpoint)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bin, tokenID, addr, &wait)
			}
			wait.Wait()
			bin.UpdatePairsPrecision()
			exchanges[bin.ID()] = bin
		}
	}
	return &ExchangePool{exchanges}
}

func NewMainnetExchangePool(addressConfig common.AddressConfig, signer *signer.FileSigner, bittrexStorage exchange.BittrexStorage) *ExchangePool {
	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "bittrex":
			endpoint := bittrex.NewRealBittrexEndpoint(signer)
			bit := exchange.NewBittrex(addressConfig.Exchanges["bittrex"], endpoint, bittrexStorage)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bit, tokenID, addr, &wait)
			}
			wait.Wait()
			bit.UpdatePairsPrecision()
			exchanges[bit.ID()] = bit
		case "binance":
			endpoint := binance.NewRealBinanceEndpoint(signer)
			bin := exchange.NewBinance(addressConfig.Exchanges["binance"], endpoint)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bin, tokenID, addr, &wait)
			}
			wait.Wait()
			bin.UpdatePairsPrecision()
			exchanges[bin.ID()] = bin
		}
	}
	return &ExchangePool{exchanges}
}

func (self *ExchangePool) FetcherExchanges() []fetcher.Exchange {
	result := []fetcher.Exchange{}
	for _, ex := range self.Exchanges {
		result = append(result, ex.(fetcher.Exchange))
	}
	return result
}

func (self *ExchangePool) CoreExchanges() []common.Exchange {
	result := []common.Exchange{}
	for _, ex := range self.Exchanges {
		result = append(result, ex.(common.Exchange))
	}
	return result
}
