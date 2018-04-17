package paystack

import (
	"fmt"
	"encoding/json"
	g "github.com/kenmobility/feezbot/gateways"
)

func InitializeTransaction(reference, customerEmail, subaccount,bearer,txSubject,txDesc string, amount int) *g.InitializeTxResponse {
	url := "https://api.paystack.co/transaction/initialize"
	postData := fmt.Sprintf(`{
		"reference": "%s", 
		"amount": %v, 
		"email": "%s",
		"subaccount": "%s",
		"bearer":"%s",
		"metadata":{
			"custom_fields":[
				{
					"display_name": "%s",
					"value":"%s"	
				}
			]
		},
		"channels": [
			"card","bank"
		]
	}`,reference,amount * 100, customerEmail,subaccount,bearer,txSubject,txDesc)
	
	resp, statusCode := sendHTTPPostRequest(url, "POST", postData)
	if statusCode == -1 {
		fmt.Println("tx_initialize.go::InitializeTransaction()::error in making request due to: ", resp)
		return &g.InitializeTxResponse {
			StatusCode: statusCode,
			Status : "error",
			ResponseMsg :  resp,
		}
	} 
	var ir paystackInitializeTransactionResponse
	err := json.Unmarshal([]byte(resp), &ir)
	if err != nil {
		fmt.Println("tx_initialize.go::InitializeTransaction()::error in unmarshalling response due to: ", err)
	}
	if ir.Status == false {
		return &g.InitializeTxResponse {
			StatusCode: statusCode,
			Status : "error",
			ResponseMsg :  ir.Message,
		}
	}
	return &g.InitializeTxResponse {
		StatusCode: statusCode,
		Status: "success",
		ResponseMsg:  ir.Message,
		AuthorizationUrl: ir.Data.AuthorizationURL,
		Reference: ir.Data.Reference,
		AccessCode: ir.Data.AccessCode,
	}	
}