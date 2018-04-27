package handlers

import (
	"fmt"
	h "github.com/kenmobility/feezbot/helper"
	"net/http"
	"io/ioutil"
	"encoding/json"
	s "strings"
	"time"

	"github.com/labstack/echo"
)

type UserLogin struct {
	Details UserCred `json:"userLogin"`
}

type UserCred struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IpAddress string `json:"ipAddress"`	
}

type UserLoginResponse struct {
	h.Response
	UserInfo UserInfo `json:"userInfo"`	
}

type UserInfo struct {
	LastName  string `json:"lastname"`
	OtherName string `json:"othername"`
	UserName  string `json:"username"`
} 

func CheckPassword(c echo.Context) error {
	password := c.QueryParam("password")
	/* passHash,err := h.CreateHash(password)
	if err != nil {
		fmt.Println("error in hashing password due to :", err)
	} */
	//fmt.Println("hash password is ", passHash)
	passHash := "AC+PON1613PrxIBRBc9+6BCePsGgHp5+LEW6criHEjWmNAIntVI0v6PdvfCHCSW6PQ=="
	//passCheck := h.ValidatePassword(password, passHash)
	passCheck := h.BcryptValidatePassword(password, passHash)
	fmt.Println("response is ", passCheck)
	return c.String(http.StatusOK, "")
} 



func Login(c echo.Context) error {
	con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()
	var u UserLogin
	defer c.Request().Body.Close()
	b,err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Printf("authenticationhandlers.go::Login(): failed to read request body due to : %s", err)
		return c.String(http.StatusInternalServerError, "error in reading the request body")
	}
	//fmt.Printf("the raw json request is %s\n", b)
	err = json.Unmarshal(b, &u)
	if err != nil {
		fmt.Printf("authenticationhandlers.go::Login(): failed to unmarshal json request body: %s", err)
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)
	}
	//Check for complete credentials
	if u.Details.Username == "" || u.Details.Password == "" || u.Details.IpAddress == "" {
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)	
	} 

	username := s.ToLower(s.Trim(u.Details.Username," "))
	uIpAddress := u.Details.IpAddress

	//fmt.Printf("json object is : %#v\n", u)
	//check if username exists on the db
	var id,email,password,phone, emailConfirmed, phoneConfirmed,lockOutEnabled, twoFaEnabled, lastName, otherNames interface{}
	var uId,uEmail,uPasswordHash, uPhone, uLastName, uOtherNames string
	var uEmailConf, uPhoneConf, uTwoFaEnabled,uLockoutEnabled bool
	q := `SELECT "AspNetUsers"."Id","AspNetUsers"."Email","AspNetUsers"."EmailConfirmed","AspNetUsers"."PasswordHash",
			"AspNetUsers"."PhoneNumber","AspNetUsers"."PhoneNumberConfirmed","AspNetUsers"."LockoutEnabled",
			"AspNetUsers"."TwoFactorEnabled", "profiles"."LastName", "profiles"."OtherNames" FROM "AspNetUsers" INNER JOIN "profiles" ON
			"AspNetUsers"."Id" = "profiles"."UserId" WHERE "AspNetUsers"."UserName" = $1`
	err = con.Db.QueryRow(q, username).Scan(&id,&email,&emailConfirmed,&password,&phone,&phoneConfirmed,&lockOutEnabled,&twoFaEnabled,&lastName,&otherNames)
	if err != nil {
		//fmt.Println("error is :", err.Error())
		if s.Contains(fmt.Sprintf("%v", err), "no rows") == true {
			//user does not exist
			res := h.Response {
				Status: "error",
				Message:"Invalid Username!",
			}
			return c.JSON(http.StatusNotFound, res)	
		}else{
			fmt.Println("authenticationhandlers.go::Login()::Select From AspNetUsers table Failed due to:", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	if id != nil {
		uId = id.(string)
	}
	if email != nil {
		uEmail = email.(string)
	}
	if emailConfirmed != nil {
		uEmailConf = emailConfirmed.(bool)
	}
	if password != nil {
		uPasswordHash = password.(string)
	}
	if phone != nil {
		uPhone = phone.(string)
	}
	if phoneConfirmed != nil {
		uPhoneConf = phoneConfirmed.(bool)
	}
	if lockOutEnabled != nil {
		uLockoutEnabled = lockOutEnabled.(bool)
	}
	if twoFaEnabled != nil {
		uTwoFaEnabled = twoFaEnabled.(bool)
	}
	if lastName != nil {
		uLastName = lastName.(string)
	}
	if otherNames != nil {
		uOtherNames = otherNames.(string)
	}
	fmt.Printf("SELECTED: id-%s, email-%s, password-%s,phone-%s,emailConfirmed-%v,phoneConfirmed-%v,lockoutEnabled-%v,twoFaEnabled-%v,lastName-%s,otherName-%s\n", 
		uId,uEmail,uPasswordHash,uPhone,uEmailConf,uPhoneConf,uLockoutEnabled,uTwoFaEnabled,uLastName,uOtherNames)

	//validate user's supplied password
	if isPasswordValid := h.BcryptValidatePassword(u.Details.Password, uPasswordHash); !isPasswordValid {
		//increment AccessFailedCount field of the user
		err := incrementAccessFailedCount(username) 
		if err != nil {
			fmt.Println("authenticationhandlers.go::Login()::failed to increment AccessFailedCount field of user due to ", err)
		}
		res := h.Response {
			Status: "error",
			Message:"Incorrect Password!",
		}
		return c.JSON(http.StatusUnauthorized, res)
	}

	//check if user account has been locked out
	if uLockoutEnabled == true {
		//insert into user Audits table with auditEvent as 'AccountLockout'
		userAuditId, err := userAuditsInsert(uId, uIpAddress, "AccountLockout")
		if err != nil {
			fmt.Println("failed to insert into user audits table due to ", err)
		}
		fmt.Println("id inserted into user audits account is ", userAuditId)
		res := h.Response {
			Status: "error",
			Message:"User Account locked out, Contact Admin",
		}
		return c.JSON(http.StatusLocked, res)
	}

	//if password is valid and account isn't locked out, check if phone number has been confirmed
	if uPhoneConf == false {
		res := h.Response {
			Status: "error",
			Message:"User Phone number yet to be confirmed",
		}
		return c.JSON(http.StatusUnauthorized, res)
	}

	//Update or insert login_counts table for user
	_,err = loginCountsUpdateOrInsert(uId)
	if err != nil {
		fmt.Println("failed to update or insert login numberOfTimes field of login_counts table due to ", err)
	}

	//insert into user Audits table with auditEvent as 'Login'
	_, err = userAuditsInsert(uId, uIpAddress, "Login")
	if err != nil {
		fmt.Println("failed to insert into user audits table for successfull login due to ", err)
	}	
	dRes := h.Response {
		Status: "success",
		Message:"Logged In Successfully",
	}
	uDetail := UserInfo {
		LastName: uLastName,
		OtherName: uOtherNames,
		UserName:username,
	}
	res := UserLoginResponse {
		dRes,
		uDetail,
	} 
	return c.JSON(http.StatusOK, res)
}

func CheckHash(nonce, apiKey, apiSecret, signature string) bool {
	concat := nonce + apiKey + apiSecret
	hashedString := h.Hash256(concat)
	fmt.Println("hashed string is - ", hashedString)
	if hashedString != signature {
		return false
	}
	return true
}

func ValidateSignature(apiKey, apiSecret, signature string) (bool, string) {
	/* con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("validate.go. Failed to Open connection :", err)
	}
	var userId string
	err = con.Db.QueryRow(`SELECT "UserId" FROM api_accounts WHERE "ApiKey" = $1 AND "ApiSecret" = $2 AND "Signature" = $3`, apiKey, apiSecret, signature).Scan(&userId)
	fmt.Println("error is ", err) */
	q := `SELECT "UserId" FROM api_accounts WHERE "ApiKey" = $1 AND "ApiSecret" = $2 AND "Signature" = $3` 
	//userId := h.DBSelectNoErr(q, apiKey, apiSecret, signature)
	userId,err := h.DBSelect(q, apiKey, apiSecret, signature)
	if err != nil {
		if err == h.NoRows {
			return false, "Invalid Api credentials"
		}
		fmt.Println("authenticationhandlers.go::ValidateSignature(): failed selecting from api_accounts due to ", err)	
	} 
	/* if userId.(string) == "null" {
		fmt.Println("user id equal to empty")
		return false
	}*/
	if userId != nil {
		//fmt.Println("user id not equal to nil")
		if userId.(string) != "" {
			query := `SELECT "Enabled" FROM api_accounts WHERE "UserId" = $1` 
			status,_ := h.DBSelect(query,userId)
			if status.(bool) == false{
				return false, "Api Account yet to be Enabled"
			}
			//fmt.Println(userId.(string))
			return true, ""
		}		
	}
	/* if userId == nil {
		//fmt.Println("user id equal to nil")
		return false		
	} */
	return false,""
}

func incrementAccessFailedCount(username string) error {
	con, err := h.OpenConnection()
	if err != nil {
		return err
	}
	defer con.Close()
	var userId string
	var incrementedCount, iCount int
	q := `SELECT "AccessFailedCount" FROM "AspNetUsers" WHERE "UserName"= $1`
	initialCount,err := h.DBSelect(q,username)
	if err != nil {
		fmt.Println("authenticationhandlers.go::incrementAccessFailedCount()::main error encountered is ", err)
		return err
	}	
	if initialCount != nil {
		fmt.Println("authenticationhandlers.go::incrementAccessFailedCount()::Selected AccessFailedCount of user as",initialCount.(int64))
		iCount = int(initialCount.(int64))
		incrementedCount =  iCount + 1
	}
	fmt.Printf("current access failed count of %s is [%v]", username, incrementedCount) 

	//update AspNetUsers table with the incremented count
	updateQuery := `UPDATE "AspNetUsers" SET "AccessFailedCount"= $1 WHERE "UserName" = $2 RETURNING "Id"`
	err = con.Db.QueryRow(updateQuery, incrementedCount,username).Scan(&userId)
	if err != nil {
		fmt.Println("authenticationhandlers.go::incrementAccessFailedCount()::update error encountered is ", err)
		return err
	}
	fmt.Println("authenticationhandlers.go::incrementAccessFailedCount()::user id incremented is ", userId)
	return nil
}

func loginCountsUpdateOrInsert(uId string) (string,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "",err
	}
	defer con.Close()
	var userId string
	var count int
	err = con.Db.QueryRow(`SELECT "UserId","NumberOfTimes" FROM "login_counts" WHERE "UserId" = $1`,uId).Scan(&userId, &count)
	if err != nil {
		if err == h.NoRows {
			//if userId doesn't exist in the table, insert
			err = con.Db.QueryRow(`INSERT INTO "login_counts"("UserId","NumberOfTimes","LastLoggedInDate") VALUES($1,$2,$3) RETURNING "UserId"`,uId,1,time.Now()).Scan(&userId)
			if err != nil {
				fmt.Println("authenticationhandlers.go::loginCountsUpdateOrInsert()::error encountered is ", err)
			}
			fmt.Println("authenticationhandlers.go::loginCountsUpdateOrInsert()::inserted User Id is",userId)
		}
		fmt.Println("authenticationhandlers.go::loginCountsUpdateOrInsert()::main error encountered is ", err)
	}	
	if userId != "" && count > 0 {
		//update login_counts table with the incremented count
		fmt.Println("Selected 'UserId' as",userId, "number of times logged in selected is ", count)
		updateQuery := `UPDATE "login_counts" SET "NumberOfTimes"= $1, "LastLoggedInDate"=$2 WHERE "UserId" = $3 RETURNING "UserId"`
		err = con.Db.QueryRow(updateQuery, count + 1,time.Now(),uId).Scan(&userId)
		if err != nil {
			fmt.Println("authenticationhandlers.go::loginCountsUpdateOrInsert()::update error encountered is ", err)
			return "",err
		}
		fmt.Println("authenticationhandlers.go::loginCountsUpdateOrInsert()::user id incremented is ", userId)
	} 
	return userId,nil
}


/* func createJwtToken(name string, emailConfirmed, phoneConfirmed bool) (string,error) {
	claims := JwtClaims {
		Name: name,
		EmailConfirmed: emailConfirmed,
		PhoneConfirmed: phoneConfirmed,
		DefaultClaims: jwt.StandardClaims {
			Id: "main_user_id",
			ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
		},
	}
	rawToken := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
} */

