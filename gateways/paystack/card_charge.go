package paystack

import (
	"fmt"
	"encoding/json"
	"net/http"

	g "github.com/kenmobility/feezbot/gateways"
)

//ChargeByCardDetails is a func that collects the detials of a customer's card and tries to charge it directory
//it returns a struct object
func ChargeByCardDetails(customerEmail, cardNumber,cardCvv,cardExpMonth,cardExpYear string, amount int) *g.ChargeResponse {
	url := "https://api.paystack.co/charge"
	//mainAmount := amount * 100 
	postData := fmt.Sprintf(`{
		"email": "%s",
		"amount": %v,
		"card" : {
			"number": "%s",
			"cvv": "%s",
			"expiry_month": "%s",
			"expiry_year": "%s"
			}
		}`,customerEmail,amount * 100,cardNumber,cardCvv,cardExpMonth,cardExpYear)
	//fmt.Println("the req body is ", postData)
	resp, statusCode := sendHTTPPostRequest(url, "POST", postData)
	if statusCode == -1 {
		fmt.Println("card_charge.go::ChargeByCardDetails()::error in making request due to: ", resp)
		return &g.ChargeResponse {
			StatusCode : http.StatusInternalServerError,
			ResponseMsg :  resp,
			ResponseStatus : "http request failed",
		}
	} 
	var cr paystackChargeResponse
	err := json.Unmarshal([]byte(resp), &cr)
	if err != nil {
		fmt.Println("card_charge.go::ChargeByCardDetails()::error in unmarshalling response due to: ", err)
	}
	if cr.Status == false {
		fmt.Println("card_charge.go::ChargeByCardDetails()::card charge failed due to: ", cr.Message)
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Message,
		}
	}
	if cr.Data.Status == "failed" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Data.Message,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	if cr.Data.Status == "pending" {
		//TODO: insert the reference code into the db and check the tx status after 30 seconds
	}
	if cr.Data.Status == "timeout" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Data.Message,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	if cr.Data.Status == "send_pin" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Data.DisplayText,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	if cr.Data.Status == "send_phone" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Data.DisplayText,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	if cr.Data.Status == "send_birthday" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Data.DisplayText,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	if cr.Data.Status == "send_otp" {
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Data.DisplayText,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	if cr.Data.Status == "open_url" {
		
	}
	if cr.Data.Status == "success" {
		return successChargeMsg(resp,statusCode)
	}
	return nil
}

func successChargeMsg(resp string, statusCode int) *g.ChargeResponse {
	var scr paystackSuccessChargeResponse
		err := json.Unmarshal([]byte(resp), &scr)
		if err != nil {
			fmt.Println("card_charge.go::returnSuccessCardChargeMsg()::error in unmarshalling the success response due to: ", err)
		}
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			//ResponseMsg: scr.Data.Message,
			ResponseStatus: scr.Data.Status,
			Reference: scr.Data.Reference,
			TxAmount: scr.Data.Amount / 100,
			TxFees: scr.Data.Fees / 100,
			TxCurrency: scr.Data.Currency,
			TxDate: scr.Data.TransactionDate,
			TxChannel: scr.Data.Channel,
			TxSubject: scr.Data.Metadata.CustomFields[0].DisplayName,
			TxDescription: scr.Data.Metadata.CustomFields[0].Value,
			AuthorizationCode: scr.Data.Authorization.AuthorizationCode,
			Bank: scr.Data.Authorization.Bank,
			CardLast4: scr.Data.Authorization.Last4,
			CardType: scr.Data.Authorization.CardType,
			Email: scr.Data.Customer.Email,
		}
}

//TokenizeCard is a func that is used to tokenize the ATM card in order to get the authorization code for 
//subsequent charging of the card
func TokenizeCard(customerEmail, cardNumber,cardCvv,cardExpMonth,cardExpYear string) *tokenizeCardResponse {
	url := "https://api.paystack.co/charge/tokenize"
	postData := fmt.Sprintf(`{
		"email": %s,
		"card" : {
			"number": %s,
			"cvv": %s,
			"expiry_month": %s,
			"expiry_year": %s
		}
	}`,customerEmail,cardNumber,cardCvv,cardExpMonth,cardExpYear)
	
	resp, statusCode := sendHTTPPostRequest(url, "POST", postData)
	if statusCode == -1 {
		fmt.Println("card_charge.go::ChargeByCardDetails()::error in making request due to: ", resp)
		return &tokenizeCardResponse {
			Status : "error",
			Message :  resp,
		}
	} 
	var tcr paystackTokenizeCardResponse
	err := json.Unmarshal([]byte(resp), &tcr)
	if err != nil {
		fmt.Println("card_charge.go::ChargeByCardDetails()::error in unmarshalling response due to: ", err)
	}
	if tcr.Status == false {
		return &tokenizeCardResponse {
			Status : "error",
			Message :  tcr.Message,
		}
	} 
	return &tokenizeCardResponse {
		Status: "success",
		Message:  tcr.Message,
		AuthorizationCode: tcr.Data.AuthorizationCode,
	}
}





