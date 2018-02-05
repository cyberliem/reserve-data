package verification

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"time"

	reserve "github.com/KyberNetwork/reserve-data"
	"github.com/KyberNetwork/reserve-data/common"
	ihttp "github.com/KyberNetwork/reserve-data/http"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const BASE_URL = "http://localhost:8000"

type Verification struct {
	app  reserve.ReserveData
	core reserve.ReserveCore
	auth ihttp.Authentication
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
		sig.Set("signature", self.auth.KNSign(q.Encode()))
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
		BASE_URL+"/immediate-pending-activities",
		map[string]string{},
		true,
		timepoint,
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
		BASE_URL+"/activities",
		map[string]string{},
		true,
		timepoint,
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
		BASE_URL+"/authdata",
		map[string]string{},
		true,
		timepoint,
	)
	if err == nil {
		err = json.Unmarshal(resp_body, &result)
	}
	return result, err
}

func (self *Verification) VerifyDeposit(amount *big.Int) error {
	var err error
	timepoint := common.GetTimepoint()
	token, _ := common.GetToken("ETH")
	// deposit to exchanges
	for _, exchange := range common.SupportedExchanges {
		activityID, err := self.core.Deposit(exchange, token, amount, timepoint)
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot deposit: %s", err.Error()))
		}
		log.Printf("Deposit id: %s", activityID)
		// check deposit data from api
		// pending activities
		pendingActivities, err := self.GetPendingActivities(timepoint)
		if err != nil {
			return errors.New(fmt.Sprintf("Deposit error, getting pending activities: %s", err.Error()))
		}
		log.Printf("Pending activities after deposit: %v", pendingActivities)
		// authdata
		authData, err := self.GetAuthData(timepoint)
		if err != nil {
			return errors.New(fmt.Sprintf("Deposit error, geting authdata: %s", err.Error()))
		}
		log.Printf("Auth data after deposit: %v", authData)
		// activities
		activities, err := self.GetActivities(timepoint)
		if err != nil {
			return errors.New(fmt.Sprintf("Deposit error, getting activities: %s", err.Error()))
		}
		log.Printf("Activity data after deposit: %v", activities)
	}
	return err
}

func (self *Verification) VerifyWithdraw(amount *big.Int) error {
	var err error
	timepoint := common.GetTimepoint()
	token, _ := common.GetToken("ETH")
	for _, exchange := range common.SupportedExchanges {
		activityID, err := self.core.Withdraw(exchange, token, amount, timepoint)
		if err != nil {
			log.Printf("Cannot withdraw: %s", err.Error())
		}
		log.Printf("Withdraw ID: %s", activityID)
		// check withdraw data from api
		// pending activities
		pendingActivities, err := self.GetPendingActivities(timepoint)
		if err != nil {
			return errors.New(fmt.Sprintf("Withdraw error, getting pending activities: %s", err.Error()))
		}
		log.Printf("Pending activities after withdraw: %v", pendingActivities)
		// authdata
		authdata, err := self.GetAuthData(timepoint)
		if err != nil {
			return errors.New(fmt.Sprintf("Withdraw error, getting auth data: %s", err.Error()))
		}
		log.Printf("Auth data after withdraw: %s", authdata)
		// activities
		activities, err := self.GetActivities(timepoint)
		if err != nil {
			return errors.New(fmt.Sprintf("Withdraw error, getting activities: %s", err.Error()))
		}
		log.Printf("Activities after withdraw: %v", activities)
	}
	return err
}

func (self *Verification) RunVerification() error {
	amount, _ := hexutil.DecodeBig("1")
	var err error
	err = self.VerifyDeposit(amount)
	if err != nil {
		log.Printf(err.Error())
	}
	// err = self.VerifyWithdraw(amount)
	// if err != nil {
	// 	log.Printf(err.Error())
	// }
	// log.Printf("Verify deployment successfully")
	return err
}

func NewVerification(
	app reserve.ReserveData,
	core reserve.ReserveCore,
	auth ihttp.Authentication) *Verification {
	return &Verification{
		app,
		core,
		auth,
	}
}
