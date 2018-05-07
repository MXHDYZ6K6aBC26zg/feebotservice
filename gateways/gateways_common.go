package gateways

type ChargeResponse struct {
	StatusCode 			int
	ResponseBody 		string 
	ResponseStatus 		string
	ResponseMsg 		string
	Reference 			string
	TxAmount			int
	TxFees				int
	TxCurrency			string
	TxDate				string
	TxChannel			string
	TxSubject			string
	TxDescription		string
	AuthorizationCode 	string
	CardLast4			string
	CardType			string
	Email				string
	Bank				string
}

type InitializeTxResponse struct {
	StatusCode 			int
	Status 				string
	ResponseMsg			string
	AuthorizationUrl	string
	AccessCode			string
	Reference			string
}

type SubAccountResponse struct {
	StatusCode				int
	Status 					string
	ResponseMsg				string
	AccountCode				string
	PercentageCharge		float64
	SettlementBank			string
	AccountNumber			string
	MerchantName			string
	MerchantId				string
	MerchantFeeId			string
	SettlementByMerchant	bool
}

type TxVerifyResponse struct {
	StatusCode			int
	Status 				string
	ResponseMsg			string
	ResponseBody 		string 
	ResponseStatus 		string
	GatewayResponse     string
	Reference 			string
	TxAmount			int
	TxFees				float64
	TxCurrency			string
	TxDate				string
	TxChannel			string
	TxSubject			string
	TxDescription		string
	AuthorizationCode 	string
	CardLast4			string
	CardType			string
	Email				string
	Bank				string
}