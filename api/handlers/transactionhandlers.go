package handlers

import (
	"github.com/kenmobility/feezbot/gateways/paystack"
	"github.com/kenmobility/feezbot/rand"
	//g "github.com/kenmobility/feezbot/gateways"
	"fmt"
	"net/http"
	h "github.com/kenmobility/feezbot/helper"
	"io/ioutil"
	"encoding/json"
	s "strings"
	"errors"
	"time"

	"github.com/labstack/echo"
)

type InitializeTransaction struct {
	InitiateTransaction struct {
		Amount     int      `json:"amount"`
		Channels   []string `json:"channels"`
		FeeID      string   `json:"fee_id"`
		MerchantID string   `json:"merchant_id"`
		MerchantFeeID string   `json:"merchant_fee_id"`
		Metadata   struct {
			CustomFields []struct {
				DisplayName string `json:"display_name"`
				Value       string `json:"value"`
			} `json:"custom_fields"`
		} `json:"metadata"`
		Username string `json:"username"`
	} `json:"initiateTransaction"`
}

//InitiateTransaction is a POST request handler used to initiate a transaction by the user
func InitiateTransaction(c echo.Context) error {
	var it InitializeTransaction
	defer c.Request().Body.Close()
	b,err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Printf("transactionhandlers.go::ChargeUserByCard()::failed to read request body due to : %s\n", err)
		r := h.Response {
			Status: "error",
			Message:"error occured, please try again",//err.Error(),
		}
		return c.JSON(http.StatusBadRequest, r)
	}
	//fmt.Printf("the raw json request is %s\n", b)
	err = json.Unmarshal(b, &it)
	if err != nil {
		fmt.Println("transactionhandlers.go::InitiateTransaction()::failed to unmarshal json request body: ", err)
		r := h.Response {
			Status: "error",
			Message:"error occured, please try again",//err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, r)
	}
	username := it.InitiateTransaction.Username
	merchantId := it.InitiateTransaction.MerchantID
	merchantFeeId := it.InitiateTransaction.MerchantFeeID
	feeId := it.InitiateTransaction.FeeID
	amount := it.InitiateTransaction.Amount

	if username == "" || merchantId == "" || merchantFeeId == "" || feeId == "" || amount <= 0 {
		r := h.Response {
			Status: "error",
			Message:"Invalid request format Or required parameters not complete",
		}
		return c.JSON(http.StatusBadRequest, r)	
	}

	uId,email,emailConfStatus := isEmailConfirmed(username)
	if emailConfStatus == false {
		r := h.Response {
			Status: "error",
			Message:"Your Email address has not yet been confirmed, click 'Confirm My Email' to confirm ur Address before proceeding to make payments",
		}
		return c.JSON(http.StatusForbidden, r)	
	}

	//Generate a unique reference for the transaction
	reference := rand.RandStr(18, "alphanum")
	fmt.Println("generated reference is - ", reference)

	//Get the subaccount code for the merchant / fee 
	subaccount,feeBearer,err := getSettlementAccount(merchantFeeId)
	if err != nil {
		fmt.Printf("transactionhandlers.go::InitiateTransaction()::error encountered trying to get settlement account for merchantFeeId - %s; is %s", merchantFeeId,err)
		r := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusForbidden, r)
	}

	//TODO: call a function to insert the details of the user and the transaction into a table
	_,err = dbInsertUserTransaction(uId, reference, merchantId, feeId, amount)
	if err != nil {
		fmt.Println("error while inserting user transaction detail : ", err)
	}

	res := paystack.InitializeTransaction(reference, email, subaccount, feeBearer, "", "", amount)
	
	bs,_:= json.Marshal(res)
	r := h.Response {
		Status: res.Status,
		Message:res.ResponseMsg,
		Data: bs,
	}
	return c.JSON(res.StatusCode, r)
}

/*
func checkResponseStatus(res *g.ChargeResponse, uId,uEmail,merchantId,feeId string,amount int,channel string) h.Response {
	if s.Contains(res.ResponseStatus, "success") == true { 
		fmt.Println("returned response for card transaction is success")
		_,err := dbinsertSuccessChargeCardResponse(uId,res.Reference,res.Email,res.TxDate,res.ResponseStatus,res.TxCurrency,res.TxChannel,
		res.AuthorizationCode,"Paystack",res.CardLast4,res.ResponseBody,res.Bank,res.CardType,res.StatusCode,res.TxAmount,res.TxFees)

		if err != nil {
			fmt.Println("transactionhandlers.go::checkResponseStatus()::error occured during success card response insert is ",err)
		}
		d := map[string]string {
			"status" : res.ResponseStatus,
			"reference" : res.Reference,
			"message": "transaction is successful",
		}
		bs,_:= json.Marshal(d)
		r := h.Response {
			Status: "success",
			Message:"Your transaction is successful",
			Data: bs,
		}
		return r //c.JSON(http.StatusOK, r)
	}

	if s.Contains(res.ResponseStatus, "pending") == true { 
		fmt.Println("returned response for card transaction is pending")
		_,err := dbinsertChargeCardResponse(uId,res.Reference,res.ResponseStatus,"Paystack",res.ResponseBody,merchantId,feeId,uEmail,res.StatusCode,amount,channel)

		if err != nil {
			fmt.Println("transactionhandlers.go::checkResponseStatus()::error occured during pending card response insert is ",err)
		}
		d := map[string]string {
			"status" : res.ResponseStatus,
			"reference" : res.Reference,
			"message": "transaction is Pending",
		}
		bs,_:= json.Marshal(d)
		r := h.Response {
			Status: "error",
			Message:"Your transaction is pending, you will receive a notification on your phone or email as regards the transaction.",
			Data: bs,
		}
		return r //c.JSON(http.StatusOK, r)
	}

	if s.Contains(res.ResponseStatus, "timeout") == true { 
		fmt.Println("returned response for card transaction is timeout")
		_,err := dbinsertChargeCardResponse(uId,res.Reference,res.ResponseStatus,"Paystack",res.ResponseBody,merchantId,feeId,uEmail,res.StatusCode,amount,channel)

		if err != nil {
			fmt.Println("transactionhandlers.go::checkResponseStatus()::timeout error occured during card response insert is ",err)
		}
		d := map[string]string {
			"status" : res.ResponseStatus,
			"reference" : res.Reference,
			"message": res.ResponseMsg,
		}
		bs,_:= json.Marshal(d)
		r := h.Response {
			Status: "error",
			Message: "transaction is timeout",
			Data: bs,
		}
		return r //c.JSON(http.StatusOK, r)
	}

	if s.Contains(res.ResponseStatus, "send_otp") == true { 
		fmt.Println("returned response for card transaction is send_otp")
		_,err := dbinsertChargeCardResponse(uId,res.Reference,res.ResponseStatus,"Paystack",res.ResponseBody,merchantId,feeId,uEmail,res.StatusCode,amount,channel)

		if err != nil {
			fmt.Println("transactionhandlers.go::checkResponseStatus()::timeout error occured during card response insert is ",err)
		}
		d := map[string]string {
			"status" : res.ResponseStatus,
			"reference" : res.Reference,
			"message": res.ResponseMsg,
		}
		bs,_:= json.Marshal(d)
		r := h.Response {
			Status: "success",
			Message: "transaction sent, an OTP is required",
			Data: bs,
		}
		return r //c.JSON(http.StatusOK, r)
	}

	if s.Contains(res.ResponseStatus, "send_pin") == true { 
		fmt.Println("returned response for transaction is send_pin")
		_,err := dbinsertChargeCardResponse(uId,res.Reference,res.ResponseStatus,"Paystack",res.ResponseBody,merchantId,feeId,uEmail,res.StatusCode,amount,channel)

		if err != nil {
			fmt.Println("transactionhandlers.go::checkResponseStatus()::error occured during card send_pin response insert is ",err)
		}
		d := map[string]string {
			"status" : res.ResponseStatus,
			"reference" : res.Reference,
			"message": res.ResponseMsg,
		}
		bs,_:= json.Marshal(d)
		r := h.Response {
			Status: "success",
			Message: "transaction sent, PIN is required",
			Data: bs,
		}
		return r //c.JSON(http.StatusOK, r)
	}

	if s.Contains(res.ResponseStatus, "send_birthday") == true { 
		fmt.Println("returned response for card transaction is send_birthday")
		_,err := dbinsertChargeCardResponse(uId,res.Reference,res.ResponseStatus,"Paystack",res.ResponseBody,merchantId,feeId,uEmail,res.StatusCode,amount,channel)

		if err != nil {
			fmt.Println("transactionhandlers.go::checkResponseStatus()::error occured during card send_pin response insert is ",err)
		}
		d := map[string]string {
			"status" : res.ResponseStatus,
			"reference" : res.Reference,
			"message": res.ResponseMsg,
		}
		bs,_:= json.Marshal(d)
		r := h.Response {
			Status: "success",
			Message: "transaction sent, birthday is required",
			Data: bs,
		}
		return r //c.JSON(http.StatusOK, r)
	}
	if s.Contains(res.ResponseStatus, "failed") == true { 
		fmt.Println("returned response for transaction is failed")
		// _,err := dbinsertChargeCardResponse(uId,res.Reference,res.ResponseStatus,"Paystack",res.ResponseBody,merchantId,feeId,uEmail,res.StatusCode,amount,channel)

		//if err != nil {
		//	fmt.Println("transactionhandlers.go::checkResponseStatus()::error occured during card send_pin response insert is ",err)
		//} 
		d := map[string]string {
			"status" : res.ResponseStatus,
			"reference" : res.Reference,
			"message": res.ResponseMsg,
		}
		bs,_:= json.Marshal(d)
		r := h.Response {
			Status: "error",
			Message: "transaction failed, please try again",
			Data: bs,
		}
		return r //c.JSON(http.StatusOK, r)
	}
	r := h.Response {
		Status: "error",
		Message: res.ResponseMsg,
		//Data: bs,
	}
	return r
} */

func dbUpdateChargeCardResponse(userId,txReference,txStatus,txPaymentGateway,responseBody,merchantId,feeId,userEmail string, responseCode,amount int,channel string) (string,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
		//return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()
	var insertedTxId string
	insertQuery := `INSERT INTO "payment_transactions"("Id","UserId","TxReference","TxProvidedEmail","TxDate","TxStatus","TxAmount","ResponseBody","ResponseCode","TxChannel","TxPaymentGateway","MerchantId","FeeId") 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),userId,txReference,userEmail,time.Now(),txStatus,amount,responseBody,responseCode,channel,txPaymentGateway,merchantId,feeId).Scan(&insertedTxId)
	if err != nil {
		fmt.Println("transactionhandlers.go::dbinsertChargeCardResponse()::error encountered while inserting into transactions for not success card transaction response is ", err)
		return "",err
	}
	//check if the row was inserted successfully
	if insertedTxId == "" {
		return "", errors.New("insertion into transactions failed")
	}
	return insertedTxId, nil
}

func getSettlementAccount(merchantFeeId string) (string,string, error) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("transactionhandlers.go::getSettlementAccount()::error in connecting to database due to ",err)
		return "","",err
	}
	defer con.Close()
	var code,bearer interface{}
	//var chargeByMerchant bool
	var accountCode,feeBearer string 
	q := `SELECT "merchant_accounts"."AccountCode","merchant_fees"."FeeBearer" FROM "merchant_accounts" 
	INNER JOIN "merchant_fees" ON "merchant_fees"."Id" = "merchant_accounts"."MerchantFeeId" WHERE "merchant_accounts"."MerchantFeeId" = $1 AND "merchant_accounts"."Enabled" = $2` 
	err = con.Db.QueryRow(q, merchantFeeId,true).Scan(&code,&bearer)
	if err != nil {
		fmt.Println("transactionhandlers.go::getSettlementAccount()::error in fetching account code from database due to ",err)
		if s.Contains(fmt.Sprintf("%v", err), "no rows") == true {
			return "","",errors.New("Sorry, selected fee is yet to be enabled")
		}
	}
	if code == nil {
		return "","",errors.New("account code for merchant/fee not yet generated")
	}
	if bearer == nil {
		return "","",errors.New("fee bearer for merchant/fee is empty")
	}
	accountCode = code.(string)
	feeBearer = bearer.(string)
	return accountCode,feeBearer, nil 
}

func dbInsertUserTransaction(uId,reference,merchantId,feeId string, amount int) (string, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
		//return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()

	//txTimeStamp, _ := time.Parse(time.RFC3339,txDate) 
	var insertedTxId string
	insertQuery := `INSERT INTO "payment_transactions"("Id","UserId","TxReference","TxDate","TxAmount","TxPaymentGateway","MerchantId","FeeId") 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),uId,reference,time.Now(),amount,"PayStack",merchantId,feeId).Scan(&insertedTxId)
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