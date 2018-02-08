package verification

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	ihttp "github.com/KyberNetwork/reserve-data/http"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const BASE_URL = "http://localhost:8000"

type Verification struct {
	auth      ihttp.Authentication
	exchanges []string
}

type DepositWithdrawResponse struct {
	Success bool              `json:"success"`
	ID      common.ActivityID `json:"id"`
	Reason  string            `json:"reason"`
}

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

func InitLogger(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
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
	resp, err := client.Do(req)
	if err != nil {
		return resp_body, err
	} else {
		defer resp.Body.Close()
		resp_body, err = ioutil.ReadAll(resp.Body)
		Info.Printf("request to %s, got response: %s\n", req.URL, common.TruncStr(resp_body))
		return resp_body, err
	}
}

func (self *Verification) GetPendingActivities(timepoint uint64) ([]common.ActivityRecord, error) {
	result := []common.ActivityRecord{}
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

func (self *Verification) Deposit(
	exchange, token, amount string, timepoint uint64) (common.ActivityID, error) {
	result := DepositWithdrawResponse{}
	resp_body, err := self.GetResponse(
		"POST",
		BASE_URL+"/deposit/"+exchange,
		map[string]string{
			"amount": amount,
			"token":  token,
		},
		true,
		timepoint,
	)
	if err != nil {
		return result.ID, err
	}
	json.Unmarshal(resp_body, &result)
	if result.Success != true {
		err = errors.New(fmt.Sprintf("Cannot deposit: %s", result.Reason))
	}
	return result.ID, err
}

func (self *Verification) Withdraw(
	exchange, token, amount string, timepoint uint64) (common.ActivityID, error) {
	result := DepositWithdrawResponse{}
	resp_body, err := self.GetResponse(
		"POST",
		BASE_URL+"/withdraw/"+exchange,
		map[string]string{
			"amount": amount,
			"token":  token,
		},
		true,
		timepoint,
	)
	if err != nil {
		return result.ID, err
	}
	json.Unmarshal(resp_body, &result)
	if result.Success != true {
		err = errors.New(fmt.Sprintf("Cannot withdraw: %s", result.Reason))
	}
	return result.ID, nil
}

func (self *Verification) VerifyDeposit() error {
	var err error
	timepoint := common.GetTimepoint()
	token := "ETH"
	amount := hexutil.EncodeUint64(1)
	// deposit to exchanges
	Info.Println("Start deposit to exchanges")
	for _, exchange := range self.exchanges {
		activityID, err := self.Deposit(exchange, token, amount, timepoint)
		if err != nil {
			Error.Println(err.Error())
			return err
		}
		Info.Printf("Deposit id: %s", activityID)
		// check deposit data from api
		// pending activities
		pendingActivities, err := self.GetPendingActivities(timepoint)
		if err != nil {
			Error.Println(err.Error())
			return err
		}
		Info.Printf("Pending activities after deposit: %v", pendingActivities)
		// check if activities available in pending
		available := false
		for _, pending := range pendingActivities {
			if pending.ID == activityID {
				available = true
				break
			}
		}
		if !available {
			Error.Println("Deposit activity did not store")
		}
		// authdata
		authData, err := self.GetAuthData(timepoint)
		if err != nil {
			Error.Println(err.Error())
			return err
		}
		Info.Printf("Auth data after deposit: %v", authData)
		// activities
		activities, err := self.GetActivities(timepoint)
		if err != nil {
			Error.Println(err.Error())
			return err
		}
		Info.Printf("Activity data after deposit: %v", activities)
	}
	return err
}

func (self *Verification) VerifyWithdraw() error {
	var err error
	timepoint := common.GetTimepoint()
	token := "ETH"
	amount := hexutil.EncodeUint64(1)
	for _, exchange := range self.exchanges {
		activityID, err := self.Withdraw(exchange, token, amount, timepoint)
		if err != nil {
			Error.Println(err.Error())
			return err
		}
		Info.Printf("Withdraw ID: %s", activityID)
		// check withdraw data from api
		// pending activities
		pendingActivities, err := self.GetPendingActivities(timepoint)
		if err != nil {
			Error.Println(err.Error())
			return err
		}
		Info.Printf("Pending activities after withdraw: %v", pendingActivities)
		available := false
		for _, pending := range pendingActivities {
			if pending.ID == activityID {
				available = true
				break
			}
		}
		if !available {
			Error.Println("Withdraw activity did not store")
		}
		// authdata
		authdata, err := self.GetAuthData(timepoint)
		if err != nil {
			Error.Println(err.Error())
			return err
		}
		Info.Printf("Auth data after withdraw: %s", authdata)
		// activities
		activities, err := self.GetActivities(timepoint)
		if err != nil {
			Error.Println(err.Error())
			return err
		}
		Info.Printf("Activities after withdraw: %v", activities)
	}
	return err
}

func (self *Verification) RunVerification() {
	InitLogger(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	Info.Println("Start verification")
	self.VerifyDeposit()
	// self.VerifyWithdraw()
}

func NewVerification(
	auth ihttp.Authentication) *Verification {
	params := os.Getenv("KYBER_EXCHANGES")
	exchanges := strings.Split(params, ",")
	return &Verification{
		auth,
		exchanges,
	}
}
