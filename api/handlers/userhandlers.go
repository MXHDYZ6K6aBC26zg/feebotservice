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
	"encoding/json"
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
			Status: "error",
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
	confirmationCode := s.ToUpper(s.Trim(c.FormValue("code")," "))
	password := c.FormValue("password")

	if username == "" || email == "" || confirmationCode == "" || password == "" {
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
	codeSent,timeSent,resetCount := getConfrimationCodeAndTimeSent(uId)
	if codeSent != confirmationCode {
		res := h.Response {
			Status: "error",
			Message:"Confirmation code is invalid!",
		}
		return c.JSON(http.StatusBadRequest, res)
	}
	now := time.Now()
	diff := now.Sub(timeSent)
	if mins := int(diff.Minutes()); mins > 10 {
		res := h.Response {
			Status: "error",
			Message:"Sorry!!...Confirmation code has expired!",
		}
		return c.JSON(http.StatusRequestTimeout, res)
	}
	passHash,err := h.BcryptHashPassword(password)
	if err != nil {
		fmt.Println("userhandlers.go::ResetPassword():: error in hashing provided password due to:", err)
		res := h.Response {
			Status: "error",
			Message:"Oops, something went wrong please try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	err = updatePassword(uId, passHash, resetCount)
	if err != nil {
		fmt.Println("userhandlers.go::ResetPassword():: error in updating hashed password due to:", err)
		res := h.Response {
			Status: "error",
			Message:"Oops, something went wrong please try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	res := h.Response {
		Status: "success",
		Message:"Password Reset Successful",
	}
	return c.JSON(http.StatusOK, res)
}

func getConfrimationCodeAndTimeSent(userId string) (string, time.Time,int) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("userhandlers.go::getConfrimationCodeAndTimeSent():: error in connectiong to database due to :", err)
	}
	defer con.Close()
	var iConfCode, iTimeSent interface{}
	var code string 
	var timeSent time.Time
	var resetCount int
	q := `SELECT "ResetPasswordConfirmationCode", "ResetPasswordCount", "CodeSentAt" FROM "profiles" WHERE "UserId" = $1`
	err = con.Db.QueryRow(q, userId).Scan(&iConfCode,&resetCount,&iTimeSent)
	if err != nil {
		fmt.Println("userhandlers.go::getConfrimationCodeAndTimeSent():: error in select from profiles table due to :", err)
	}
	if iConfCode != nil {
		code = iConfCode.(string)
	}
	if iTimeSent != nil {
		timeSent = iTimeSent.(time.Time)
	}
	return code,timeSent,resetCount
}

func updatePassword(userId,hashedPassword string, resetCount int) error {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("userhandlers.go::updatePassword():: error in connecting to database due to :", err)
		return err
	}
	defer con.Close()
	var uId string
	q := `UPDATE "AspNetUsers" SET "PasswordHash" = $1 WHERE "Id" = $2 RETURNING "Id"`
	err = con.Db.QueryRow(q, hashedPassword, userId).Scan(&uId)
	if err != nil {
		fmt.Println("userhandlers.go::updatePassword():: error in updating PasswordHash of AspNetUsers table due to :", err)
		return err
	}
	pq := `UPDATE "profiles" SET "ResetPasswordCount" = $1 WHERE "UserId" = $2 RETURNING "UserId"`
	err = con.Db.QueryRow(pq, resetCount + 1, userId).Scan(&uId)
	if err != nil {
		fmt.Println("userhandlers.go::updatePassword():: error in updating ResetPasswordCount of profiles table due to :", err)
		return err
	}
	return nil
}

func UpdateVerifiedPhoneNumber(c echo.Context) error {
	username := s.ToLower(s.Trim(c.FormValue("username")," "))
	phone := s.ToLower(s.Trim(c.FormValue("phone")," "))

	if username == "" || phone == "" {
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)
	}

	con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()

	var userid string
	q := `UPDATE "AspNetUsers" SET "PhoneNumber" = $1, "PhoneNumberConfirmed" = $2 WHERE "UserName" = $3 RETURNING "Id"`
	err = con.Db.QueryRow(q,phone,true,username).Scan(&userid)	
	if err != nil {
		fmt.Println("error encountered is ", err)
		if s.Contains(err.Error(),`pq: duplicate key value violates unique constraint "AspNetUsers_PhoneNumber_key"`) {
			res := h.Response {
				Status: "error",
				Message: "Sorry...the phone number you verified is already in use by another user, please provide another phone number",
			}
			return c.JSON(http.StatusBadRequest, res)
		}
		res := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusBadRequest, res)
	}
	res := h.Response {
		Status: "success",
		Message:"Phone number verification done successfully",
	}
	return c.JSON(http.StatusOK, res)
}

func CreateUser(c echo.Context) error {
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
	if othername == "" || lastname == "" || username == "" || password == "" || ipAddress == "" || email == "" || phone == "" || deviceImei == "" || deviceUuid == "" || deviceModel == ""{
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)	
	}

	con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()

	if emailStatus := isEmailExists(email); emailStatus {
		res := h.Response {
			Status: "error",
			Message:"Email address already exist, register with another email address",
		}
		return c.JSON(http.StatusForbidden, res)
	}

	if phoneStatus := isPhoneExists(phone); phoneStatus {
		res := h.Response {
			Status: "error",
			Message:"Phone number already exist, register with another phone number",
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
	//fmt.Printf("Body: username-%s, email-%s, lastname-%s,phone-%s,othername-%s,deviceName-%s,deviceModel-%s,deviceUuid-%s,deviceImei-%s,ipAddress-%s\n", 
	//	username,email,lastname,phone,othername,deviceName,deviceModel,deviceUuid,deviceImei,ipAddress)

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
	fmt.Println("AspNetUsers id is ", userId)

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

	//TODO: Send the user a confirmation code to inorder to confirm the email address registered
	uDetail := map[string]string {
		"user_id": userId,
	}
	bs,_:= json.Marshal(uDetail)

	res := h.Response {
		Status: "success",
		Message:"You have successfully registered, a confirmation code has been sent to your email address",
		Data: bs,
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
	queryString := `INSERT INTO "AspNetUsers"("Id","Email","PasswordHash","SecurityStamp","PhoneNumber","PhoneNumberConfirmed","UserName") VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING "Id"`
	//re, err := con.Db.Exec(queryString,h.GenerateUuid(),email,passHash, h.GenerateUuid(),phone, username)//.Scan(&userid)
	err = con.Db.QueryRow(queryString,h.GenerateUuid(),email,passHash, h.GenerateUuid(),phone,verify, username).Scan(&userid)
	fmt.Println("userid inserted for AspNetUsers is ", userid)
	if err != nil {
		fmt.Println("error encountered is ", err)
		if s.Contains(err.Error(),`pq: duplicate key value violates unique constraint "AspNetUsers_Email_key"`) {
			return "",errors.New("Email Address already Exists")		
		}
		if s.Contains(err.Error(),`pq: duplicate key value violates unique constraint "AspNetUsers_PhoneNumber_key"`) {
			return "",errors.New("Phone Number already Exists")
		}
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

func isPhoneExists(phone string) (bool) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("userhandlers.go::isPhoneExists()::error in connecting to database due to ",err)
		return false
	}
	defer con.Close()
	var existingPhone interface{}
	q := `SELECT "AspNetUsers"."PhoneNumber" FROM "AspNetUsers" WHERE "PhoneNumber" = $1` 
	err = con.Db.QueryRow(q, phone).Scan(&existingPhone)
	if err != nil {
		fmt.Println("userhandlers.go::isPhoneExists()::error in fetching phone number status from database due to ",err)
		return false
	}
	if existingPhone == nil {
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
	code := s.ToUpper(rand.RandStr(6, "alphanum"))
	fmt.Println("code to be emailed to user is :", code)
	msgBody := fmt.Sprintf(`You have requested to reset your FeeRack App Login password, Enter the confirmation code below within 2 hours as the code expires after 2 hours from the time recieved. Ignore if you didn't make this request. \n %s`,code)
	//send this code to the user's email
	mailObj := e.MailConfig("feeracksolution@gmail.com", "Password1@", email, "FeeRack Reset Password Confirmation code", msgBody)
	err = e.SendMail(mailObj)
	if err != nil {
		fmt.Println("userhandlers.go::sendConfirmationCode():: error encountered while sending mail is ", err)
		return err
	}
	//update the sent code to the user's profile table on the db
	q := `UPDATE "profiles" SET "ResetPasswordConfirmationCode" = $1, "CodeSentAt" = $2 WHERE "UserId" = $3 RETURNING "Id"`
	err = con.Db.QueryRow(q, code,time.Now(),userId).Scan(&ruId)
	fmt.Println("user id that just reset mail is :", ruId)
	return err
}