package handlers

import (
	"github.com/labstack/echo"
	"github.com/kenmobility/feezbot/gateways/paystack"
	h "github.com/kenmobility/feezbot/helper"
	"net/http"
	"fmt"
	//"strings"
)

func HandleCallbackResponse(c echo.Context) error{
	reference := c.QueryParam("reference")
	if reference == "" {
		fmt.Println("no reference found")
		return c.JSON(http.StatusInternalServerError, "no reference found")
	}
	resp := paystack.VerifyTransaction(reference)
	if resp.StatusCode != 200 {
		fmt.Printf("transaction with reference %s failed due to %s\n", reference, resp.ResponseMsg)
	}
	insertedId,err := dbUpdateChargeResponse(resp.Reference,resp.Email,resp.TxDate,resp.ResponseStatus,resp.TxCurrency,resp.TxChannel,resp.AuthorizationCode,resp.CardLast4,resp.ResponseBody,
	resp.Bank,resp.CardType,resp.GatewayResponse,resp.StatusCode,resp.TxAmount,resp.TxFees)
	if err != nil {
		fmt.Println("error encountered is ", err)
	}
	fmt.Println("inserted in is ", insertedId)
	return c.String(http.StatusOK, insertedId)

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


