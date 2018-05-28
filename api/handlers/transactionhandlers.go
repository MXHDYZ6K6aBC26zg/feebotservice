package handlers

import (
	"github.com/kenmobility/feezbot/gateways/paystack"
	"github.com/kenmobility/feezbot/rand"
	"fmt"
	"net/http"
	h "github.com/kenmobility/feezbot/helper"
	"encoding/json"
	s "strings"
	"strconv"
	"errors"
	"time"

	"github.com/labstack/echo"
)

//InitiateTransaction is a POST request handler used to initiate a transaction by the user
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
	intAmount := int(floatAmount)//strconv.Atoi(amount)
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
			Message:"Your Email address has not yet been confirmed, click 'Confirm My Email' to confirm ur Address before proceeding to make payments",
		}
		return c.JSON(http.StatusForbidden, r)	
	}
	if phoneConfStatus == false {
		r := h.Response {
			Status: "error",
			Message:"Your phone number is yet to be verified, click 'Verify Phone Number' to verify your phone number before proceeding to make payments",
		}
		return c.JSON(http.StatusForbidden, r)	
	}
	//Generate a unique reference for the transaction
	reference := rand.RandStr(18, "alphanum")
	//fmt.Println("generated reference is - ", reference)

	//Get the subaccount code for the merchant / fee 
	merchantName,subaccount,feeTitle,feeBearer,err := getSettlementAccount(merchantFeeId)
	if err != nil {
		fmt.Printf("transactionhandlers.go::InitiateTransaction()::error encountered trying to get settlement account for merchantFeeId - %s; is %s", merchantFeeId,err)
		r := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusForbidden, r)
	}

	//TODO: call a function to insert the details of the user and the transaction into a table
	_,err = dbInsertUserTransaction(userId, reference,categoryName,merchantId,feeId,paymentReferenceName,paymentReferenceId,intAmount)
	if err != nil {
		fmt.Println("error while inserting user transaction detail : ", err)
	}

	res := paystack.InitializeTransaction(reference, email, subaccount, feeBearer, paymentReferenceName,paymentReferenceId,categoryName,merchantName,feeTitle,intAmount)
	
	bs,_:= json.Marshal(res)
	r := h.Response {
		Status: res.Status,
		Message:res.ResponseMsg,
		Data: bs,
	}
	return c.JSON(res.StatusCode, r)
}

func getSettlementAccount(merchantFeeId string) (string,string,string,string, error) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("transactionhandlers.go::getSettlementAccount()::error in connecting to database due to ",err)
		return "","","","",err
	}
	defer con.Close()
	var code,bearer,iMerchantName,iFeeName interface{}
	var accountCode,feeBearer,sMerchantName,sFeeName string 
	q := `SELECT "merchant_accounts"."AccountCode","merchant_fees"."FeeBearer","merchants"."Title", get_fee_title("merchant_fees"."FeeId") FROM "merchant_accounts" 
	INNER JOIN "merchant_fees" ON "merchant_fees"."Id" = "merchant_accounts"."MerchantFeeId" INNER JOIN "merchants" ON "merchants"."Id" = "merchant_accounts"."MerchantId" WHERE "merchant_accounts"."MerchantFeeId" = $1 AND "merchant_accounts"."Enabled" = $2` 
	err = con.Db.QueryRow(q, merchantFeeId,true).Scan(&code,&bearer,&iMerchantName,&iFeeName)
	if err != nil {
		fmt.Println("transactionhandlers.go::getSettlementAccount()::error in fetching account code from database due to ",err)
		if s.Contains(fmt.Sprintf("%v", err), "no rows") == true {
			return "","","","",errors.New("Sorry, selected fee is yet to be enabled")
		}
	}
	if code == nil {
		return "","","","",errors.New("account code for merchant/fee not yet generated")
	}
	if bearer == nil {
		return "","","","",errors.New("fee bearer for merchant/fee is empty")
	}
	if iMerchantName != nil {
		sMerchantName = iMerchantName.(string)
	}
	if iFeeName != nil {
		sFeeName = iFeeName.(string)
	}
	accountCode = code.(string)
	feeBearer = bearer.(string)
	return sMerchantName,accountCode,sFeeName,feeBearer, nil 
}

func dbInsertUserTransaction(uId,reference,categoryName,merchantId,feeId,referenceName,referenceId string, amount int) (string, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()

	var insertedTxId string
	insertQuery := `INSERT INTO "payment_transactions"("Id","UserId","TxReference","TxDate","TxAmount","TxPaymentGateway","MerchantId","FeeId","CategoryName","TxPaymentReferenceName","TxPaymentReferenceId") 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),uId,reference,time.Now(),amount,"PayStack",merchantId,feeId,categoryName,referenceName,referenceId).Scan(&insertedTxId)
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