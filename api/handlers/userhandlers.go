package handlers

import (
	"net/http"
	h "github.com/kenmobility/feezbot/helper"
	"fmt"
	"github.com/labstack/echo"
	"encoding/json"
	"io/ioutil"
	"strings"
	"unicode"
	"errors"
	"time"
)

type User struct {
	Details UserDetail `json:"createUser"`
}

type UserDetail struct {
	Email    	string 		`json:"email"`
	Password 	string 		`json:"password"`
	Phone    	string 		`json:"phone"`
	UserName 	string 		`json:"userName"`
	LastName 	string 		`json:"lastName"`
	OtherName 	string 		`json:"otherName"`
	DeviceName 	string 		`json:"deviceName"`
	DeviceModel string 		`json:"deviceModel"`
	DeviceUUID 	string 		`json:"deviceUuid"`
	DeviceIMEI 	string 		`json:"deviceImei"`
	IpAddress 	string 		`json:"ipAddress"`	
}

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

func CreateUser(c echo.Context) error {
	con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()

	var u User
	defer c.Request().Body.Close()
	b,err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Printf("failed to read request body due to : %s", err)
		return c.String(http.StatusInternalServerError, "error in reading the request body")
	}
	fmt.Printf("the raw json request is %s\n", b)
	err = json.Unmarshal(b, &u)
	if err != nil {
		fmt.Printf("failed to unmarshal json request body: %s", err)
		res := h.Response {
			Status: "error",
			Message:err.Error(),
		}
		return c.JSON(http.StatusBadRequest, res)
	}
	fmt.Printf("json object is : %#v\n", u)

	//Check for complete credentials
	if u.Details.UserName == "" || u.Details.Password == "" || u.Details.IpAddress == "" {
		res := h.Response {
			Status: "error",
			Message:"Invalid request format Or required Credentials not complete",
		}
		return c.JSON(http.StatusBadRequest, res)	
	}

	email := strings.ToLower(u.Details.Email)
	username := strings.ToLower(u.Details.UserName)
	phone := strings.ToLower(u.Details.Phone)
	lastname := strings.ToLower(u.Details.LastName)
	othername := strings.ToLower(u.Details.OtherName)
	deviceName := u.Details.DeviceName
	deviceModel := u.Details.DeviceModel
	deviceUuid := strings.ToLower(u.Details.DeviceUUID)
	deviceImei := strings.ToLower(u.Details.DeviceIMEI)
	ipAddress := strings.ToLower(u.Details.IpAddress)

	numComplete,numPresent,upperPresent,specialChar := verifyPassword(u.Details.Password)
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

	//passHash,err := h.CreateHash(u.Details.Password)
	passHash,err := h.BcryptHashPassword(u.Details.Password)
	if err != nil {
		fmt.Println("error in hashing password due to :", err)
	}
	fmt.Printf("Body: username-%s, email-%s, lastname-%s,phone-%s,othername-%s,deviceName-%s,deviceModel-%s,deviceUuid-%s,deviceImei-%s,ipAddress-%s\n", 
		username,email,lastname,phone,othername,deviceName,deviceModel,deviceUuid,deviceImei,ipAddress)

	//Execute AspNetUsers insert
	userId, err := aspNetUsersInsert(username, email, passHash, phone)
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
	userAuditId, err := userAuditsInsert(userId, ipAddress, "AccountCreated")
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


func aspNetUsersInsert(username,email,passHash,phone string) (string,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()
	var userid string
	queryString := `INSERT INTO "AspNetUsers"("Id","Email","PasswordHash","SecurityStamp","PhoneNumber","UserName") VALUES($1,$2,$3,$4,$5,$6) RETURNING "Id"`
	//re, err := con.Db.Exec(queryString,h.GenerateUuid(),email,passHash, h.GenerateUuid(),phone, username)//.Scan(&userid)
	err = con.Db.QueryRow(queryString,h.GenerateUuid(),email,passHash, h.GenerateUuid(),phone, username).Scan(&userid)
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
		if strings.Contains(err.Error(),`pq: duplicate key value violates unique constraint "AspNetUsers_UserName_key"`) {
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

func userAuditsInsert(userId,ipAddress,auditEvent string) (string,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()
	var insertedId string
	insertQuery := `INSERT INTO "user_audits"("Id","UserId","UserClient","IpAddress","AuditEvent","TimeStamp") VALUES($1,$2,$3,$4,$5,$6) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),userId,"mobile",ipAddress,auditEvent, time.Now()).Scan(&insertedId)
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
