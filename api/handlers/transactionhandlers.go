package handlers

import (
	"github.com/kenmobility/feezbot/gateways/paystack"
	//"github.com/kenmobility/feezbot/rand"
	"fmt"
	"net/http"
	h "github.com/kenmobility/feezbot/helper"
	"encoding/json"
	s "strings"
	"strconv"
	"errors"
	"time"
	"log"
	"github.com/labstack/echo"
	g "github.com/kenmobility/feezbot/gateways"
)

type MyJsonName struct {
	Data []struct {
		Amount        int `json:"amount"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bank              string `json:"bank"`
			Bin               string `json:"bin"`
			Brand             string `json:"brand"`
			CardType          string `json:"card_type"`
			Channel           string `json:"channel"`
			CountryCode       string `json:"country_code"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Last4             string `json:"last4"`
			Reusable          bool   `json:"reusable"`
			Signature         string `json:"signature"`
		} `json:"authorization"`
		Channel   string `json:"channel"`
		CreatedAt string `json:"createdAt"`
		Created_At string `json:"created_at"`
		Currency  string `json:"currency"`
		Customer  struct {
			CustomerCode string      `json:"customer_code"`
			Email        string      `json:"email"`
			FirstName    interface{} `json:"first_name"`
			ID           int         `json:"id"`
			LastName     interface{} `json:"last_name"`
			Metadata     interface{} `json:"metadata"`
			Phone        interface{} `json:"phone"`
			RiskAction   string      `json:"risk_action"`
		} `json:"customer"`
		Domain          string      `json:"domain"`
		Fees            int         `json:"fees"`
		FeesSplit       string      `json:"fees_split"`
		GatewayResponse string      `json:"gateway_response"`
		ID              int         `json:"id"`
		IPAddress       interface{} `json:"ip_address"`
		Log             interface{} `json:"log"`
		Message         interface{} `json:"message"`
		Metadata        struct {
			CustomFields []struct {
				DisplayName  string `json:"display_name"`
				Value        string `json:"value"`
				VariableName string `json:"variable_name"`
			} `json:"custom_fields"`
		} `json:"metadata"`
		PaidAt     string   `json:"paidAt"`
		Paid_At     string   `json:"paid_at"`
		Plan       struct{} `json:"plan"`
		Reference  string   `json:"reference"`
		Status     string   `json:"status"`
		Subaccount struct{} `json:"subaccount"`
	} `json:"data"`
	Message string `json:"message"`
	Meta    struct {
		Page        int `json:"page"`
		PageCount   int `json:"pageCount"`
		PerPage     int `json:"perPage"`
		Skipped     int `json:"skipped"`
		Total       int `json:"total"`
		TotalVolume int `json:"total_volume"`
	} `json:"meta"`
	Status bool `json:"status"`
}


/*InitiateTransaction is a POST request handler used to initiate a transaction by the user
func InitiateTransaction(c echo.Context) error {
	userId := s.Trim(c.FormValue("userId")," ")
	merchantId := s.Trim(c.FormValue("merchantId")," ")
	merchantFeeId := s.Trim(c.FormValue("merchantFeeId")," ")
	feeId := s.Trim(c.FormValue("feeId")," ")
	amount := c.FormValue("amount")
	paymentReferenceName := c.FormValue("paymentReferenceName")
	paymentReferenceId := c.FormValue("paymentReferenceId")
	categoryName := s.Trim(c.FormValue("categoryName")," ")

	if userId == "" || merchantId == "" || merchantFeeId == "" || feeId == "" || amount == "" || categoryName == "" || paymentReferenceName == "" || paymentReferenceId == ""{
		r := h.Response {
			Status: "error",
			Message:"Invalid request format Or required parameters not complete",
		}
		return c.JSON(http.StatusBadRequest, r)	
	}
	floatAmount, err := strconv.ParseFloat(amount, 64)
	intAmount := int(floatAmount)
	if intAmount <= 0 || err != nil {
		fmt.Println("error occured trying to convert amount string to integer is :", err)
		r := h.Response {
			Status: "error",
			Message: "Amount can not be less than or equal to zero",	
		}
		return c.JSON(http.StatusForbidden, r)
	}
	email,_,_ := isEmailAndPhoneConfirmed(userId)
	if emailConfStatus == false {
		r := h.Response {
			Status: "error",
			Message:"Your email address has not yet been confirmed. Please confirm your email to proceed with payment",
		}
		return c.JSON(http.StatusForbidden, r)	
	}
	if phoneConfStatus == false {
		r := h.Response {
			Status: "error",
			Message:"Your phone number is yet to be verified. Please verify your phone number to proceed with payment",
		}
		return c.JSON(http.StatusForbidden, r)	
	} 
	//Generate a unique reference for the transaction
	reference := rand.RandStr(18, "alphanum")
	//fmt.Println("generated reference is - ", reference)

	//Get the subaccount code for the merchant / fee 
	merchantName,subaccount,feeTitle,feeBearer,_,err := getSettlementAccount(merchantFeeId)
	if err != nil {
		fmt.Printf("transactionhandlers.go::InitiateTransaction()::error encountered trying to get settlement account for merchantFeeId - %s; is %s", merchantFeeId,err)
		r := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusForbidden, r)
	}

	res := paystack.InitializeTransaction(reference, email, subaccount, feeBearer, paymentReferenceName,paymentReferenceId,categoryName,merchantName,feeTitle,intAmount)
	
	bs,_:= json.Marshal(res)
	r := h.Response {
		Status: res.Status,
		Message:res.ResponseMsg,
		Data: bs,
	}
	//TODO: call a function to insert the details of the user and the transaction into a table
	_,err = dbInsertUserTransaction(userId, reference,categoryName,merchantId,feeId,paymentReferenceName,paymentReferenceId,res.AuthorizationUrl,res.AccessCode,intAmount)
	if err != nil {
		fmt.Println("error while inserting user transaction detail : ", err)
	}
	 
	return c.JSON(res.StatusCode, r)
}
*/

func InitiatePaymentTransaction(c echo.Context) error {
	userId := s.Trim(c.FormValue("userId")," ")
	merchantId := s.Trim(c.FormValue("merchantId")," ")
	merchantFeeId := s.Trim(c.FormValue("merchantFeeId")," ")
	feeId := s.Trim(c.FormValue("feeId")," ")
	amount := c.FormValue("amount")
	paymentReferenceName := c.FormValue("paymentReferenceName")
	paymentReferenceId := c.FormValue("paymentReferenceId")
	categoryName := s.Trim(c.FormValue("categoryName")," ")
	academicSession := s.Trim(c.FormValue("academicSession")," ")
	academicSemester := s.Trim(c.FormValue("academicSemester")," ")
	vReference := s.Trim(c.FormValue("virtualReference")," ")

	if userId == "" || merchantId == "" || merchantFeeId == "" || feeId == "" || amount == "" || categoryName == "" || paymentReferenceName == "" || paymentReferenceId == "" || vReference == ""{
		r := h.Response {
			Status: "error",
			Message:"Invalid request format Or required parameters not complete",
		}
		return c.JSON(http.StatusBadRequest, r)	
	}
	floatAmount, err := strconv.ParseFloat(amount, 64)
	intAmount := int(floatAmount)
	if intAmount <= 0 || err != nil {
		fmt.Println("error occured trying to convert amount string to integer is :", err)
		r := h.Response {
			Status: "error",
			Message: "Amount can not be less than or equal to zero",	
		}
		return c.JSON(http.StatusForbidden, r)
	}
	email,emailConfStatus,phoneConfStatus := isEmailAndPhoneConfirmed(userId)
	if emailConfStatus == false {
		r := h.Response {
			Status: "error",
			Message:"Your email address has not yet been confirmed. Please confirm your email to proceed with payment",
		}
		return c.JSON(http.StatusForbidden, r)	
	}
	if phoneConfStatus == false {
		r := h.Response {
			Status: "error",
			Message:"Your phone number is yet to be verified. Please verify your phone number to proceed with payment",
		}
		return c.JSON(http.StatusForbidden, r)	
	}
	var feeBearer string
	var flatFee float64
	//Get the subaccount code for the merchant / fee 
	_,subaccount,_,_,percCharge,err := getSettlementAccount(merchantFeeId)
	if err != nil {
		fmt.Printf("transactionhandlers.go::InitiatePaymentTransaction()::error encountered trying to get settlement account for merchantFeeId - %s; is %s", merchantFeeId,err)
		r := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusForbidden, r)
	}
	//TODO: call a function to insert the details of the user and the transaction into a table
	go dbInsertUserTransaction(userId, vReference,categoryName,merchantId,feeId,paymentReferenceName,paymentReferenceId,academicSession,academicSemester,intAmount)
	/* if err != nil {
		fmt.Println("error while inserting user transaction detail : ", err)
	} */
	//check if amount is less than 2500 Naira
	if intAmount < 2500 {
		fee := (percCharge / 100) * (floatAmount - 100)
		flatFee = fee + 100
		feeBearer = "account"
	}else{
		//if amount is greater than or equal to 2500 Naira
		flatFee = amountByPercentageCharge(floatAmount, percCharge)
		feeBearer = "account"
	}
	pDetail := map[string]interface{} {
		"email": email,
		"subaccount_code": subaccount,
		"fee_bearer": feeBearer,
		"virtual_reference": vReference,
		"transaction_charge": int(flatFee) * 100,
	}
	bs,_:= json.Marshal(pDetail)
	r := h.Response {
		Status: "success",
		Message:"Transaction details recorded",
		Data: bs,
	}
	return c.JSON(http.StatusOK, r)
}

func VerifyTransaction(c echo.Context) error {
	txReference := c.QueryParam("txReference")
	vReference := c.QueryParam("virtualReference")
	if txReference == "" {
		log.Println("no tx reference found")
		r := h.Response {
			Status: "error",
			Message:"no reference found",
		}
	 	return c.JSON(http.StatusNotFound, r)
	}
	if vReference == "" {
		log.Println("no virtual reference found")
		r := h.Response {
			Status: "error",
			Message:"no virtual reference found",
		}
		return c.JSON(http.StatusNotFound, r)
    }
	log.Println("tx reference is ", txReference)
	log.Println("virtual reference is ", vReference)
	
	resp := paystack.VerifyTransaction(txReference)
	if resp.StatusCode != 200 {
		fmt.Printf("query to verify transaction with reference %s failed due to %s\n", txReference, resp.ResponseMsg)
		r := h.Response {
			Status: "error",
			Message: fmt.Sprintf("query to verify transaction with reference %s failed due to %s\n", txReference, resp.ResponseMsg),
		  }
		return c.JSON(http.StatusBadRequest, r)
	}
	//fmt.Printf("%+v",resp)
	/* if resp.ResponseStatus != "success" {
		r := h.Response {
			Status: "success",
			Message: fmt.Sprintf("Payment transaction with reference - %s failed due to %s \n",txReference,resp.GatewayResponse),
		}
		return c.JSON(http.StatusOK, r)
	} */
	//fmt.Println("tx id is ", txId)
	//TODO: calculate the allocation for associate account(s) based on the settlement merchant 
	go runDbUpdateInfo(txReference,vReference,resp)
	r := h.Response {
	  Status: resp.ResponseStatus,
	  Message: fmt.Sprintf("Payment transaction with reference %s: %s => [%s].\n DO YOU WISH TO MAKE ANOTHER PAYMENT?\n",txReference,resp.ResponseStatus,resp.GatewayResponse),
	}
	return c.JSON(http.StatusOK, r)
}

func runDbUpdateInfo(txReference,vReference string,resp *g.TxVerifyResponse) error {
	var updatedStatus bool
	var txId,merchantId string
	var err error
	q := `SELECT "IsUpdated" FROM "payment_transactions" WHERE "VirtualReference"= $1`
	uStatus,_ := h.DBSelect(q,vReference)
	if uStatus != nil {
		updatedStatus = uStatus.(bool)
	}	  
	if updatedStatus == false {
		txId,merchantId,err = dbUpdateChargeResponse(txReference,vReference,resp.Email,resp.TxCreatedAt,resp.PaidAt,resp.ResponseStatus,resp.TxCurrency,resp.TxChannel,resp.AuthorizationCode,resp.CardLast4,resp.ResponseBody,
				resp.Bank,resp.CardType,resp.GatewayResponse,resp.TxFeeBearer,resp.PercentageCharged,resp.SubAccountSettlementAmount.(float64),resp.MainAccountSettlementAmount.(float64),resp.StatusCode,resp.TxAmount,resp.TxFees)
		if err != nil {
			fmt.Println("error encountered while updating payment_transactions table is ", err)
		}
	}
	associateSettlement(resp.TxAmount / 100, merchantId, txId)
	contributorSettlement(resp.TxAmount / 100, txId)

	return nil
}

func getSettlementAccount(merchantFeeId string) (string,string,string,string,float64, error) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("transactionhandlers.go::getSettlementAccount()::error in connecting to database due to ",err)
		return "","","","",-1.0,err
	}
	defer con.Close()
	var code,bearer,iMerchantName,iFeeName, ifeeChargePerc interface{}
	var accountCode,feeBearer,sMerchantName,sFeeName string 
	var feeChargePercFloat float64
	q := `SELECT "merchant_accounts"."AccountCode","merchant_fees"."FeeBearer","merchant_fees"."PercentageChargeByFee","merchants"."Title", get_fee_title("merchant_fees"."FeeId") FROM "merchant_accounts" 
	INNER JOIN "merchant_fees" ON "merchant_fees"."Id" = "merchant_accounts"."MerchantFeeId" INNER JOIN "merchants" ON "merchants"."Id" = "merchant_accounts"."MerchantId" WHERE "merchant_accounts"."MerchantFeeId" = $1 AND "merchant_accounts"."Enabled" = $2` 
	err = con.Db.QueryRow(q, merchantFeeId,true).Scan(&code,&bearer,&ifeeChargePerc,&iMerchantName,&iFeeName)
	if err != nil {
		fmt.Println("transactionhandlers.go::getSettlementAccount()::error in fetching account code from database due to ",err)
		if s.Contains(fmt.Sprintf("%v", err), "no rows") == true {
			return "","","","",-1.0,errors.New("Sorry, selected fee is yet to be enabled")
		}
	}
	if code == nil {
		return "","","","",-1.0,errors.New("account code for merchant/fee not yet generated")
	}
	if bearer == nil {
		return "","","","",-1.0,errors.New("fee bearer for merchant/fee is empty")
	}
	if iMerchantName != nil {
		sMerchantName = iMerchantName.(string)
	}
	if iFeeName != nil {
		sFeeName = iFeeName.(string)
	}
	if ifeeChargePerc != nil {
		feeChargePercFloat = ifeeChargePerc.(float64)
	}
	accountCode = code.(string)
	feeBearer = bearer.(string)
	return sMerchantName,accountCode,sFeeName,feeBearer,feeChargePercFloat, nil 
}

func dbInsertUserTransaction(uId,vReference,categoryName,merchantId,feeId,referenceName,referenceId,academicSession,academicSemester string, amount int) (string, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()

	var insertedTxId string
	insertQuery := `INSERT INTO "payment_transactions"("Id","UserId","VirtualReference","TxDate","TxAmount","TxPaymentGateway","MerchantId","FeeId","CategoryName","TxPaymentReferenceName","TxPaymentReferenceId","AcademicSession","AcademicSemester") 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),uId,vReference,time.Now(),amount,"PayStack",merchantId,feeId,categoryName,referenceName,referenceId,academicSession,academicSemester).Scan(&insertedTxId)
	if err != nil {
		fmt.Println("transactionhandlers.go::dbInsertUserTransaction()::error encountered while inserting into payment_transactions : ", err)
		return "",err
	}
	//check if the row was inserted successfully
	if insertedTxId == "" {
		return "", errors.New("inserting into payment_transactions failed")
	}
	return insertedTxId, nil	
}

func dbUpdateChargeResponse(txReference,vReference,txEmail,txDate,paidAt,txStatus,txCurrency,txChannel,txAuthCode,cardLast4,responseBody, bank,cardType,gatewayResponse,feeBearer,percentageCharged string,subAccountSettlementAmount,mainAccountSettlementAmount float64, 
	responseCode,txAmount int, txFee float64) (string,string,error) {	
	con, err := h.OpenConnection()
	if err != nil {
		return "","",err
	}
	defer con.Close()

	var updatedTxId,merchantId string
	updateQuery := `UPDATE "payment_transactions" SET "TxProvidedEmail" = $1, "TxCreatedAt" = $2, "TxStatus" = $3, "AmountPaid" = $4, "ResponseBody" = $5, "ResponseCode" = $6,"TxCurrency" = $7, "TxChannel" = $8,"TxAuthorizationCode" = $9 ,"CardLast4" = $10, "GatewayResponse"= $11, "TxFees" = $12,"Bank" = $13,"CardType" = $14,"PaidAt" = $15,"TxFeeBearer" = $16, "PercentageCharged" = $17, "SubAccountSettlementAmount" = $18, "MainAccountSettlementAmount" = $19, "IsUpdated" = $20,"GatewayReference" = $21 WHERE "VirtualReference" = $22  RETURNING "Id","MerchantId"`
	err = con.Db.QueryRow(updateQuery,txEmail,txDate,txStatus,txAmount / 100,responseBody,responseCode,txCurrency,txChannel,txAuthCode,cardLast4,gatewayResponse,txFee / 100,bank,cardType,paidAt,feeBearer,percentageCharged,subAccountSettlementAmount / 100,mainAccountSettlementAmount / 100,true,txReference,vReference).Scan(&updatedTxId,&merchantId)
	if err != nil {
		fmt.Println("transactionhandlers.go::dbUpdateChargeResponse()::error encountered while inserting into transactions for success card response is ", err)
		return "","",err
	}
	//check if the row was inserted successfully
	if updatedTxId == "" {
		return "","", errors.New("updating payment_transactions failed")
	} 
	return updatedTxId,merchantId, nil
} 

func amountByPercentageCharge(amount,percCharge float64) float64 {
	chargeAmount := (percCharge / 100) * (amount - 100)
	return chargeAmount + 100
}

type txDetail struct {
	Merchant	  string `json:"merchant"`
	Category      string `json:"category"`
	Date          string `json:"date"`
	ReferenceID   string `json:"reference_id"`
	ReferenceName string `json:"reference_name"`
	Status		  string  `json:"status"`
	ResponseMessage string  `json:"response_message"`
}
type transactionLists struct {
	FeeTitle    string `json:"fee_title"`
	TxReference string `json:"tx_reference"`
	Amount  int64 `json:"amount"`
	Details txDetail `json:"details"`
}
type UserTransactions struct {
	UserTx []transactionLists  `json:"transaction_lists"`
}

func TransactionList(c echo.Context) error {
	userId := s.Trim(c.FormValue("userId")," ")
	limit := c.FormValue("limit")
	search := s.Title(c.FormValue("search"))
	if userId == "" {
		r := h.Response {
			Status: "error",
			Message:"'UserId' not supplied",
		}
		return c.JSON(http.StatusBadRequest, r)	
	}
	if limit == "" { 
		r := h.Response {
			Status: "error",
			Message:"'limit' parameter can not be empty",
		}
		return c.JSON(http.StatusBadRequest, r)	
	}
	con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()
	transactions := make([]transactionLists,0)
	var iTxReference,iAmount,iReferenceName,iReferenceId,iPaidAt,iCategoryName,iStatus,iResponse,iFeeTitle,iMerchant interface{}
	var sTxReference, sReferenceName,sReferenceId,sCategoryName,sStatus,sResponse,sFeeTitle,sMerchant string
	var paidAt time.Time
	var amountPaid int64
	q := `SELECT "payment_transactions"."GatewayReference","payment_transactions"."AmountPaid","payment_transactions"."TxPaymentReferenceName","payment_transactions"."TxPaymentReferenceId","payment_transactions"."PaidAt","payment_transactions"."CategoryName","payment_transactions"."TxStatus","payment_transactions"."GatewayResponse","fees"."Title","merchants"."Title" FROM "public"."payment_transactions" INNER JOIN "fees" ON "fees"."Id" = "payment_transactions"."FeeId" INNER JOIN "merchants" ON "merchants"."Id" = "payment_transactions"."MerchantId" where "payment_transactions"."UserId" = $1 AND "IsUpdated" = $2 AND "fees"."Title" LIKE '%` + search +`%' ORDER BY "payment_transactions"."PaidAt" DESC LIMIT $3`
	txRows,err := con.Db.Query(q,userId,true,limit)
	defer txRows.Close()
	if err != nil {
		if s.Contains(fmt.Sprintf("%v", err), "no records") == true {
			res := h.Response {
				Status: "error",
				Message:"No record found for the search value",
			}
			return c.JSON(http.StatusOK, res)	
		}
		fmt.Println("transactionhandlers.go::TransactionList()::error in fetching transaction list from payment_transactions due to ",err)
		res := h.Response {
			Status:  "error",
			Message: "Error occured, please try again!",
		}	
		return c.JSON(http.StatusInternalServerError, res)		
	}
	for txRows.Next() {
		err = txRows.Scan(&iTxReference,&iAmount,&iReferenceName,&iReferenceId,&iPaidAt,&iCategoryName,&iStatus,&iResponse,&iFeeTitle,&iMerchant)
		if err != nil {
			fmt.Println("transactionhandlers.go::TransactionList()::error in storing transaction list values from payment_transactions due to ",err)
			res := h.Response {
				Status: "error",
				Message:"Error occured, please try again!",
			}	
			return c.JSON(http.StatusInternalServerError, res)	
		}
		if iTxReference != nil {
			sTxReference = iTxReference.(string)
		}
		if iReferenceName != nil {
			sReferenceName = iReferenceName.(string)
		}
		if iReferenceId != nil {
			sReferenceId = iReferenceId.(string)
		}
		if iPaidAt != nil {
			paidAt = iPaidAt.(time.Time)
		}
		if iCategoryName != nil {
			sCategoryName = iCategoryName.(string)
		}
		if iStatus != nil {
			sStatus = iStatus.(string)
		}
		if iResponse != nil {
			sResponse = iResponse.(string)
		}
		if iFeeTitle != nil {
			sFeeTitle = iFeeTitle.(string)
		}
		if iMerchant != nil {
			sMerchant = iMerchant.(string)
		}
		if iAmount != nil {
			amountPaid = iAmount.(int64)
		}

		txDetail := txDetail	{
			Merchant: sMerchant,	 
			Category: sCategoryName,    
			Date: paidAt.Format("2006-01-02 15:04:05.99"),      
			ReferenceID: sReferenceId,  
			ReferenceName: sReferenceName,
			Status: sStatus,  
			ResponseMessage: sResponse,
		}
		txLists := transactionLists {
			FeeTitle: sFeeTitle,
			TxReference: sTxReference,
			Amount: amountPaid,
			Details: txDetail,
		}
		transactions = append(transactions,txLists)
	}
	lists := UserTransactions {
		UserTx: transactions,
	}

	bs,_:= json.Marshal(lists)
	res := h.Response {
		Status: "success",
		Message: "User transactions fetched successfully",
		Data: bs,
	}
	return c.JSON(http.StatusOK,res)
}

func associateSettlement(txAmount int, merchantId,txId string) error {
	fmt.Println(".......................calculating associate settlement amount.........................")
	defer fmt.Println("...........................end calculation.............................")
	con, err := h.OpenConnection()
	if err != nil {
		return err
	}
	defer con.Close()
	var payablePercentage float64
	var insertedId,associateId string
	queryAssociateCount := `SELECT count(*) FROM "associate_merchant_accounts" WHERE "MerchantId" = $1`
	aCount,err := h.DBSelect(queryAssociateCount, merchantId)
	if err != nil {
		fmt.Println("transactionhandlers.go::associateSettlement()::error in getting the count of associates of a merchant due to ",err)
	}
	fmt.Printf("transactionhandlers.go::associateSettlement():: selected count of associates with merchant Id as %s is %v\n",merchantId,aCount)
	if aCount.(int64) == 0 {
		//this implies that the merchant does not have any associate 
		fmt.Println("transactionhandlers.go::associateSettlement():: selected count of associates is 0")
		return nil 
	}
	if aCount.(int64) == 1 {
		fmt.Println("this implies that the merchant has only one (1) associate")
		aq := `SELECT "UserId","PayablePercentage" FROM "associate_merchant_accounts" WHERE "MerchantId" = $1`
		err := con.Db.QueryRow(aq,merchantId).Scan(&associateId,&payablePercentage)
		if err != nil {
			fmt.Println("transactionhandlers.go::associateSettlement()::error in getting associate merchant info due to ",err)
		}
		payAmount := (payablePercentage / 100) * (float64(txAmount - 100))
		fmt.Printf("pay amount for associate %s = %v \n",associateId,payAmount)
		insertQuery := `INSERT INTO "remuneration"("Id","UserId","PayablePercentage","SettlementAmount","TransactionId") 
		VALUES($1,$2,$3,$4,$5) RETURNING "Id"`
		err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),associateId,payablePercentage,payAmount,txId).Scan(&insertedId)
		if err != nil {
			fmt.Println("transactionhandlers.go::associateSettlement()::error encountered while inserting into remuneration table due to : ", err)
			return err
		}
		return nil
	}
	//TODO: fetch the associate(s) of the merchant and their payable percentage using the merchant id
	rowsQ := `SELECT "UserId","PayablePercentage" FROM "associate_merchant_accounts" WHERE "MerchantId" = $1`
	AssociateRows,err := con.Db.Query(rowsQ, merchantId)
	if err != nil {
		fmt.Println("transactionhandlers.go::associateSettlement()::error in getting associates merchant info due to ",err)
	}
	defer AssociateRows.Close()
	//since associates for the merchant > 1; THEN split their payable percentage according to the number of associates fetched
	for AssociateRows.Next() {
		err := AssociateRows.Scan(&associateId,&payablePercentage)
		if err != nil {
			fmt.Println("transactionhandlers.go::associateSettlement()::error in storing associate_merchant_accounts values due to ",err)
		}
		payAmount := (((payablePercentage / float64(aCount.(int64))) / 100) * float64(txAmount - 100))
		fmt.Printf("pay amount for associate %s = %v \n",associateId,payAmount)
		insertQuery := `INSERT INTO "remuneration"("Id","UserId","PayablePercentage","SettlementAmount","TransactionId") 
		VALUES($1,$2,$3,$4,$5) RETURNING "Id"`
		err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),associateId,payablePercentage / float64(aCount.(int64)),payAmount,txId).Scan(&insertedId)
		if err != nil {
			fmt.Println("transactionhandlers.go::associateSettlement()::error encountered while inserting into remuneration table due to : ", err)
			return err
		}		
	}
	return nil 
}

func contributorSettlement(txAmount int, txId string) error {
	fmt.Println("......................calculating contributor(s) settlement amount..............................")
	defer fmt.Println("..........................end calculation.............................")

	con, err := h.OpenConnection()
	if err != nil {
		return err
	}
	defer con.Close()
	var payablePercentage float64
	var insertedId,contributorId string
	cRowsQ := `SELECT "UserId","PayablePercentage" FROM "contributors_account" WHERE "Enabled" = $1`
	ContributorRows,err := con.Db.Query(cRowsQ, true)
	if err != nil {
		fmt.Println("transactionhandlers.go::contributorSettlement()::error in getting contributors_account table rows due to ",err)
	}
	defer ContributorRows.Close()

	for ContributorRows.Next() {
		err := ContributorRows.Scan(&contributorId,&payablePercentage)
		if err != nil {
			fmt.Println("transactionhandlers.go::contributorSettlement()::error in storing contributors_account table values due to ",err)
		}
		payAmount := ((payablePercentage / 100) * float64(txAmount - 100))
		insertQuery := `INSERT INTO "remuneration"("Id","UserId","PayablePercentage","SettlementAmount","TransactionId") 
		VALUES($1,$2,$3,$4,$5) RETURNING "Id"`
		err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),contributorId,payablePercentage,payAmount,txId).Scan(&insertedId)
		if err != nil {
			fmt.Println("transactionhandlers.go::contributorSettlement()::error encountered while inserting into remuneration due to : ", err)
			return err
		}
		fmt.Printf("%s contributor pay amount is = %v\n",contributorId, payAmount)
	}
	return nil
}