package paystack

import (
	"fmt"
	"encoding/json"
	"net/http"
	g "github.com/kenmobility/feezbot/gateways"
)

//ChargeByBankDetails is a func that collects the detials of a customer's card and tries to charge it directory
//it returns a struct object
func ChargeByBankDetails(customerEmail, bankCode,accountNumber,birthday,txSubject,txDesc string, amount int) *g.ChargeResponse {
	url := "https://api.paystack.co/charge"
	//mainAmount := amount * 100 
	postData := fmt.Sprintf(`{
		"email": "%s",
		"amount": %v,
		"metadata":{
			"custom_fields":[
			  {
				"display_name": "%s",
				"value":"%s"	
			  }
			]
		  },
		"bank" : {
			"code": "%s",
			"account_number": "%s"
		},
		"birthday":"%s"
	}`,customerEmail,amount * 100,txSubject,txDesc,bankCode,accountNumber,birthday)
	//fmt.Println("the req body is ", postData)
	resp, statusCode := sendHTTPPostRequest(url, "POST", postData)
	if statusCode == -1 {
		fmt.Println("bank_charge.go::ChargeByBankDetails()::error in making request due to: ", resp)
		return &g.ChargeResponse {
			StatusCode : http.StatusInternalServerError,
			ResponseMsg :  resp,
			ResponseStatus : "http request failed",
		}
	} 
	var br paystackChargeResponse
	err := json.Unmarshal([]byte(resp), &br)
	if err != nil {
		fmt.Println("bank_charge.go::ChargeByBankDetails()::error in unmarshalling response due to: ", err)
	}
	if br.Status == false {
		fmt.Println("bank_charge.go::ChargeByBankDetails()::card charge failed due to: ", br.Message)
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: br.Message,
		}
	}
	if br.Data.Status == "failed" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: br.Data.Message,
			ResponseStatus: br.Data.Status,
			Reference: br.Data.Reference,
		}
	}
	if br.Data.Status == "pending" {
		//TODO: insert the reference code into the db and check the tx status after 30 seconds
	}
	if br.Data.Status == "timeout" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: br.Data.Message,
			ResponseStatus: br.Data.Status,
			Reference: br.Data.Reference,
		}
	}
	if br.Data.Status == "send_pin" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: br.Data.DisplayText,
			ResponseStatus: br.Data.Status,
			Reference: br.Data.Reference,
		}
	}
	if br.Data.Status == "send_phone" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: br.Data.DisplayText,
			ResponseStatus: br.Data.Status,
			Reference: br.Data.Reference,
		}
	}
	if br.Data.Status == "send_birthday" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: br.Data.DisplayText,
			ResponseStatus: br.Data.Status,
			Reference: br.Data.Reference,
		}
	}
	if br.Data.Status == "send_otp" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: br.Data.DisplayText,
			ResponseStatus: br.Data.Status,
			Reference: br.Data.Reference,
		}
	}
	if br.Data.Status == "open_url" {
		
	}
	if br.Data.Status == "success" {
		return successChargeMsg(resp,statusCode)
	}
	return nil
}