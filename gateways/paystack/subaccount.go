package paystack

import (
	"fmt"
	"encoding/json"
	g "github.com/kenmobility/feezbot/gateways"
)

func CreateSubAccount(businessName,bankName,accountNumber,contactEmail, contactName, contactPhone string, percentageCharge float64) *g.SubAccountResponse {
	url := "https://api.paystack.co/subaccount"
	postData := fmt.Sprintf(`{
		"business_name": "%s",
		"settlement_bank": "%s",
		"account_number": "%s",
		"percentage_charge": %v,
		"primary_contact_email": "%s",
		"primary_contact_name": "%s",
		"primary_contact_phone": "%s"
	}`,businessName,bankName,accountNumber,percentageCharge,contactEmail,contactName,contactPhone)
	
	resp, statusCode := sendHTTPPostRequest(url, "POST", postData)
	if statusCode == -1 {
		fmt.Println("subaccount.go::CreateSubAccount()::error in making request due to: ", resp)
		return &g.SubAccountResponse {
			StatusCode: statusCode,
			Status : "error",
			ResponseMsg :  resp,
		}
	}
	var psa createSubAccountResponse
	err := json.Unmarshal([]byte(resp), &psa)
	if err != nil {
		fmt.Println("subaccount.go::CreateSubAccount()::error in unmarshalling response due to: ", err)
	}
	if psa.Status == false {
		return &g.SubAccountResponse {
			StatusCode: statusCode,
			Status : "error",
			ResponseMsg :  psa.Message,
		}
	} 
	return &g.SubAccountResponse {
		StatusCode: statusCode,
		Status : "success",
		ResponseMsg :  psa.Message,
		PercentageCharge: psa.Data.PercentageCharge,
		SettlementBank: psa.Data.SettlementBank,
		AccountCode: psa.Data.SubaccountCode,
		AccountNumber: psa.Data.AccountNumber,
		MerchantName: psa.Data.BusinessName,
	}
}