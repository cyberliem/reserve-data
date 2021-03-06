package core

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type ReserveCore struct {
	blockchain      Blockchain
	activityStorage ActivityStorage
	rm              ethereum.Address
}

func NewReserveCore(
	blockchain Blockchain,
	storage ActivityStorage,
	rm ethereum.Address) *ReserveCore {
	return &ReserveCore{
		blockchain,
		storage,
		rm,
	}
}

func timebasedID(id string) common.ActivityID {
	return common.NewActivityID(uint64(time.Now().UnixNano()), id)
}

func (self ReserveCore) CancelOrder(id common.ActivityID, exchange common.Exchange) error {
	return exchange.CancelOrder(id)
}

func (self ReserveCore) Trade(
	exchange common.Exchange,
	tradeType string,
	base common.Token,
	quote common.Token,
	rate float64,
	amount float64,
	timepoint uint64) (common.ActivityID, float64, float64, bool, error) {

	id, done, remaining, finished, err := exchange.Trade(tradeType, base, quote, rate, amount, timepoint)
	var status string
	if err != nil {
		status = "failed"
	} else {
		if finished {
			status = "done"
		} else {
			status = "submitted"
		}
	}
	uid := timebasedID(id)
	self.activityStorage.Record(
		"trade",
		uid,
		string(exchange.ID()),
		map[string]interface{}{
			"exchange":  exchange,
			"type":      tradeType,
			"base":      base,
			"quote":     quote,
			"rate":      rate,
			"amount":    strconv.FormatFloat(amount, 'f', -1, 64),
			"timepoint": timepoint,
		}, map[string]interface{}{
			"id":        id,
			"done":      done,
			"remaining": remaining,
			"finished":  finished,
			"error":     err,
		},
		status,
		"",
		timepoint,
	)
	log.Printf(
		"Core ----------> %s on %s: base: %s, quote: %s, rate: %s, amount: %s, timestamp: %d ==> Result: id: %s, done: %s, remaining: %s, finished: %t, error: %s",
		tradeType, exchange.ID(), base.ID, quote.ID,
		strconv.FormatFloat(rate, 'f', -1, 64),
		strconv.FormatFloat(amount, 'f', -1, 64), timepoint,
		uid,
		strconv.FormatFloat(done, 'f', -1, 64),
		strconv.FormatFloat(remaining, 'f', -1, 64),
		finished, err,
	)
	return uid, done, remaining, finished, err
}

func (self ReserveCore) Deposit(
	exchange common.Exchange,
	token common.Token,
	amount *big.Int,
	timepoint uint64) (common.ActivityID, error) {

	address, supported := exchange.Address(token)
	tx := ethereum.Hash{}
	var err error
	if !supported {
		tx = ethereum.Hash{}
		err = errors.New(fmt.Sprintf("Exchange %s doesn't support token %s", exchange.ID(), token.ID))
	} else if self.activityStorage.HasPendingDeposit(token, exchange) {
		tx = ethereum.Hash{}
		err = errors.New(fmt.Sprintf("There is a pending %s deposit to %s currently, please try again", token.ID, exchange.ID()))
	} else {
		tx, err = self.blockchain.Send(token, amount, address)
	}
	var status string
	if err != nil {
		status = "failed"
	} else {
		status = "submitted"
	}
	amountFloat := common.BigToFloat(amount, token.Decimal)
	uid := timebasedID(tx.Hex() + "|" + token.ID + "|" + strconv.FormatFloat(amountFloat, 'f', -1, 64))
	self.activityStorage.Record(
		"deposit",
		uid,
		string(exchange.ID()),
		map[string]interface{}{
			"exchange":  exchange,
			"token":     token,
			"amount":    strconv.FormatFloat(amountFloat, 'f', -1, 64),
			"timepoint": timepoint,
		}, map[string]interface{}{
			"tx":    tx.Hex(),
			"error": err,
		},
		"",
		status,
		timepoint,
	)
	log.Printf(
		"Core ----------> Deposit to %s: token: %s, amount: %s, timestamp: %d ==> Result: tx: %s, error: %s",
		exchange.ID(), token.ID, amount.Text(10), timepoint, tx.Hex(), err,
	)
	return uid, err
}

func (self ReserveCore) Withdraw(
	exchange common.Exchange, token common.Token,
	amount *big.Int, timepoint uint64) (common.ActivityID, error) {

	_, supported := exchange.Address(token)
	var err error
	var id string
	if !supported {
		err = errors.New(fmt.Sprintf("Exchange %s doesn't support token %s", exchange.ID(), token.ID))
	} else {
		id, err = exchange.Withdraw(token, amount, self.rm, timepoint)
	}
	var status string
	if err != nil {
		status = "failed"
	} else {
		status = "submitted"
	}
	uid := timebasedID(id)
	self.activityStorage.Record(
		"withdraw",
		uid,
		string(exchange.ID()),
		map[string]interface{}{
			"exchange":  exchange,
			"token":     token,
			"amount":    strconv.FormatFloat(common.BigToFloat(amount, token.Decimal), 'f', -1, 64),
			"timepoint": timepoint,
		}, map[string]interface{}{
			"error": err,
			"id":    id,
			// this field will be updated with real tx when data fetcher can fetch it
			// from exchanges
			"tx": "",
		},
		status,
		"",
		timepoint,
	)
	log.Printf(
		"Core ----------> Withdraw from %s: token: %s, amount: %s, timestamp: %d ==> Result: id: %s, error: %s",
		exchange.ID(), token.ID, amount.Text(10), timepoint, id, err,
	)
	return uid, err
}

func (self ReserveCore) SetRates(
	tokens []common.Token,
	buys []*big.Int,
	sells []*big.Int,
	block *big.Int) (common.ActivityID, error) {

	lentokens := len(tokens)
	lenbuys := len(buys)
	lensells := len(sells)
	tx := ethereum.Hash{}
	var err error
	if lentokens != lenbuys || lentokens != lensells {
		err = errors.New("Tokens, buys and sells must have the same length")
	} else {
		tokenAddrs := []ethereum.Address{}
		for _, token := range tokens {
			tokenAddrs = append(tokenAddrs, ethereum.HexToAddress(token.Address))
		}
		tx, err = self.blockchain.SetRates(tokenAddrs, buys, sells, block)
	}
	var status string
	if err != nil {
		status = "failed"
	} else {
		status = "submitted"
	}
	uid := timebasedID(tx.Hex())
	self.activityStorage.Record(
		"set_rates",
		uid,
		"blockchain",
		map[string]interface{}{
			"tokens": tokens,
			"buys":   buys,
			"sells":  sells,
			"block":  block,
		}, map[string]interface{}{
			"tx":    tx.Hex(),
			"error": err,
		},
		"",
		status,
		common.GetTimepoint(),
	)
	log.Printf(
		"Core ----------> Set rates: ==> Result: tx: %s, error: %s",
		tx.Hex(), err,
	)
	return uid, err
}
