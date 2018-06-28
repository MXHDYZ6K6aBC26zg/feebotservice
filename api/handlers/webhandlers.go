package handlers

import(
	"github.com/kenmobility/feezbot/gateways/paystack"
	g "github.com/kenmobility/feezbot/gateways"
	e "github.com/kenmobility/feezbot/email"
	"fmt"
	"net/http"
	"github.com/labstack/echo"
	"io/ioutil"
	h "github.com/kenmobility/feezbot/helper"
	"encoding/json"
	"errors"
	s "strings"
)

type SubAccount struct {
	CreateSubAccount struct {
		MerchantID		 		string `json:"merchant_id"`
		MerchantFeeID	 		string `json:"merchant_fee_id"`
		SettlementByMerchant	bool	`json:"settlement_by_merchant"`
		AccountNumber   	 	string `json:"account_number"`
		BankName         		string `json:"bank_name"`
		BusinessName     		string `json:"merchant_name"`
		ContactName     		string `json:"contact_name"`
		ContactEmail     		string `json:"contact_email"`
		ContactPhone     		string `json:"contact_phone"`
		PercentageCharge 		float64 `json:"percentage_charge"`
	} `json:"createSubAccount"`
}

func CreateSubAccount(c echo.Context) error {
	var csa SubAccount
	defer c.Request().Body.Close()
	b,err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Printf("webhandlers.go::CreateSubAccount()::failed to read request body due to : %s\n", err)
		r := h.Response {
			Status: "error",
			Message: err.Error(), //"error occured, please try again",//err.Error(),
		}
		return c.JSON(http.StatusBadRequest, r)
	}
	//fmt.Printf("the raw json request is %s\n", b)
	err = json.Unmarshal(b, &csa)
	if err != nil {
		fmt.Printf("webhandlers.go::CreateSubAccount()::failed to unmarshal json request body: %s\n", err)
		r := h.Response {
			Status: "error",
			Message: err.Error(),//"error occured, please try again",
		}
		return c.JSON(http.StatusInternalServerError, r)
	}

	businessName := csa.CreateSubAccount.BusinessName
	bankName := csa.CreateSubAccount.BankName
	accountNumber := csa.CreateSubAccount.AccountNumber
	percCharge := csa.CreateSubAccount.PercentageCharge
	merchantId := csa.CreateSubAccount.MerchantID
	merchantFeeId := csa.CreateSubAccount.MerchantFeeID
	settlementByMerchant := csa.CreateSubAccount.SettlementByMerchant
	contactEmail := csa.CreateSubAccount.ContactEmail
	contactPhone := csa.CreateSubAccount.ContactPhone
	contactName := csa.CreateSubAccount.ContactName

	if businessName == "" || bankName == "" || accountNumber == "" || merchantId == "" || merchantFeeId == "" || percCharge <= 0.00 {
		r := h.Response {
			Status: "error",
			Message:"Invalid request format Or required parameters not complete",
		}
		return c.JSON(http.StatusBadRequest, r)	
	}
	status,err := isSubAccountExist(merchantId, merchantFeeId)
	fmt.Printf("isSubAccountExist func returns error as '%s' and status as '%v' for merchant_id - '%s' and merchant_fee_id - '%s'\n",err,status,merchantId, merchantFeeId)
	if err != nil && status == false {
		r := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, r)
	}
	if err != nil && status == true {
		r := h.Response {
			Status: "error",
			Message: err.Error(),
		}
		return c.JSON(http.StatusBadRequest, r)
	}
	resp := paystack.CreateSubAccount(businessName, bankName, accountNumber, contactEmail, contactName, contactPhone, percCharge)
	
	if settlementByMerchant == false {
		d := g.SubAccountResponse {
			StatusCode: resp.StatusCode,
			Status: resp.Status,
			ResponseMsg: resp.ResponseMsg,
			AccountCode: resp.AccountCode,
			PercentageCharge: resp.PercentageCharge,
			SettlementBank: resp.SettlementBank,
			AccountNumber: resp.AccountNumber,
			MerchantName: resp.MerchantName,
			MerchantId:   merchantId,
			MerchantFeeId: merchantFeeId,
			SettlementByMerchant: settlementByMerchant,
		}
		bs,_:= json.Marshal(d)
		r := h.Response {
			Status: resp.Status,
			Message:resp.ResponseMsg,
			Data: bs,
		}
		return c.JSON(resp.StatusCode, r)
	}
	d := g.SubAccountResponse {
		StatusCode: resp.StatusCode,
		Status: resp.Status,
		ResponseMsg: resp.ResponseMsg,
		AccountCode: resp.AccountCode,
		PercentageCharge: resp.PercentageCharge,
		SettlementBank: resp.SettlementBank,
		AccountNumber: resp.AccountNumber,
		MerchantName: resp.MerchantName,
		MerchantId:   merchantId,
		SettlementByMerchant: settlementByMerchant,
	}
	bs,_:= json.Marshal(d)
	r := h.Response {
		Status: resp.Status,
		Message:resp.ResponseMsg,
		Data: bs,
	}

	return c.JSON(resp.StatusCode, r)
}

func SendMailFromWeb(c echo.Context) error {
	recipient := s.ToLower(c.FormValue("recipient"))
	subject := c.FormValue("subject")
	body := c.FormValue("body")

	mailObj := e.MailConfig("feeracksolution@gmail.com", "Password1@", recipient, subject, body)
	err := e.SendMail(mailObj)
	if err != nil {
		fmt.Println("webhandlers.go::SendMailFromWeb():: error encountered while sending mail is ", err)
		r := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, r)
	}
	r := h.Response {
		Status: "success",
		Message: "Email Sent Successfully",
	}
	return c.JSON(http.StatusOK, r)
}

func isSubAccountExist(merchantId,merchantFeeId string) (bool,error) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("webhandlers.go::isSubAccountExist()::error in connecting to database due to ",err)
		return false,err
	}
	defer con.Close()
	var code interface{}
	var enabled bool

	q := `SELECT "merchant_accounts"."AccountCode","merchant_accounts"."Enabled" FROM "merchant_accounts" 
		  WHERE "merchant_accounts"."MerchantId" = $1 AND "merchant_accounts"."MerchantFeeId" = $2`
	err = con.Db.QueryRow(q, merchantId,merchantFeeId).Scan(&code,&enabled)
	if err != nil {
		fmt.Println("webhandlers.go::isSubAccountExist()::error in fetching account code from database due to ",err)
		return false,err
	}
	if code != nil && enabled == true {
		return true, errors.New("this account is already active and enabled")
	}
	if code != nil && enabled == false {
		return false, nil
	}
	if code == nil && enabled == false {
		return false,nil
	}
	return false, nil
}