package handlers

import (
	"github.com/kenmobility/feezbot/gateways/paystack"
	//g "github.com/kenmobility/feezbot/gateways"
	"fmt"
	"net/http"
	h "github.com/kenmobility/feezbot/helper"
	"io/ioutil"
	"encoding/json"
	//s "strings"
	//"errors"
	//"time"

	"github.com/labstack/echo"
)

type CardPay struct {
	CardPayDetails CardPayment `json:"cardPayment"`
}

type CardPayment struct {
	CardDetails Card  	`json:"card"`
	Username string 	`json:"username"`
	MerchantId string 	`json:"merchant_id"`
	FeeId string 		`json:"fee_id"`
	Amount int			`json:"amount"`
}

type Card struct {
	Cvv         string 	`json:"cvv"`
	ExpiryMonth string 	`json:"expiry_month"`
	ExpiryYear  string 	`json:"expiry_year"`
	Number      string 	`json:"number"`
}

type BankPay struct {
	BankPayDetails BankPayment `json:"bankPayment"`
}

type BankPayment struct {
	BankDetails Bank  	`json:"bank"`
	Username string 	`json:"username"`
	MerchantId string 	`json:"merchant_id"`
	FeeId string 		`json:"fee_id"`
	Amount int			`json:"amount"`
	Birthday string     `json:"birthday"`
}

type Bank struct {
	Code          string `json:"code"`
	AccountNumber string `json:"account_number"`
}

type UserPin struct {
	SubmitPin struct {
		Username  string `json:"username"`
		Pin       string `json:"pin"`
		Reference string `json:"reference"`
	} `json:"submitPin"`
}

type UserOtp struct {
	SubmitOtp struct {
		Username  string `json:"username"`
		Otp       string `json:"otp"`
		Reference string `json:"reference"`
	} `json:"submitOtp"`
}

//ChargeUserByCard is a POST request handler used to charge user using their ATM cards
func ChargeUserByCard(c echo.Context) error {
	var cp CardPay
	defer c.Request().Body.Close()
	b,err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Printf("transactionhandlers.go::ChargeUserByCard()::failed to read request body due to : %s\n", err)
		return c.String(http.StatusInternalServerError, "error in reading the request body")
	}
	//fmt.Printf("the raw json request is %s\n", b)
	err = json.Unmarshal(b, &cp)
	if err != nil {
		fmt.Printf("transactionhandlers.go::ChargeUserByCard()::failed to unmarshal json request body: %s\n", err)
		r := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusBadRequest, r)
	}
	//fmt.Printf("json object is : %+v\n", cp)

	username := cp.CardPayDetails.Username
	cardNumber := cp.CardPayDetails.CardDetails.Number
	cardCvv := cp.CardPayDetails.CardDetails.Cvv
	cardExpiryMonth := cp.CardPayDetails.CardDetails.ExpiryMonth
	cardExpiryYear := cp.CardPayDetails.CardDetails.ExpiryYear
	amount := cp.CardPayDetails.Amount
	merchantId := cp.CardPayDetails.MerchantId
	feeId := cp.CardPayDetails.FeeId	

	//if cp.CardPayDetails.Username == "" || cp.CardPayDetails.CardDetails.Number == "" || cp.CardPayDetails.CardDetails.Cvv == "" || cp.CardPayDetails.CardDetails.ExpiryMonth == "" || cp.CardPayDetails.CardDetails.ExpiryYear == "" || cp.CardPayDetails.Amount <= 0 {
	if username == "" || cardNumber == "" || cardCvv == "" || cardExpiryMonth == "" || cardExpiryYear == "" || merchantId == "" || feeId == "" || amount <= 0 {
		r := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, r)	
	} 
	email,_,_ := isEmailAndPhoneConfirmed(username)
/* 	if emailConfStatus == false {
		r := h.Response {
			Status: "error",
			Message:"Your Email address has not yet been confirmed, click 'Confirm My Email' to confirm ur Address before proceeding to make payments",
		}
		return c.JSON(http.StatusForbidden, r)	
	} */
	//fmt.Println("uid = ",uId, "userEmail = ", email, "merchantId = ", merchantId, "feeId = ",feeId)
	res := paystack.ChargeByCardDetails(email, cardNumber, cardCvv, cardExpiryMonth, cardExpiryYear,amount)
	fmt.Printf("%+v\n", res)

	//call a function that will insert the details returned into the db depending on the transaction status	
	//statusResponse := checkResponseStatus(res,uId,email,merchantId,feeId,amount/100, "card")

	return c.JSON(res.StatusCode, "hurray")
}

//ChargeUserByBank is a POST request handler used to charge user by supported banks
/* func ChargeUserByBank(c echo.Context) error {
	var bp BankPay
	defer c.Request().Body.Close()
	b,err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Printf("transactionhandlers.go::ChargeUserByBank()::failed to read request body due to : %s\n", err)
		return c.String(http.StatusInternalServerError, "error in reading the request body")
	}
	//fmt.Printf("the raw json request is %s\n", b)
	err = json.Unmarshal(b, &bp)
	if err != nil {
		fmt.Printf("transactionhandlers.go::ChargeUserByBank()::failed to unmarshal json request body: %s\n", err)
		r := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusBadRequest, r)
	}
	username := bp.BankPayDetails.Username	
	amount := bp.BankPayDetails.Amount
	merchantId := bp.BankPayDetails.MerchantId
	feeId := bp.BankPayDetails.FeeId	
	birthday := bp.BankPayDetails.Birthday	
	bankCode := bp.BankPayDetails.BankDetails.Code
	accountNumber := bp.BankPayDetails.BankDetails.AccountNumber

	if username == "" || bankCode == "" || accountNumber == "" || merchantId == "" || feeId == "" || birthday == "" || amount <= 0 {
		r := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
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

	res := paystack.ChargeByBankDetails(email, bankCode, accountNumber, birthday, "", "", amount)

	//fmt.Printf("%+v\n", res)

	statusResponse := checkResponseStatus(res,uId,email,merchantId,feeId,amount / 100, "bank")

	return c.JSON(res.StatusCode, statusResponse)
} 

//ProccessPin is a POST request handler used to process the submitted user's PIN during transaction
func ProccessPin(c echo.Context) error {
	var up UserPin
	defer c.Request().Body.Close()
	b,err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Printf("transactionhandlers.go::ProccessPin()::failed to read request body due to : %s\n", err)
		return c.String(http.StatusInternalServerError, "error in reading the request body")
	}
	//fmt.Printf("the raw json request is %s\n", b)
	err = json.Unmarshal(b, &up)
	if err != nil {
		fmt.Printf("transactionhandlers.go::ProccessPin()::failed to unmarshal json request body: %s\n", err)
		r := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusBadRequest, r)
	}
	pin := up.SubmitPin.Pin
	reference := up.SubmitPin.Reference
	username := up.SubmitPin.Username

	if pin == "" || reference == "" || username == "" {
		r := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, r)
	}
	email,_,_ := isEmailAndPhoneConfirmed(username)
 	if emailConfStatus == false {
		r := h.Response {
			Status: "error",
			Message:"Your Email address has not yet been confirmed, click 'Confirm My Email' to confirm ur Address before proceeding to make payments",
		}
		return c.JSON(http.StatusForbidden, r)	
	} 

	res := paystack.ProcessPin(pin, reference)

	statusResponse := checkResponseStatus(res,uId,email,"","",0,"")

	return c.JSON(res.StatusCode, statusResponse)
}

//ProccessOtp is a POST request handler used to process the submitted user's OTP during transaction
func ProccessOtp(c echo.Context) error {
	var uo UserOtp
	defer c.Request().Body.Close()
	b,err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Printf("transactionhandlers.go::ProccessOtp()::failed to read request body due to : %s\n", err)
		return c.String(http.StatusInternalServerError, "error in reading the request body")
	}
	//fmt.Printf("the raw json request is %s\n", b)
	err = json.Unmarshal(b, &uo)
	if err != nil {
		fmt.Printf("transactionhandlers.go::ProccessOtp()::failed to unmarshal json request body: %s\n", err)
		r := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusBadRequest, r)
	}
	otp := uo.SubmitOtp.Otp
	reference := uo.SubmitOtp.Reference
	username := uo.SubmitOtp.Username

	if otp == "" || reference == "" || username == "" {
		r := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
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

	res := paystack.ProcessOtp(otp, reference)

	statusResponse := checkResponseStatus(res,uId,email,"","",0,"")

	return c.JSON(res.StatusCode, statusResponse)
}
*/