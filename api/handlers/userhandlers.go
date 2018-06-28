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
	err := SendConfirmationCode(uId, email,"passwordReset","","")
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
	codeSent,timeSent,resetCount := getPasswordResetConfrimationCodeAndTimeSent(uId)
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

func getPasswordResetConfrimationCodeAndTimeSent(userId string) (string, time.Time,int) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("userhandlers.go::getConfrimationCodeAndTimeSent():: error in connectiong to database due to :", err)
	}
	defer con.Close()
	var iConfCode, iTimeSent interface{}
	var code string 
	var timeSent time.Time
	var resetCount int
	q := `SELECT "ResetPasswordConfirmationCode", "ResetPasswordCount", "ResetPasswordCodeSentAt" FROM "profiles" WHERE "UserId" = $1`
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

func SendEmailConfirmationCode(c echo.Context) error {
	userId := s.ToLower(s.Trim(c.FormValue("userId")," "))
	email := s.ToLower(s.Trim(c.FormValue("email")," "))

	if userId == "" || email == "" {
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)	
	}
	err := SendConfirmationCode(userId, email, "emailConfirmation","","")
	if err != nil {
		fmt.Println("userhandlers.go::SendEmailConfirmationCode():: error encountered in sending email confirmation code is :", err)
		res := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	res := h.Response {
		Status: "success",
		Message: fmt.Sprintf("Email Confirmation Code sent successfully to %s\n",email),
	}
	return c.JSON(http.StatusOK, res)
}

func ConfirmEmailAddress(c echo.Context) error {
	userId := s.ToLower(s.Trim(c.FormValue("userId")," "))
	confirmationCode := s.ToUpper(s.Trim(c.FormValue("code")," "))

	if userId == "" || confirmationCode == "" {
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)	
	}
	codeSent,timeSent := getEmailConfrimationCodeAndTimeSent(userId)
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
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("userhandlers.go::getEmailConfrimationCodeAndTimeSent():: error in connectiong to database due to :", err)
	}
	defer con.Close()
	var userid string
	q := `UPDATE "AspNetUsers" SET "EmailConfirmed" = $1 WHERE "Id" = $2 RETURNING "Id"`
	err = con.Db.QueryRow(q,true,userId).Scan(&userid)	
	if err != nil {
		res := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, res)	
	}
	res := h.Response {
		Status: "success",
		Message:"Email Confirmation done Successfully",
	}
	return c.JSON(http.StatusOK, res)
}

func getEmailConfrimationCodeAndTimeSent(userId string) (string, time.Time) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("userhandlers.go::getEmailConfrimationCodeAndTimeSent():: error in connectiong to database due to :", err)
	}
	defer con.Close()
	var iConfCode, iTimeSent interface{}
	var code string 
	var timeSent time.Time
	q := `SELECT "EmailConfirmationCode","EmailConfirmationCodeSentAt" FROM "profiles" WHERE "UserId" = $1`
	err = con.Db.QueryRow(q, userId).Scan(&iConfCode,&iTimeSent)
	if err != nil {
		fmt.Println("userhandlers.go::getEmailConfrimationCodeAndTimeSent():: error in select from profiles table due to :", err)
	}
	if iConfCode != nil {
		code = iConfCode.(string)
	}
	if iTimeSent != nil {
		timeSent = iTimeSent.(time.Time)
	}
	return code,timeSent
}

func UpdateVerifiedPhoneNumber(c echo.Context) error {
	userId := s.ToLower(s.Trim(c.FormValue("userId")," "))
	phone := s.ToLower(s.Trim(c.FormValue("phone")," "))
	phoneVerificationId := c.FormValue("phoneVerificationId")

	if userId == "" || phone == "" {
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

	var userid,profilesId string
	q := `UPDATE "AspNetUsers" SET "PhoneNumber" = $1, "PhoneNumberConfirmed" = $2 WHERE "Id" = $3 RETURNING "Id"`
	err = con.Db.QueryRow(q,phone,true,userId).Scan(&userid)	
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
	pq := `UPDATE "profiles" SET "PhoneNumberVerificationId" = $1 WHERE "UserId" = $2 RETURNING "Id"`
	err = con.Db.QueryRow(pq,phoneVerificationId,userId).Scan(&profilesId)	
	if err != nil {
		res := h.Response {
			Status: "error",
			Message: err.Error(),//"Error encountered, please try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	uPhone := map[string]string {
		"phone_number" : phone,
	}
	bs,_:= json.Marshal(uPhone)
	res := h.Response {
		Status: "success",
		Message: "Phone number verification done successfully",
		Data: bs,
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
	phoneVerificationId := c.FormValue("phoneVerificationId")

	//Check for complete credentials
	if othername == "" || lastname == "" || username == "" || password == "" || ipAddress == "" || email == "" || phone == "" || deviceImei == "" || deviceUuid == "" || deviceModel == ""{
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)	
	}

	if isDeviceEnabled := isDeviceEnabled(deviceUuid,deviceImei); !isDeviceEnabled {
		res := h.Response {
			Status: "error",
			Message: "This Device has been disabled, contact admin",
		}
		return c.JSON(http.StatusLocked, res)
	}

	if phoneVerificationStatus == "1" && phoneVerificationId == "" {
		res := h.Response {
			Status: "error",
			Message:"Phone number is verified, expecting phone number verification id",
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
	isNew := true 
	fUId,count := CheckUuidCount(deviceUuid, deviceImei)
	if count == 1 {
		res := h.Response {
			Status: "error",
			Message:"This device has been used to create an account in feerack solution. Would you want to ignore and continue? To continue means that you would want to lock the existing account",
		}
		return c.JSON(http.StatusConflict, res)
	}
	if count == 2 {
		go lockUserAccount(fUId)
		/* if err != nil {
			fmt.Println("error encountered while trying to lock user account for count equal to 2 is ", err)
		} */
		isNew = false
	}
	if count > 2 {
		go lockUserAccount(fUId)
		/* if err != nil {
			fmt.Println("error encountered while trying to lock user account for count greater than 2 is ", err)
		} */
		go disableUserDevice(deviceUuid, deviceImei)
		/* if err != nil {
			fmt.Println("error encountered while trying to disable user device for count greater than 2 is ", err)
		} */
		res := h.Response {
			Status: "error",
			Message: "Device locked...multiple accounts not allowed per device, contact admin",
		}
		return c.JSON(http.StatusLocked, res)
	}
	
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
			//fmt.Println("inserted 'User' role as",uRoleId)
		}
		fmt.Println("main error encountered is ", err)
	}
	
	if roleId != nil {
		//fmt.Println("Selected 'User' role as",roleId.(string))
		uRoleId = roleId.(string)
	} 
	//fmt.Println("user role Id is ", uRoleId)

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
			Message: err.Error(),//"User creation unsuccessful, pls try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	//fmt.Println("AspNetUserRolesInserted User id is ", AspNetUserRolesInsertedUserId)

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
			Message: err.Error(),//"User creation unsuccessful, pls try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	//fmt.Println("data records id is ", dataRecordId)

	//Execute profiles table insert 
	profilesId, err := profilesInsert(userId,lastname,othername,dataRecordId,phoneVerificationId)
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
			Message: err.Error(),//"User creation unsuccessful, pls try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	//fmt.Println("profiles id is ", profilesId)

	//Execute user_devices table insert
	userDevicesId, err := userDevicesInsertOrUpdate(userId, deviceName, deviceModel, deviceUuid, deviceImei, dataRecordId,isNew)
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
			Message: err.Error(),//"User creation unsuccessful, pls try again",
		}
		return c.JSON(http.StatusInternalServerError, res)
	}
	//fmt.Println("user_devices id ", userDevicesId)

	//Execute user_audits table insert
	_, err = userAuditsInsert(userId, ipAddress, "mobile","AccountCreated")
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
				Message: err.Error(),//"User creation unsuccessful, pls try again",
			}
		return c.JSON(http.StatusInternalServerError, res)
	}
	//fmt.Println("user_audits id ", userAuditId)

	//TODO: Send the user a confirmation code to his/her email address inorder to confirm the email address registered
	go SendConfirmationCode(userId, email, "emailConfirmation","","")

	uDetail := map[string]string {
		"user_id": userId,
		"phone_number" : phone,
	}
	bs,_:= json.Marshal(uDetail)

	res := h.Response {
		Status: "success",
		Message:"You have successfully registered, a confirmation code has been sent to your email address, use it to confirm your email address after login in.",
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
	err = con.Db.QueryRow(queryString,h.GenerateUuid(),email,passHash,h.GenerateUuid(),phone,verify, username).Scan(&userid)
	fmt.Println("userid inserted for AspNetUsers is ", userid)
	if err != nil {
		fmt.Println("error encountered while inserting for AspNetUsers is ", err)
		if s.Contains(err.Error(),`pq: duplicate key value violates unique constraint "AspNetUsers_Email_key"`) {
			return "",errors.New("Email Address already Exists")		
		}
		if s.Contains(err.Error(),`pq: duplicate key value violates unique constraint "AspNetUsers_PhoneNumber_key"`) {
			return "",errors.New("Phone Number already Exists")
		}
		if s.Contains(err.Error(),`pq: duplicate key value violates unique constraint "AspNetUsers_UserName_key"`) {
			return "",errors.New("Username already Exists")
		}
		return "", errors.New("User creation unsuccessful, pls try again")
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
		return "",errors.New("User creation unsuccessful, pls try again")
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
		return "",errors.New("User creation unsuccessful, pls try again")
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

func profilesInsert(userId,lastName,otherName,dataRecordId,phoneVerificationId string) (string, error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()
	var insertedId string
	insertQuery := `INSERT INTO "profiles"("Id","UserId","LastName","OtherNames","DataRecordId","PhoneNumberVerificationId") VALUES($1,$2,$3,$4,$5,$6) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),userId,lastName,otherName,dataRecordId,phoneVerificationId).Scan(&insertedId)
	if err != nil {
		fmt.Println("error encountered while inserting into profiles table due to ", err)
		return "",errors.New("User creation unsuccessful, pls try again")
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
		return "",errors.New("User creation unsuccessful, pls try again")
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

func userDevicesInsertOrUpdate(userId,deviceName,deviceModel,deviceUuid,deviceImei,dataRecordId string, isNew bool) (string,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()
	var id string
	if isNew == true {
		insertQuery := `INSERT INTO "user_devices"("Id","UserId","DeviceName","DeviceModel","DeviceUUID","DeviceIMEI","DataRecordId") VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING "Id"`
		err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),userId,deviceName,deviceModel,deviceUuid,deviceImei,dataRecordId).Scan(&id)
		if err != nil {
			fmt.Println("error encountered while inserting into user_devices table due to ", err)
			/* if s.Contains(err.Error(),`pq: duplicate key value violates unique constraint "UserDevices_DeviceUUID_key"`) {
				return "",errors.New("This device has been used to create account before. Please contact admin")		
			}
			if s.Contains(err.Error(),`pq: duplicate key value violates unique constraint "UserDevices_DeviceIMEI_key"`) {
				return "",errors.New("This device has been used to create account before. Please contact admin")
			} */
			return "", err//errors.New("User creation unsuccessful, pls try again")
		}
	}else {
		updateQuery := `UPDATE "user_devices" SET "UserId" = $1, "DeviceName" = $2, "DeviceModel" = $3, "DataRecordId" = $4 WHERE "DeviceUUID" = $5 AND "DeviceIMEI" = $6 RETURNING "Id"`
		err = con.Db.QueryRow(updateQuery,userId,deviceName,deviceModel,dataRecordId,deviceUuid,deviceImei).Scan(&id)
		if err != nil {
			fmt.Println("error encountered while updating into user_devices table due to ", err)
			return "", err//errors.New("User creation unsuccessful, pls try again")
		}
	}	
	return id, nil
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

func isEmailAndPhoneConfirmed(userId string) (string,bool,bool) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("transactionhandlers.go::isEmailAndPhoneConfirmed()::error in connecting to database due to ",err)
		return "",false,false
		//return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()
	var email interface{}
	var emailConfirmed,phoneConfirmed bool 
	var uEmail string
	q := `SELECT "AspNetUsers"."Email","AspNetUsers"."EmailConfirmed","AspNetUsers"."PhoneNumberConfirmed" FROM "AspNetUsers" WHERE "Id" = $1` 
	err = con.Db.QueryRow(q, userId).Scan(&email,&emailConfirmed,&phoneConfirmed)
	if err != nil {
		fmt.Println("transactionhandlers.go::isEmailConfirmed()::error in fetching user's id and email confirmed status from database due to ",err)
		return "",false,false
	}
	if email != nil {
		uEmail = email.(string)
	}
	/* 	if emailConfirmed == false {
		return uId,uEmail,false
	} */
	return uEmail,emailConfirmed,phoneConfirmed
}

func isDeviceEnabled(uuid,imei string) (bool) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("userhandlers.go::isDeviceDisabled()::error in connecting to database due to ",err)
		return false
	}
	defer con.Close()
	var status bool
	q := `SELECT "user_devices"."Enabled" FROM "user_devices" WHERE "DeviceUUID" = $1 AND "DeviceIMEI" = $2` 
	err = con.Db.QueryRow(q, uuid,imei).Scan(&status)
	if err != nil {
		if s.Contains(fmt.Sprintf("%v", err), "no rows") == true {
			return true	
		}
		fmt.Println("userhandlers.go::isDeviceDisabled()::error in fetching device enabled status from database due to ",err)
		return true
	}
	return status
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

func CheckUuidCount(uuid,imei string) (string,int) {
	con, err := h.OpenConnection()
	if err != nil {
		fmt.Println("userhandlers.go::CheckUuidCount()::error in connecting to database due to ",err)
		return "",-1
	}
	defer con.Close()
	var count, incrementedCount int
	var userId string
	q := `SELECT "user_devices"."DeviceCount" FROM "user_devices" WHERE user_devices."DeviceUUID" = $1 AND user_devices."DeviceIMEI" = $2` 
	err = con.Db.QueryRow(q, uuid,imei).Scan(&count)
	if err != nil {
		fmt.Println("userhandlers.go::CheckUuidCount()::error in fetching phone number status from database due to ",err)
		return "",-1
	}
	incrementedCount = count + 1
	//update user_devices table with the incremented device count
	updateQuery := `UPDATE "user_devices" SET "DeviceCount"= $1 WHERE user_devices."DeviceUUID" = $2 AND user_devices."DeviceIMEI" = $3 RETURNING "UserId"`
	err = con.Db.QueryRow(updateQuery, incrementedCount,uuid,imei).Scan(&userId)
	if err != nil {
		fmt.Println("userhandlers.go::CheckUuidCount()::user_devices.DeviceUUID update error encountered is ", err)
		return "",-1
	}
	return userId,incrementedCount
}

func lockUserAccount(userId string) error {
	con, err := h.OpenConnection()
	if err != nil {
		return err
	}
	defer con.Close()
	//update AspNetUsers table with the incremented count
	updateQuery := `UPDATE "AspNetUsers" SET "LockoutEnabled"= $1 WHERE "Id" = $2 RETURNING "Id"`
	err = con.Db.QueryRow(updateQuery, true,userId).Scan(&userId)
	if err != nil {
		fmt.Printf("userhandlers.go::lockUserAccount()::error encountered while trying to lock user account for user id - [%s] is %v\n",userId, err)
		return err
	}
	return nil
}

func disableUserDevice(uuid,imei string) error {
	con, err := h.OpenConnection()
	if err != nil {
		return err
	}
	defer con.Close()
	var id string
	//update user_devices table and set enabled to false
	updateQuery := `UPDATE "user_devices" SET "Enabled"= $1 WHERE "DeviceUUID" = $2 AND "DeviceIMEI" = $3 RETURNING "Id"`
	err = con.Db.QueryRow(updateQuery, false,uuid,imei).Scan(&id)
	if err != nil {
		fmt.Println("userhandlers.go::disableUserDevice()::update error encountered is ", err)
		return err
	}
	return nil
}

func SendConfirmationCode(userId,email,purpose,body,subjectParam string) error {	
	con, err := h.OpenConnection()
	if err != nil {
		return err
	}
	defer con.Close()
	var ruId,msgBody,subject,dbCodeColumnName, dbTimeSentColumnName string 
	code := s.ToUpper(rand.RandStr(6, "alphanum"))
	fmt.Println("code to be emailed to user is :", code)
	if purpose == "passwordReset" {
		msgBody = fmt.Sprintf(`You have requested to reset your FeeRack App Login password, Enter the confirmation code below within 10 minutes as the code expires after 10 minutes from the time recieved. Ignore if you didn't make this request. %s`,code)
		//send this code to the user's email
		subject = "FeeRack solution Reset Password Confirmation code"
		dbCodeColumnName = "ResetPasswordConfirmationCode"
		dbTimeSentColumnName = "ResetPasswordCodeSentAt"
	}
	if purpose == "emailConfirmation" {
		msgBody = fmt.Sprintf(`Enter the confirmation code below within 10 minutes as the code expires after 10 minutes from the time recieved inorder to confirm your email address. %s`,code)
		//send this code to the user's email
		subject = "FeeRack solution Email address Confirmation code"
		dbCodeColumnName = "EmailConfirmationCode"
		dbTimeSentColumnName = "EmailConfirmationCodeSentAt"
	}
	if purpose == "webMail" {
		subject = subjectParam
		msgBody = body
	}
	mailObj := e.MailConfig("feeracksolution@gmail.com", "Password1@", email, subject, msgBody)
	err = e.SendMail(mailObj)
	if err != nil {
		fmt.Println("userhandlers.go::sendConfirmationCode():: error encountered while sending mail is ", err)
		return err
	}
	//update the sent code to the user's profile table on the db
	q := fmt.Sprintf(`UPDATE "profiles" SET "%s" = $1, "%s" = $2 WHERE "UserId" = $3 RETURNING "Id"`,dbCodeColumnName,dbTimeSentColumnName)
	err = con.Db.QueryRow(q, code,time.Now(),userId).Scan(&ruId)
	fmt.Println("user id that was sent mail is :", ruId)
	return err
}
