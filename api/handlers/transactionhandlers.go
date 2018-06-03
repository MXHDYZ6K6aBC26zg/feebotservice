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
)

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
	reference := s.Trim(c.FormValue("reference")," ")

	if userId == "" || merchantId == "" || merchantFeeId == "" || feeId == "" || amount == "" || categoryName == "" || paymentReferenceName == "" || paymentReferenceId == "" || reference == ""{
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
	_,err = dbInsertUserTransaction(userId, reference,categoryName,merchantId,feeId,paymentReferenceName,paymentReferenceId,intAmount)
	if err != nil {
		fmt.Println("error while inserting user transaction detail : ", err)
	}
	//check if amount is less than 2500 Naira
	if intAmount < 2500 {
		fee := (percCharge / 100) * (floatAmount - 100)
		flatFee = fee + 100
		feeBearer = "account"
	}else{
		//if amount is greater than or equal to 2500 Naira
		flatFee = amountByPercentageCharge(floatAmount, percCharge)
		feeBearer = "subaccount"
	}
	pDetail := map[string]interface{} {
		"email": email,
		"subaccount_code": subaccount,
		"fee_bearer": feeBearer,
		"transaction_reference": reference,
		"transaction_charge": int(flatFee),
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
	reference := c.QueryParam("reference")
	  if reference == "" {
		  log.Println("no reference found")
		  r := h.Response {
		Status: "error",
		Message:"no reference found",
	  }
	  return c.JSON(http.StatusNotFound, r)
	  }
	  log.Println("reference is ", reference)
  
	  resp := paystack.VerifyTransaction(reference)
	  if resp.StatusCode != 200 {
		  fmt.Printf("transaction with reference %s failed due to %s\n", reference, resp.ResponseMsg)
	  }
  
	  var updatedStatus bool
	  q := `SELECT "IsUpdated" FROM "payment_transactions" WHERE "TxReference"= $1`
	  uStatus,_ := h.DBSelect(q,reference)
	  if uStatus != nil {
		  updatedStatus = uStatus.(bool)
	  }
	  
	  if updatedStatus == false {
		  _,err := dbUpdateChargeResponse(resp.Reference,resp.Email,resp.TxCreatedAt,resp.PaidAt,resp.ResponseStatus,resp.TxCurrency,resp.TxChannel,resp.AuthorizationCode,resp.CardLast4,resp.ResponseBody,
				  resp.Bank,resp.CardType,resp.GatewayResponse,resp.TxFeeBearer,resp.PercentageCharged,resp.SubAccountSettlementAmount,resp.MainAccountSettlementAmount,resp.StatusCode,resp.TxAmount,resp.TxFees)
		  if err != nil {
			  fmt.Println("error encountered while updating payment_transactions table is ", err)
		  }
	  }
	  if resp.ResponseStatus != "success" {
		  r := h.Response {
		Status: "success",
		Message: fmt.Sprintf("Payment transaction with reference - %s failed due to %s",reference,resp.GatewayResponse),
	  }
	  return c.JSON(http.StatusOK, r)
	}
	r := h.Response {
	  Status: "success",
	  Message: fmt.Sprintf("Payment transaction with reference - %s was successful",reference),
	}
	return c.JSON(http.StatusOK, r)
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

func dbUpdateChargeResponse(txReference,txEmail,txDate,paidAt,txStatus,txCurrency,txChannel,txAuthCode,cardLast4,responseBody, bank,cardType,gatewayResponse,feeBearer,percentageCharged string,subAccountSettlementAmount,mainAccountSettlementAmount, 
	responseCode,txAmount int, txFee float64) (string,error) {	
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()

	var insertedTxId string
	insertQuery := `UPDATE "payment_transactions" SET "TxProvidedEmail" = $1, "TxCreatedAt" = $2, "TxStatus" = $3, "AmountPaid" = $4, "ResponseBody" = $5, "ResponseCode" = $6,"TxCurrency" = $7, "TxChannel" = $8,"TxAuthorizationCode" = $9 ,
	"CardLast4" = $10, "GatewayResponse"= $11, "TxFees" = $12,"Bank" = $13,"CardType" = $14,"PaidAt" = $15,"TxFeeBearer" = $16, "PercentageCharged" = $17, "SubAccountSettlementAmount" = $18, "MainAccountSettlementAmount" = $19, "IsUpdated" = $20 WHERE "TxReference" = $21  RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,txEmail,txDate,txStatus,txAmount / 100,responseBody,responseCode,txCurrency,txChannel,txAuthCode,cardLast4,gatewayResponse,txFee / 100,bank,cardType,paidAt,feeBearer,percentageCharged,subAccountSettlementAmount / 100,mainAccountSettlementAmount / 100,true,txReference).Scan(&insertedTxId)
	if err != nil {
		fmt.Println("transactionhandlers.go::dbUpdateChargeResponse()::error encountered while inserting into transactions for success card response is ", err)
		return "",err
	}
	//check if the row was inserted successfully
	if insertedTxId == "" {
		return "", errors.New("inserting into transactions failed")
	} 
	return insertedTxId, nil
} 

func amountByPercentageCharge(amount,percCharge float64) float64 {
	mainCharge := percCharge - 1.5
	chargeAmount := (mainCharge / 100) * (amount - 100)
	return chargeAmount * 100
}