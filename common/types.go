package common

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Version uint64
type Timestamp string

func (self Timestamp) ToUint64() uint64 {
	res, err := strconv.ParseUint(string(self), 10, 64)
	if err != nil {
		panic(err)
	}
	return res
}

func GetTimestamp() Timestamp {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return Timestamp(strconv.Itoa(int(timestamp)))
}

func GetTimepoint() uint64 {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	return uint64(timestamp)
}

func TimeToTimepoint(t time.Time) uint64 {
	timestamp := t.UnixNano() / int64(time.Millisecond)
	return uint64(timestamp)
}

func TimepointToTime(t uint64) time.Time {
	return time.Unix(0, int64(t)*int64(time.Millisecond))
}

type ExchangePrecisionLimit struct {
	Precision   TokenPairPrecision
	AmountLimit TokenPairAmountLimit
	PriceLimit  TokenPairPriceLimit
}

// ExchangeInfo is written and read concurrently
type ExchangeInfo struct {
	mu   sync.RWMutex
	data map[TokenPairID]ExchangePrecisionLimit
}

func NewExchangeInfo() *ExchangeInfo {
	return &ExchangeInfo{
		mu:   sync.RWMutex{},
		data: map[TokenPairID]ExchangePrecisionLimit{},
	}
}

func (self *ExchangeInfo) Update(pair TokenPairID, data ExchangePrecisionLimit) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.data[pair] = data
}

func (self *ExchangeInfo) Get(pair TokenPairID) (ExchangePrecisionLimit, error) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	if info, exist := self.data[pair]; exist {
		return info, nil
	} else {
		return info, errors.New("Token pair is not existed")
	}
}

func (self *ExchangeInfo) GetData() map[TokenPairID]ExchangePrecisionLimit {
	self.mu.RLock()
	defer self.mu.RUnlock()
	return self.data
}

type TokenPairPrecision struct {
	Amount int
	Price  int
}

type TokenPairAmountLimit struct {
	Min float32
	Max float32
}

type TokenPairPriceLimit struct {
	Min float32
	Max float32
}

type TradingFee map[string]float32

type FundingFee struct {
	Withdraw map[string]float32
	Deposit  map[string]float32
}

type ExchangeFees struct {
	Trading TradingFee
	Funding FundingFee
}

func NewExchangeFee(tradingFee TradingFee, fundingFee FundingFee) ExchangeFees {
	return ExchangeFees{
		Trading: tradingFee,
		Funding: fundingFee,
	}
}

func NewFundingFee(withdraw map[string]float32, deposit map[string]float32) FundingFee {
	return FundingFee{
		withdraw,
		deposit,
	}
}

type TokenPairID string

func NewTokenPairID(base, quote string) TokenPairID {
	return TokenPairID(fmt.Sprintf("%s-%s", base, quote))
}

type ExchangeID string

type ActivityID struct {
	Timepoint uint64
	EID       string
}

func (self ActivityID) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s|%s", strconv.FormatUint(self.Timepoint, 10), self.EID)), nil
}

func (self *ActivityID) UnmarshalText(b []byte) error {
	id, err := StringToActivityID(string(b))
	if err != nil {
		return err
	} else {
		self.Timepoint = id.Timepoint
		self.EID = id.EID
		return nil
	}
}

func (self ActivityID) String() string {
	res, _ := self.MarshalText()
	return string(res)
}

func StringToActivityID(id string) (ActivityID, error) {
	result := ActivityID{}
	parts := strings.Split(id, "|")
	if len(parts) < 2 {
		return result, errors.New("Invalid activity id")
	} else {
		timeStr := parts[0]
		eid := strings.Join(parts[1:], "|")
		timepoint, err := strconv.ParseUint(timeStr, 10, 64)
		if err != nil {
			return result, err
		} else {
			result.Timepoint = timepoint
			result.EID = eid
			return result, nil
		}
	}
}

func NewActivityID(t uint64, id string) ActivityID {
	return ActivityID{t, id}
}

type ActivityRecord struct {
	Action         string
	ID             ActivityID
	Destination    string
	Params         map[string]interface{}
	Result         map[string]interface{}
	ExchangeStatus string
	MiningStatus   string
	Timestamp      Timestamp
}

func (self ActivityRecord) IsExchangePending() bool {
	switch self.Action {
	case "withdraw":
		return (self.ExchangeStatus == "" || self.ExchangeStatus == "submitted") &&
			self.MiningStatus != "failed"
	case "deposit":
		return (self.ExchangeStatus == "" || self.ExchangeStatus == "pending") &&
			self.MiningStatus != "failed"
	case "trade":
		return self.ExchangeStatus == "" || self.ExchangeStatus == "submitted"
	}
	return true
}

func (self ActivityRecord) IsBlockchainPending() bool {
	switch self.Action {
	case "withdraw", "deposit", "set_rates":
		return (self.MiningStatus == "" || self.MiningStatus == "submitted") && self.ExchangeStatus != "failed"
	}
	return true
}

func (self ActivityRecord) IsPending() bool {
	switch self.Action {
	case "withdraw":
		return (self.ExchangeStatus == "" || self.ExchangeStatus == "submitted" ||
			self.MiningStatus == "" || self.MiningStatus == "submitted") &&
			self.MiningStatus != "failed" && self.ExchangeStatus != "failed"
	case "deposit":
		return (self.ExchangeStatus == "" || self.ExchangeStatus == "pending" ||
			self.MiningStatus == "" || self.MiningStatus == "submitted") &&
			self.MiningStatus != "failed" && self.ExchangeStatus != "failed"
	case "trade":
		return (self.ExchangeStatus == "" || self.ExchangeStatus == "submitted") &&
			self.ExchangeStatus != "failed"
	case "set_rates":
		return (self.MiningStatus == "" || self.MiningStatus == "submitted") &&
			self.ExchangeStatus != "failed"
	}
	return true
}

type ActivityStatus struct {
	ExchangeStatus string
	Tx             string
	MiningStatus   string
	Error          error
}

type PriceEntry struct {
	Quantity float64
	Rate     float64
}

type AllPriceEntry struct {
	Block uint64
	Data  map[TokenPairID]OnePrice
}

type AllPriceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[TokenPairID]OnePrice
	Block      uint64
}

type OnePriceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       OnePrice
	Block      uint64
}

type OnePrice map[ExchangeID]ExchangePrice

type ExchangePrice struct {
	Valid      bool
	Error      string
	Timestamp  Timestamp
	Bids       []PriceEntry
	Asks       []PriceEntry
	ReturnTime Timestamp
}

func BigToFloat(b *big.Int, decimal int64) float64 {
	f := new(big.Float).SetInt(b)
	power := new(big.Float).SetInt(new(big.Int).Exp(
		big.NewInt(10), big.NewInt(decimal), nil,
	))
	res := new(big.Float).Quo(f, power)
	result, _ := res.Float64()
	return result
}

type RawBalance big.Int

func (self *RawBalance) ToFloat(decimal int64) float64 {
	return BigToFloat((*big.Int)(self), decimal)
}

func (self RawBalance) MarshalJSON() ([]byte, error) {
	selfInt := (big.Int)(self)
	return selfInt.MarshalJSON()
}

func (self *RawBalance) UnmarshalJSON(text []byte) error {
	selfInt := (*big.Int)(self)
	return selfInt.UnmarshalJSON(text)
}

type BalanceEntry struct {
	Valid      bool
	Error      string
	Timestamp  Timestamp
	ReturnTime Timestamp
	Balance    RawBalance
}

func (self BalanceEntry) ToBalanceResponse(decimal int64) BalanceResponse {
	return BalanceResponse{
		Valid:      self.Valid,
		Error:      self.Error,
		Timestamp:  self.Timestamp,
		ReturnTime: self.ReturnTime,
		Balance:    self.Balance.ToFloat(decimal),
	}
}

type BalanceResponse struct {
	Valid      bool
	Error      string
	Timestamp  Timestamp
	ReturnTime Timestamp
	Balance    float64
}

type AllBalanceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[string]BalanceResponse
}

type Order struct {
	ID          string `standard id across multiple exchanges`
	Base        string
	Quote       string
	OrderId     string
	Price       float64
	OrigQty     float64 `original quantity`
	ExecutedQty float64 `matched quantity`
	TimeInForce string
	Type        string `market or limit`
	Side        string `buy or sell`
	StopPrice   string
	IcebergQty  string
	Time        uint64
}

type OrderEntry struct {
	Valid      bool
	Error      string
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       []Order
}

type AllOrderEntry map[ExchangeID]OrderEntry

type AllOrderResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       AllOrderEntry
}

type EBalanceEntry struct {
	Valid            bool
	Error            string
	Timestamp        Timestamp
	ReturnTime       Timestamp
	AvailableBalance map[string]float64
	LockedBalance    map[string]float64
	DepositBalance   map[string]float64
}

type AllEBalanceResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[ExchangeID]EBalanceEntry
}

type AuthDataSnapshot struct {
	Valid             bool
	Error             string
	Timestamp         Timestamp
	ReturnTime        Timestamp
	ExchangeBalances  map[ExchangeID]EBalanceEntry
	ReserveBalances   map[string]BalanceEntry
	PendingActivities []ActivityRecord
	Block             uint64
}

type AuthDataResponse struct {
	Version    Version
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       struct {
		Valid             bool
		Error             string
		Timestamp         Timestamp
		ReturnTime        Timestamp
		ExchangeBalances  map[ExchangeID]EBalanceEntry
		ReserveBalances   map[string]BalanceResponse
		PendingActivities []ActivityRecord
		Block             uint64
	}
}

type RateEntry struct {
	BaseBuy     *big.Int
	CompactBuy  int8
	BaseSell    *big.Int
	CompactSell int8
	Block       uint64
}

type RateResponse struct {
	Valid       bool
	Error       string
	Timestamp   Timestamp
	ReturnTime  Timestamp
	BaseBuy     float64
	CompactBuy  int8
	BaseSell    float64
	CompactSell int8
	Rate        float64
	Block       uint64
}

type AllRateEntry struct {
	Valid      bool
	Error      string
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[string]RateEntry
}

type AllRateResponse struct {
	Version    Version
	Valid      bool
	Error      string
	Timestamp  Timestamp
	ReturnTime Timestamp
	Data       map[string]RateResponse
}
