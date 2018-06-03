package paystack

import (
	"fmt"
	"encoding/json"
	s "strings"
	g "github.com/kenmobility/feezbot/gateways"
	//"strconv"
)

func VerifyTransaction(reference string) *g.TxVerifyResponse {
	url := fmt.Sprintf("https://api.paystack.co/transaction/verify/%s", reference)

	resp,statusCode := sendHTTPRequest(url, "GET")
	if statusCode == -1 {
		fmt.Println("verify_tx.go::VerifyTransaction()::error in making request due to: ", resp)
		return &g.TxVerifyResponse {
			StatusCode: statusCode,
			Status : "error",
			ResponseMsg :  resp,
		}
	}
	fmt.Println(resp)
	var vtr verifyTransactionResponse
	err := json.Unmarshal([]byte(resp), &vtr)
	if err != nil {
		fmt.Println("verify_tx.go::VerifyTransaction()::error in unmarshalling response due to: ", err)
	}
	if vtr.Status == false {
		return &g.TxVerifyResponse {
			StatusCode: statusCode,
			Status : "error",
			ResponseMsg :  vtr.Message,
		}
	} 
	var txFees float64

	if vtr.Data.Fees != nil {
		txFees = vtr.Data.Fees.(float64)
	}
	
	if s.Contains(fmt.Sprintf("%v", vtr.Data.Status), "success") == false { 
		return &g.TxVerifyResponse {
			StatusCode: statusCode,
			Status : "error",
			ResponseMsg: vtr.Message,
			ResponseBody: resp,
			ResponseStatus: vtr.Data.Status,
			GatewayResponse: vtr.Data.GatewayResponse,
			Reference: vtr.Data.Reference,
			TxAmount: vtr.Data.Amount,
			TxFees: txFees,
			TxCreatedAt: vtr.Data.TransactionDate,
			PaidAt: vtr.Data.PaidAt,
			TxCurrency: vtr.Data.Currency,
			TxChannel: vtr.Data.Channel,
			AuthorizationCode: vtr.Data.Authorization.AuthorizationCode,
			CardLast4:	vtr.Data.Authorization.Last4,
			CardType: vtr.Data.Authorization.CardType,
			Email:		vtr.Data.Customer.Email,
			Bank: vtr.Data.Authorization.Bank,
		}
	}
	//mainAccountCharge,_ := strconv.Atoi(vtr.Data.FeesSplit.Integration)
	return &g.TxVerifyResponse {
		StatusCode: statusCode,
		Status : "success",
		ResponseMsg: vtr.Message,
		ResponseBody: resp,
		ResponseStatus: vtr.Data.Status,
		GatewayResponse: vtr.Data.GatewayResponse,
		Reference: vtr.Data.Reference,
		TxAmount: vtr.Data.Amount,
		SubAccountSettlementAmount: vtr.Data.FeesSplit.Subaccount,
		MainAccountSettlementAmount: vtr.Data.FeesSplit.Integration,
		TxFeeBearer: vtr.Data.FeesSplit.Params.Bearer,
		PercentageCharged: vtr.Data.FeesSplit.Params.PercentageCharge,
		TxFees: txFees,
		TxCreatedAt: vtr.Data.TransactionDate,
		PaidAt: vtr.Data.PaidAt,
		TxCurrency: vtr.Data.Currency,
		TxChannel: vtr.Data.Channel,
		AuthorizationCode: vtr.Data.Authorization.AuthorizationCode,
		CardLast4:	vtr.Data.Authorization.Last4,
		CardType: vtr.Data.Authorization.CardType,
		Email:		vtr.Data.Customer.Email,
		Bank: vtr.Data.Authorization.Bank,
	}
}