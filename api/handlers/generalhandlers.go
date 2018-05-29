package handlers

import (
	"github.com/labstack/echo"
	"github.com/kenmobility/feezbot/gateways/paystack"
	h "github.com/kenmobility/feezbot/helper"
	"net/http"
	"fmt"
	"log"
	"errors"

)

const txSuccessResp = `
<!DOCTYPE html>
<html >
<head>
  <meta charset="UTF-8">
  <title>FeeRack::About Transaction</title>
      <style type="text/css">
	  body {
  background: #e9e9e9;
  color: #666666;
  font-family: 'RobotoDraft', 'Roboto', sans-serif;
  font-size: 14px;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Pen Title */
.pen-title {
  padding: 50px 0;
  text-align: center;
  letter-spacing: 2px;
}
.pen-title h1 {
  margin: 0 0 20px;
  font-size: 48px;
  font-weight: 300;
}
.pen-title span {
  font-size: 12px;
}
.pen-title span .fa {
  color: #33b5e5;
}
.pen-title span a {
  color: #33b5e5;
  font-weight: 600;
  text-decoration: none;
}

/* Form Module */
.form-module {
  position: relative;
  background: #ffffff;
  max-width: 320px;
  width: 100%;
  border-top: 5px solid #33b5e5;
  box-shadow: 0 0 3px rgba(0, 0, 0, 0.1);
  margin: 0 auto;
}
.form-module .toggle {
  cursor: pointer;
  position: absolute;
  top: -0;
  right: -0;
  background: #33b5e5;
  width: 30px;
  height: 30px;
  margin: -5px 0 0;
  color: #ffffff;
  font-size: 12px;
  line-height: 30px;
  text-align: center;
}
.form-module .toggle .tooltip {
  position: absolute;
  top: 5px;
  right: -65px;
  display: block;
  background: rgba(0, 0, 0, 0.6);
  width: auto;
  padding: 5px;
  font-size: 10px;
  line-height: 1;
  text-transform: uppercase;
}
.form-module .toggle .tooltip:before {
  content: '';
  position: absolute;
  top: 5px;
  left: -5px;
  display: block;
  border-top: 5px solid transparent;
  border-bottom: 5px solid transparent;
  border-right: 5px solid rgba(0, 0, 0, 0.6);
}
.form-module .form {
  display: none;
  padding: 40px;
}
.form-module .form:nth-child(2) {
  display: block;
}
.form-module h2 {
  margin: 0 0 20px;
  color: #33b5e5;
  font-size: 18px;
  font-weight: 400;
  line-height: 1;
}
.form-module input {
  outline: none;
  display: block;
  width: 100%;
  border: 1px solid #d9d9d9;
  margin: 0 0 20px;
  padding: 10px 15px;
  box-sizing: border-box;
  font-wieght: 400;
  -webkit-transition: 0.3s ease;
  transition: 0.3s ease;
}
.form-module input:focus {
  border: 1px solid #33b5e5;
  color: #333333;
}
.form-module button {
  cursor: pointer;
  background: #33b5e5;
  width: 100%;
  border: 0;
  padding: 10px 15px;
  color: #ffffff;
  -webkit-transition: 0.3s ease;
  transition: 0.3s ease;
}
.form-module button:hover {
  background: #178ab4;
}
.form-module .cta {
  background: #f2f2f2;
  width: 100%;
  padding: 15px 40px;
  box-sizing: border-box;
  color: #666666;
  font-size: 12px;
  text-align: center;
}
.form-module .cta a {
  color: #333333;
  text-decoration: none;
}
</style>
</head>

<body>
  
<!-- Form Mixin-->
<!-- Input Mixin-->
<!-- Button Mixin-->
<!-- Pen Title-->
<div class="module form-module">

</div>
<div class="pen-title">
<h1>Your Payment transaction was Successful.</h1><br>
<h1>You can now close your browser to return to the App</h1>
</div>
<!-- Form Module-->
<div class="module form-module">

</div>
</body>
</html>
`

const txFailedResp = `
<!DOCTYPE html>
<html >
<head>
  <meta charset="UTF-8">
  <title>FeeRack::About Transaction</title>
      <style type="text/css">
	  body {
  background: #e9e9e9;
  color: #666666;
  font-family: 'RobotoDraft', 'Roboto', sans-serif;
  font-size: 14px;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Pen Title */
.pen-title {
  padding: 50px 0;
  text-align: center;
  letter-spacing: 2px;
}
.pen-title h1 {
  margin: 0 0 20px;
  font-size: 48px;
  font-weight: 300;
}
.pen-title span {
  font-size: 12px;
}
.pen-title span .fa {
  color: #33b5e5;
}
.pen-title span a {
  color: #33b5e5;
  font-weight: 600;
  text-decoration: none;
}

/* Form Module */
.form-module {
  position: relative;
  background: #ffffff;
  max-width: 320px;
  width: 100%;
  border-top: 5px solid #33b5e5;
  box-shadow: 0 0 3px rgba(0, 0, 0, 0.1);
  margin: 0 auto;
}
.form-module .toggle {
  cursor: pointer;
  position: absolute;
  top: -0;
  right: -0;
  background: #33b5e5;
  width: 30px;
  height: 30px;
  margin: -5px 0 0;
  color: #ffffff;
  font-size: 12px;
  line-height: 30px;
  text-align: center;
}
.form-module .toggle .tooltip {
  position: absolute;
  top: 5px;
  right: -65px;
  display: block;
  background: rgba(0, 0, 0, 0.6);
  width: auto;
  padding: 5px;
  font-size: 10px;
  line-height: 1;
  text-transform: uppercase;
}
.form-module .toggle .tooltip:before {
  content: '';
  position: absolute;
  top: 5px;
  left: -5px;
  display: block;
  border-top: 5px solid transparent;
  border-bottom: 5px solid transparent;
  border-right: 5px solid rgba(0, 0, 0, 0.6);
}
.form-module .form {
  display: none;
  padding: 40px;
}
.form-module .form:nth-child(2) {
  display: block;
}
.form-module h2 {
  margin: 0 0 20px;
  color: #33b5e5;
  font-size: 18px;
  font-weight: 400;
  line-height: 1;
}
.form-module input {
  outline: none;
  display: block;
  width: 100%;
  border: 1px solid #d9d9d9;
  margin: 0 0 20px;
  padding: 10px 15px;
  box-sizing: border-box;
  font-wieght: 400;
  -webkit-transition: 0.3s ease;
  transition: 0.3s ease;
}
.form-module input:focus {
  border: 1px solid #33b5e5;
  color: #333333;
}
.form-module button {
  cursor: pointer;
  background: #33b5e5;
  width: 100%;
  border: 0;
  padding: 10px 15px;
  color: #ffffff;
  -webkit-transition: 0.3s ease;
  transition: 0.3s ease;
}
.form-module button:hover {
  background: #178ab4;
}
.form-module .cta {
  background: #f2f2f2;
  width: 100%;
  padding: 15px 40px;
  box-sizing: border-box;
  color: #666666;
  font-size: 12px;
  text-align: center;
}
.form-module .cta a {
  color: #333333;
  text-decoration: none;
}

	  </style>

  
</head>

<body>
  
<!-- Form Mixin-->
<!-- Input Mixin-->
<!-- Button Mixin-->
<!-- Pen Title-->
<div class="module form-module">

</div>
<div class="pen-title">
<h1>Your transaction failed</h1><br>
</div>
<!-- Form Module-->
<div class="module form-module">

</div>
</body>
</html>
`

func HandleCallbackResponse(c echo.Context) error{
	reference := c.QueryParam("reference")
	if reference == "" {
		log.Println("no reference found")
		return c.JSON(http.StatusInternalServerError, "no reference found")
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
		return c.HTML(http.StatusOK,txFailedResp)
	}
	return c.HTML(http.StatusOK, txSuccessResp)
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
		fmt.Println("transactionhandlers.go::dbinsertSuccessChargeCardResponse()::error encountered while inserting into transactions for success card response is ", err)
		return "",err
	}
	//check if the row was inserted successfully
	if insertedTxId == "" {
		return "", errors.New("inserting into transactions failed")
	} 
	return insertedTxId, nil
} 


func Test(c echo.Context) error {
	/* con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()
	var email, userId, phone, username string
	var emailConf, phoneConf bool
	q := `SELECT (get_users_by_rolename('User'))`
	rows,err := con.Db.Query(q)
	defer rows.Close()
	fmt.Println("rows are ", rows)
	if err != nil{
		if strings.Contains(fmt.Sprintf("%v", err), "no records") == true {
			res := h.Response {
				Status: "error",
				Message:"No record found!",
			}
			return c.JSON(http.StatusOK, res)	
		}else{
			fmt.Println("generalhandlers.go::Test()::test sql query Failed due to:", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	for rows.Next() {
		err = rows.Scan(&userId,&email,&emailConf,&phone,&phoneConf,&username)
		if err != nil {
			fmt.Println("generalhandlers.go::Test()::test sql scan Failed due to:", err)
		}
		fmt.Println("user_id - ",userId, "email - ", email, "email conf - ", emailConf, "phone - ", phone, "phone conf - ", phoneConf, "username - ", username)
	} */
	return c.String(http.StatusOK, "Hello, World!")
}

func SeedTable(c echo.Context) error {
	roleName := c.QueryParam("name")
	
	con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()

	q := `UPDATE "AspNetRoles" SET "Id" = $1 WHERE "Name" = $2`
	re,err := con.Db.Exec(q, h.GenerateUuid(),roleName)
	if err != nil {
		fmt.Println("seeding into AspNetRoles failed due to ", err)
	}
	affRows,_ := re.RowsAffected()
	return c.String(http.StatusOK, fmt.Sprintf("affected %v Row(s)",affRows))
} 



