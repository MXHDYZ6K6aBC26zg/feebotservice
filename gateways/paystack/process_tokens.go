package paystack

import (
	"fmt"
	"encoding/json"
	"net/http"

	g "github.com/kenmobility/feezbot/gateways"
)

//ProcessPin is a function that is used to process the pin input of the customer during 
//during the card or bank charge request
func ProcessPin(pin, reference string) *g.ChargeResponse{
	url := "https://api.paystack.co/charge/submit_pin"
	postData := fmt.Sprintf(`{
		"pin": "%s",
		"reference": "%s" 
	}`,pin,reference)
	
	resp, statusCode := sendHTTPPostRequest(url, "POST", postData)
	if statusCode == -1 {
		fmt.Println("process_tokens.go::ProcessPin()::error in making request due to: ", resp)
		return &g.ChargeResponse {
			StatusCode : http.StatusInternalServerError,
			ResponseMsg :  resp,
			ResponseStatus : "http request failed",
		}
	} 
	var cr paystackChargeResponse
	err := json.Unmarshal([]byte(resp), &cr)
	if err != nil {
		fmt.Println("process_tokens.go::ProcessPin()::error in unmarshalling response due to: ", err)
	}
	if cr.Status == false {
		fmt.Println("process_tokens.go::ProcessPin()::proccess pin failed due to: ", cr.Message)
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Message,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	return checkStatus(cr, resp, statusCode)
}

//ProcessOtp is a function that is used to process the OTP input of the customer during 
//the card or bank charge request
func ProcessOtp(otp, reference string) *g.ChargeResponse{
	url := "https://api.paystack.co/charge/submit_otp"
	postData := fmt.Sprintf(`{
		"otp": "%s",
		"reference": "%s" 
	}`,otp,reference)
	
	resp, statusCode := sendHTTPPostRequest(url, "POST", postData)
	if statusCode == -1 {
		fmt.Println("process_tokens.go::ProcessOtp()::error in making request due to: ", resp)
		return &g.ChargeResponse {
			StatusCode : http.StatusInternalServerError,
			ResponseMsg :  resp,
			ResponseStatus : "http request failed",
		}
	} 
	var cr paystackChargeResponse
	err := json.Unmarshal([]byte(resp), &cr)
	if err != nil {
		fmt.Println("process_tokens.go::ProcessOtp()::error in unmarshalling response due to: ", err)
	}
	if cr.Status == false {
		//fmt.Printf("proccess otp raw response is %+v\n", cr)
		fmt.Println("process_tokens.go::ProcessOtp()::proccess Otp failed due to: ", cr.Data.Message)
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Data.Message,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	return checkStatus(cr, resp, statusCode)
}

//ProcessPhone is a function that is used to process the phone number input of the customer during 
//the card or bank charge request
func ProcessPhone(phone, reference string) *g.ChargeResponse{
	url := "https://api.paystack.co/charge/submit_phone"
	postData := fmt.Sprintf(`{
		"phone": "%s",
		"reference": "%s" 
	}`,phone,reference)
	
	resp, statusCode := sendHTTPPostRequest(url, "POST", postData)
	if statusCode == -1 {
		fmt.Println("process_tokens.go::ProcessPhone()::error in making request due to: ", resp)
		return &g.ChargeResponse {
			StatusCode : http.StatusInternalServerError,
			ResponseMsg :  resp,
			ResponseStatus : "http request failed",
		}
	} 
	var cr paystackChargeResponse
	err := json.Unmarshal([]byte(resp), &cr)
	if err != nil {
		fmt.Println("process_tokens.go::ProcessPhone()::error in unmarshalling response due to: ", err)
	}
	if cr.Status == false {
		fmt.Println("process_tokens.go::ProcessPhone()::proccess phone failed due to: ", cr.Data.Message)
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg:  cr.Data.Message,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	return checkStatus(cr, resp, statusCode)
}

func ProcessBirthday(birthday, reference string) *g.ChargeResponse{
	url := "https://api.paystack.co/charge/submit_birthday"
	postData := fmt.Sprintf(`{
		"birthday": "%v",
		"reference": "%s" 
	}`,birthday,reference)
	
	resp, statusCode := sendHTTPPostRequest(url, "POST", postData)
	if statusCode == -1 {
		fmt.Println("process_tokens.go::ProcessBirthday()::error in making request due to: ", resp)
		return &g.ChargeResponse {
			StatusCode : http.StatusInternalServerError,
			ResponseMsg :  resp,
			ResponseStatus : "http request failed",
		}
	} 
	var cr paystackChargeResponse
	err := json.Unmarshal([]byte(resp), &cr)
	if err != nil {
		fmt.Println("process_tokens.go::ProcessBirthday()::error in unmarshalling response due to: ", err)
	}
	if cr.Status == false {
		fmt.Println("process_tokens.go::ProcessBirthday()::proccess birthday failed due to: ", cr.Data.Message)
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Data.Message,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	return checkStatus(cr, resp, statusCode)
}


func checkStatus(cr paystackChargeResponse, resp string, statusCode int) *g.ChargeResponse {
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
		//fmt.Println("status is pending")
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			ResponseMsg: cr.Data.Message,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
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
		return &g.ChargeResponse {
			ResponseBody: resp,
			StatusCode: statusCode,
			//ResponseMsg: cr.Data.DisplayText,
			ResponseStatus: cr.Data.Status,
			Reference: cr.Data.Reference,
		}
	}
	return nil
}