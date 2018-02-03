package verification

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"time"

	reserve "github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/common"
)

const BASE_URL = "URL_WHEN_RELEASE"

type Verification struct {
	app  reserve.ReserveData
	core reserve.ReserveCore
}

func (self *Verification) fillRequest(req *http.Request, signNeeded bool, timepoint uint64) {
	if req.Method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Add("Accept", "application/json")
	if signNeeded {
		q := req.URL.Query()
		sig := url.Values{}
		q.Set("nonce", fmt.Sprintf("%d", timepoint))
		sig.Set("signature", self.signer.BinanceSign(q.Encode()))
		req.URL.RawQuery = q.Encode() + "&" + sig.Encode()
	}
}

func (self *Verification) GetResponse(
	method string, url string,
	params map[string]string, signNeeded bool, timepoint uint64) ([]byte, error) {

	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	self.fillRequest(req, signNeeded, timepoint)
	var err error
	var resp_body []byte
	log.Printf("request to: %s\n", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return resp_body, err
	} else {
		defer resp.Body.Close()
		resp_body, err = ioutil.ReadAll(resp.Body)
		log.Printf("request to %s, got response: %s\n", req.URL, common.TruncStr(resp_body))
		return resp_body, err
	}
}

func (self *Verification) GetPendingActivities(timepoint uint64) (common.ActivityRecord, error) {
	result := common.ActivityRecord{}
	resp_body, err := self.GetResponse(
		"GET",
		self.base + "/pendingactivities",
		map[string]string{},
		true,
		timepoint
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *Verification) GetActivities(timepoint uint64) (common.ActivityRecord, error) {
	result := common.ActivityRecord{}
	resp_body, err := self.GetResponse(
		"GET",
		self.base + "/activities",
		map[string]string{},
		true,
		timepoint
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *Verification) GetAuthData(timepoint uint64) (common.AuthDataResponse, error) {
	result := common.AuthDataResponse{}
	resp_body, err := self.GetResponse(
		"GET",
		BASE_URL + "/authdata",
		map[string]string{},
		true,
		timepoint
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *Verification) VerifyDeposit(amount *big.Int) error {
	var err error
	timepoint := common.GetTimepoint()
	token := common.GetToken("ETH")
	// deposit to exchanges
	for _, exchange := range common.SupportedExchanges {
		activityID, err := self.core.Deposit(exchange, token, amount, timepoint)
		if err != nil {
			return err
		}
		// check deposit data from api
		// pending activities
		pendingActivities, err := self.GetPendingActivities(timepoint)
		if err != nil {
			return errors.New(fmt.Sprintf("Deposit error, getting pending activities: %s", err.Error())
		}
		// authdata
		authData, err := self.GetAuthData()
		if err != nil {
			return errors.New(fmt.Sprintf("Deposit error, geting authdata: %s", err.Error()))
		}
		// activities
		activities, err := self.GetActivities()
		if err != nil {
			return errors.New(fmt.Sprintf("Deposit error, getting activities: %s", err.Error()))
		}
	}
	return err
}

func (self *Verification) VerifyWithdraw(amount *big.Int) error {
	var err error
	timepoint := common.GetTimepoint()
	token := common.GetToken("ETH")
	for _, exchange := range common.SupportedExchanges {
		activityID, err := self.core.Withdraw(exchange, token, amount, timepoint)
		// check withdraw data from api
		// pending activities
		pendingActivities, err := self.GetPendingActivities()
		if err != nil {
			return errors.New("Withdraw error, getting pending activities: %s", err.Error())
		}
		// authdata
		authdata, err := self.GetAuthData()
		if err != nil {
			return errors.New("Withdraw error, getting auth data: %s", err.Error())
		}
		// activities
		activities, err := self.GetActivities()
		if err != nil {
			return errors.New("Withdraw error, getting activities: %s", err.Error())
		}
	}
	return err
}

func NewVerification(
	app reserve.ReserveData,
	core reserve.ReserveCore) *Verification {
	return &Verification{
		app,
		core,
	}
}
