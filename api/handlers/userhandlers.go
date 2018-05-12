package handlers

import (
	"net/http"
	h "github.com/kenmobility/feezbot/helper"
	"github.com/kenmobility/feezbot/rand"
	e "github.com/kenmobility/feezbot/email"
	"fmt"
	"github.com/labstack/echo"
	s "strings"
	"unicode"
	"errors"
	"time"
)

func verifyPassword(s string) (sevenOrMore, number, upper, special bool) {
	letters := 0
	for _, s := range s {
		switch {
		case unicode.IsNumber(s):
			number = true
		case unicode.IsUpper(s):
			upper = true
			letters++
		case unicode.IsPunct(s) || unicode.IsSymbol(s):
			special = true
		case unicode.IsLetter(s) || s == ' ':
			letters++
		default:
			//return false, false, false, false
		}
	}
	sevenOrMore = letters >= 7
	return
}

func ValidateUserExistence(c echo.Context) error {
	username := s.ToLower(s.Trim(c.FormValue("username")," "))
	email := s.ToLower(s.Trim(c.FormValue("email")," "))
	if username == "" || email == "" {
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)
	}
	exists,uId := validateUserByUsernameAndEmail(username, email)
	if  !exists {
		res := h.Response {
			Status: "error",
			Message:"User Account does not exist",
		}
		return c.JSON(http.StatusBadRequest, res)
	}
	//TODO: generate a 6 digit confirmation code and email to user
	err := sendConfirmationCode(uId, email)
	if err != nil {
		res := h.Response {
			Status: "success",
			Message:err.Error(),
		}
		return c.JSON(http.StatusOK, res)
	}
	res := h.Response {
		Status: "success",
		Message:"Provide the confirmation code sent to your email address to reset your password",
	}
	return c.JSON(http.StatusOK, res)
}

func ResetPassword(c echo.Context) error {
	username := s.ToLower(s.Trim(c.FormValue("username")," "))
	email := s.ToLower(s.Trim(c.FormValue("email")," "))
	confirmationCode := s.ToLower(s.Trim(c.FormValue("code")," "))
	password := c.FormValue("password")

	if username == "" || email == "" || confirmationCode == "" || password == "" {
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)	
	}
	return c.String(http.StatusOK, "OK")
}

func CreateUser(c echo.Context) error {
	con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()

	username := s.ToLower(s.Trim(c.FormValue("username")," "))	
	password := c.FormValue("password")
	lastname := s.ToLower(s.Trim(c.FormValue("lastName")," "))
	othername := s.ToLower(s.Trim(c.FormValue("otherName")," "))
	deviceName := s.ToLower(s.Trim(c.FormValue("deviceName")," "))
	deviceModel := s.Trim(c.FormValue("deviceModel")," ")
	deviceUuid := s.Trim(c.FormValue("deviceUuid")," ")
	deviceImei := s.Trim(c.FormValue("deviceImei")," ")
	ipAddress := s.ToLower(s.Trim(c.FormValue("ipAddress")," "))
	email := s.ToLower(s.Trim(c.FormValue("email")," "))
	phone := s.ToLower(s.Trim(c.FormValue("phone")," "))
	phoneVerificationStatus := c.FormValue("phoneVerificationStatus")


	//Check for complete credentials
	if username == "" || password == "" || ipAddress == "" || email == "" || phone == "" || deviceImei == "" || deviceUuid == "" || deviceModel == ""{
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)	
	}

	if emailStatus := isEmailExists(email); emailStatus {
		res := h.Response {
			Status: "error",
			Message:"Email address already exist, register with another email address",
		}
		return c.JSON(http.StatusForbidden, res)
	}

	numComplete,numPresent,upperPresent,specialChar := verifyPassword(password)
	if !numComplete {
		res := h.Response {
			Status: "error",
			Message:"Password must be upto 7 characters",
		}
		return c.JSON(http.StatusForbidden, res)
	}
	if !numPresent {
		res := h.Response {
			Status: "error",
			Message:"Password must contain a number",
		}
		return c.JSON(http.StatusForbidden, res)
	}
	if !upperPresent {
		res := h.Response {
			Status: "error",
			Message:"Password must contain an upper case",
		}
		return c.JSON(http.StatusForbidden, res)
	}
	if !specialChar {
		res := h.Response {
			Status: "error",
			Message:"Password must contain a special character",
		}
		return c.JSON(http.StatusForbidden, res)
	}

	passHash,err := h.BcryptHashPassword(password)
	if err != nil {
		fmt.Println("error in hashing password due to :", err)
	}
	fmt.Printf("Body: username-%s, email-%s, lastname-%s,phone-%s,othername-%s,deviceName-%s,deviceModel-%s,deviceUuid-%s,deviceImei-%s,ipAddress-%s\n", 
		username,email,lastname,phone,othername,deviceName,deviceModel,deviceUuid,deviceImei,ipAddress)

	//Execute AspNetUsers insert
	userId, err := aspNetUsersInsert(username, email, passHash, phone, phoneVerificationStatus)
	if err != nil {
		res := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	if userId == "" {
		//fmt.Println("successfully inserted on AspNetUsers table with user id as: ", userId)
		res := h.Response {
			Status: "error",
			Message:"User creation unsuccessful, pls try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	//fmt.Println("AspNetUsers id is ", userId)

	//fetch User RoleId
	var uRoleId string
	roleQ := `SELECT "Id" FROM "AspNetRoles" WHERE "Name"= $1`
	roleId,err := h.DBSelect(roleQ,"User")
	if err != nil {
		if err == h.NoRows {
			//if Role doesn't exist, create one
			err = con.Db.QueryRow(`INSERT INTO "AspNetRoles"("Id","Name") VALUES($1,$2) RETURNING "Id"`, h.GenerateUuid(),"User").Scan(&uRoleId)
			if err != nil {
				fmt.Println("error encountered is ", err)
			}
			fmt.Println("inserted 'User' role as",uRoleId)
		}
		fmt.Println("main error encountered is ", err)
	}
	
	if roleId != nil {
		fmt.Println("Selected 'User' role as",roleId.(string))
		uRoleId = roleId.(string)
	} 
	fmt.Println("user role Id is ", uRoleId)

	//Execute AspNetUserRoles table Insert
	AspNetUserRolesInsertedUserId,err := aspNetUserRolesInsert(userId,uRoleId)
	if err != nil {
		//failed to insert into AspNetUserRoles table
		fmt.Println("failed to insert into AspNetUserRoles table due to ", err)
		//delete from AspNetUsers
		affRows, _ := aspNetUsersDelete(userId)
		/* if err != nil {
			fmt.Println("failed to delete from AspNetUsers table due to ", err)
		} */
		if affRows == 1 {
			fmt.Println("successfully deleted from AspNetUsers table")
		}
		fmt.Println("User creation unsuccessful due to ", err)
		res := h.Response {
			Status: "error",
			Message:"User creation unsuccessful, pls try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	fmt.Println("AspNetUserRolesInserted User id is ", AspNetUserRolesInsertedUserId)

	//Execute data_records table Insert
	dataRecordId,err := dataRecordsInsert()
	if err != nil {
		fmt.Println("inserting into data_records failed due to ", err)
		affRows1,_ := aspNetUserRolesDelete(AspNetUserRolesInsertedUserId)
		affRows2, _ := aspNetUsersDelete(userId)

		if affRows1 == 1 && affRows2 == 1{
			fmt.Println("successfully deleted from aspNetUserRoles table and AspNetUsers table")
		}
		fmt.Println("User creation unsuccessful due to ", err)
		res := h.Response {
			Status: "error",
			Message:"User creation unsuccessful, pls try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	fmt.Println("data records id is ", dataRecordId)

	//Execute profiles table insert 
	profilesId, err := profilesInsert(userId,lastname,othername,dataRecordId)
	if err != nil {
		fmt.Println("inserting into profiles table failed due to ", err)
		affRows1, _ := aspNetUsersDelete(userId)
		affRows2,_ := aspNetUserRolesDelete(AspNetUserRolesInsertedUserId)
		affRows3, _ := dataRecordsDelete(dataRecordId)
		if affRows1 == 1 && affRows2 == 1 && affRows3 == 1{
			fmt.Println("successfully deleted from aspNetUserRoles table and AspNetUsers table and data_records table")
		}
		fmt.Println("User creation unsuccessful due to ", err)
		res := h.Response {
			Status: "error",
			Message:"User creation unsuccessful, pls try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	fmt.Println("profiles id is ", profilesId)

	//Execute user_devices table insert
	userDevicesId, err := userDevicesInsert(userId, deviceName, deviceModel, deviceUuid, deviceImei, dataRecordId)
	if err != nil {
		fmt.Println("inserting into user_devices table failed due to ", err)
		affRows1, _ := aspNetUsersDelete(userId)
		affRows2,_ := aspNetUserRolesDelete(AspNetUserRolesInsertedUserId)
		affRows3, _ := dataRecordsDelete(dataRecordId)
		affRows4, _ := profilesDelete(profilesId)
		if affRows1 == 1 && affRows2 == 1 && affRows3 == 1 && affRows4 == 1{
			fmt.Println("successfully deleted from aspNetUserRoles table and AspNetUsers table and data_records table and profiles table")
		}
		fmt.Println("User creation unsuccessful due to ", err)
		res := h.Response {
			Status: "error",
			Message:"User creation unsuccessful, pls try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	fmt.Println("user_devices id ", userDevicesId)

	//Execute user_audits table insert
	userAuditId, err := userAuditsInsert(userId, ipAddress, "mobile","AccountCreated")
	if err != nil {
		fmt.Println("inserting into profiles table failed due to ", err)
		affRows1, _ := aspNetUsersDelete(userId)
		affRows2,_ := aspNetUserRolesDelete(AspNetUserRolesInsertedUserId)
		affRows3, _ := dataRecordsDelete(dataRecordId)
		affRows4, _ := profilesDelete(profilesId)
		affRows5, _ := userDevicesDelete(userDevicesId)
		if affRows1 == 1 && affRows2 == 1 && affRows3 == 1 && affRows4 == 1 && affRows5 == 1{
			fmt.Println("successfully deleted from aspNetUserRoles table and AspNetUsers table and data_records table and profiles table and user_devices table")
		}
		fmt.Println("User creation unsuccessful due to ", err)
			res := h.Response {
				Status: "error",
				Message:"User creation unsuccessful, pls try again",
			}
		return c.JSON(http.StatusInternalServerError, res)
	}
	fmt.Println("user_audits id ", userAuditId)

	res := h.Response {
		Status: "success",
		Message:"User creation successful, congratulations",
	}
	return c.JSON(http.StatusCreated, res)
}

func aspNetUsersInsert(username,email,passHash,phone,phoneVeriStatus string) (string,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()
	var verify bool
	if phoneVeriStatus == "1" {
		verify = true
	}else{
		verify = false
	}
	var userid string
	queryString := `INSERT INTO "AspNetUsers"("Id","Email","PasswordHash","SecurityStamp","PhoneNumber","PhoneNumberConfirmed","UserName") VALUES($1,$2,$3,$4,$5,$6) RETURNING "Id"`
	//re, err := con.Db.Exec(queryString,h.GenerateUuid(),email,passHash, h.GenerateUuid(),phone, username)//.Scan(&userid)
	err = con.Db.QueryRow(queryString,h.GenerateUuid(),email,passHash, h.GenerateUuid(),phone,verify, username).Scan(&userid)
	fmt.Println("userid inserted for AspNetUsers is ", userid)
	if err != nil {
		fmt.Println("error encountered is ", err)
		/* if strings.Contains(err.Error(),`pq: duplicate key value violates unique constraint "email_unique"`) {
			res := h.Response {
				Error: false,
				Message:"Email Address already Exists",
			}
			return c.JSON(http.StatusInternalServerError, res)
		} */
		/* if strings.Contains(err.Error(),`pq: duplicate key value violates unique constraint "phone_unique"`) {
			res := h.Response {
				Error: false,
				Message:"Phone Number already Exists",
			}
			return c.JSON(http.StatusInternalServerError, res)
		} */
		if s.Contains(err.Error(),`pq: duplicate key value violates unique constraint "AspNetUsers_UserName_key"`) {
			return "",errors.New("Username already Exists")
		}
		return "", err
	}
	return userid,nil
}

func aspNetUsersDelete(id string) (int64,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return -1,err
	}
	defer con.Close()
	deleteQuery := `DELETE FROM "AspNetUsers" WHERE "Id"=$1`
	re, err := con.Db.Exec(deleteQuery, id)
	if err != nil {
		return -1,err
	}
	affRows, _ := re.RowsAffected()
	return affRows,nil
}

func aspNetUserRolesInsert(userId, roleId string) (string, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
		//return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()
	var insertedUserId,insertedRoleId string
	insertQuery := `INSERT INTO "AspNetUserRoles"("UserId","RoleId") VALUES($1,$2) RETURNING "UserId","RoleId"`
	err = con.Db.QueryRow(insertQuery,userId,roleId).Scan(&insertedUserId,&insertedRoleId)
	if err != nil {
		fmt.Println("error encountered while inserting into AspNetUserRoles is ", err)
		return "",err
	}
	//check if the row was inserted successfully
	if insertedUserId == "" && insertedRoleId == "" {
		return "", errors.New("inserting into AspNetUserRoles failed")
	}
	return insertedUserId, nil
}

func aspNetUserRolesDelete(userId string) (int64,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return -1,err
	}
	defer con.Close()
	deleteQuery := `DELETE FROM "AspNetUserRoles" WHERE "UserId"=$1`
	re, err := con.Db.Exec(deleteQuery, userId)
	if err != nil {
		return -1,err
	}
	affRows, _ := re.RowsAffected()
	return affRows,nil
}

func dataRecordsInsert() (string, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()
	var insertedId string
	insertQuery := `INSERT INTO "data_records"("Id","CreatedBy","DateCreated") VALUES($1,$2,$3) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),"Self",time.Now()).Scan(&insertedId)
	if err != nil {
		fmt.Println("error encountered while inserting into data_records due to ", err)
		return "",err
	}
	return insertedId, nil
}

func dataRecordsDelete(id string) (int64, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return -1,err
	}
	defer con.Close()
	deleteQuery := `DELETE FROM "data_records" WHERE "Id"=$1`
	re, err := con.Db.Exec(deleteQuery, id)
	if err != nil {
		return -1,err
	}
	affRows, _ := re.RowsAffected()
	return affRows,nil
}

func profilesInsert(userId,lastName,otherName,dataRecordId string) (string, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()
	var insertedId string
	insertQuery := `INSERT INTO "profiles"("Id","UserId","LastName","OtherNames","DataRecordId") VALUES($1,$2,$3,$4,$5) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),userId,lastName,otherName,dataRecordId).Scan(&insertedId)
	if err != nil {
		fmt.Println("error encountered while inserting into profiles table due to ", err)
		return "",err
	}
	return insertedId, nil
}

func profilesDelete(id string) (int64, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return -1,err
	}
	defer con.Close()
	deleteQuery := `DELETE FROM "profiles" WHERE "Id"=$1`
	re, err := con.Db.Exec(deleteQuery, id)
	if err != nil {
		return -1,err
	}
	affRows, _ := re.RowsAffected()
	return affRows,nil
}

func userAuditsInsert(userId,ipAddress,deviceAgent,auditEvent string) (string,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()
	var insertedId string
	insertQuery := `INSERT INTO "user_audits"("Id","UserId","UserClient","IpAddress","AuditEvent","TimeStamp") VALUES($1,$2,$3,$4,$5,$6) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),userId,deviceAgent,ipAddress,auditEvent, time.Now()).Scan(&insertedId)
	if err != nil {
		fmt.Println("error encountered while inserting into profiles table due to ", err)
		return "",err
	}
	return insertedId, nil
} 

func userAuditsDelete(id string) (int64, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return -1,err
	}
	defer con.Close()
	deleteQuery := `DELETE FROM "user_audits" WHERE "Id"=$1`
	re, err := con.Db.Exec(deleteQuery, id)
	if err != nil {
		return -1,err
	}
	affRows, _ := re.RowsAffected()
	return affRows,nil
}

func userDevicesInsert(userId,deviceName,deviceModel,deviceUuid,deviceImei,dataRecordId string) (string,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()
	var insertedId string
	insertQuery := `INSERT INTO "user_devices"("Id","UserId","DeviceName","DeviceModel","DeviceUUID","DeviceIMEI","DataRecordId") VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),userId,deviceName,deviceModel,deviceUuid,deviceImei,dataRecordId).Scan(&insertedId)
	if err != nil {
		fmt.Println("error encountered while inserting into user_devices table due to ", err)
		return "",err
	}
	return insertedId, nil
}

func userDevicesDelete(id string) (int64, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return -1,err
	}
	defer con.Close()
	deleteQuery := `DELETE FROM "user_devices" WHERE "Id"=$1`
	re, err := con.Db.Exec(deleteQuery, id)
	if err != nil {
		return -1,err
	}
	affRows, _ := re.RowsAffected()
	return affRows,nil
}

func isEmailConfirmed(username string) (string,string,bool) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("transactionhandlers.go::isEmailConfirmed()::error in connecting to database due to ",err)
		return "","",false
		//return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()
	var id,email interface{}
	var emailConfirmed bool 
	var uId,uEmail string
	q := `SELECT "AspNetUsers"."Id","AspNetUsers"."Email","AspNetUsers"."EmailConfirmed" FROM "AspNetUsers" WHERE "UserName" = $1` 
	err = con.Db.QueryRow(q, username).Scan(&id,&email,&emailConfirmed)
	if err != nil {
		fmt.Println("transactionhandlers.go::isEmailConfirmed()::error in fetching user's id and email confirmed status from database due to ",err)
		return "","",false
	}
	if id != nil {
		uId = id.(string)
	}
	if email != nil {
		uEmail = email.(string)
	}
	if emailConfirmed == false {
		return uId,uEmail,false
	}
	return uId,uEmail,true
}

func isEmailExists(email string) (bool) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("userhandlers.go::isEmailExists()::error in connecting to database due to ",err)
		return false
	}
	defer con.Close()
	var existingEmail interface{}
	q := `SELECT "AspNetUsers"."Email" FROM "AspNetUsers" WHERE "Email" = $1` 
	err = con.Db.QueryRow(q, email).Scan(&existingEmail)
	if err != nil {
		fmt.Println("userhandlers.go::isEmailExists()::error in fetching email status from database due to ",err)
		return false
	}
	if existingEmail == nil {
		return false
	}
	return true
}

func sendConfirmationCode(userId,email string) error {	
	con, err := h.OpenConnection()
	if err != nil {
		return err
	}
	defer con.Close()
	var ruId string 
	code := s.ToLower(rand.RandStr(6, "alphanum"))
	//TODO, send this code to the client's email
	err = e.SendMail(email,code)
	if err != nil {
		fmt.Println("userhandlers.go::sendConfirmationCode():: error encountered while sending mail is ", err)
		return err
	}
	q := `UPDATE "profiles" SET "ResetPasswordConfirmationCode" = $1, "CodeSentAt" = $2 WHERE userId = $3 RETURNING "Id"`
	err = con.Db.QueryRow(q, code,time.Now()).Scan(&ruId)
	return err
}