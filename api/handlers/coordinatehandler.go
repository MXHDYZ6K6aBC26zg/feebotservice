package handlers

import(
	"fmt"
	"net/http"
	"github.com/labstack/echo"
	"io/ioutil"
	h "github.com/kenmobility/feezbot/helper"
	"encoding/json"
	"errors"
	"time"
)

type DeviceCoordinate struct {
	PushCoordinate struct {
		DeviceUUID  string `json:"device_uuid"`
		Email       string `json:"email"`
		Latitude    string `json:"latitude"`
		Longitude   string `json:"longitude"`
		PhoneNumber string `json:"phone_number"`
		UserID      string `json:"user_id"`
		Username    string `json:"username"`
	} `json:"pushCoordinate"`
}

func UserDeviceCoordinate(c echo.Context) error {
	var dc DeviceCoordinate
	defer c.Request().Body.Close()
	b,err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Printf("coordinatehandler.go::UserDeviceCoordinate()::failed to read request body due to : %s\n", err)
		r := h.Response {
			Status: "error",
			Message: err.Error(), //"error occured, please try again",//err.Error(),
		}
		return c.JSON(http.StatusBadRequest, r)
	}
	//fmt.Printf("the raw json request is %s\n", b)
	err = json.Unmarshal(b, &dc)
	if err != nil {
		fmt.Printf("coordinatehandler.go::UserDeviceCoordinate()::failed to unmarshal json request body: %s\n", err)
		r := h.Response {
			Status: "error",
			Message: err.Error(),//"error occured, please try again",
		}
		return c.JSON(http.StatusInternalServerError, r)
	}

	uuid := dc.PushCoordinate.DeviceUUID
	email := dc.PushCoordinate.Email
	latitude := dc.PushCoordinate.Latitude
	longitude := dc.PushCoordinate.Longitude
	phone := dc.PushCoordinate.PhoneNumber
	uId := dc.PushCoordinate.UserID
	username := dc.PushCoordinate.Username

	if uuid == "" || longitude == "" || latitude == "" {
		r := h.Response {
			Status: "error",
			Message:"Invalid request format Or required parameters not complete",
		}
		return c.JSON(http.StatusBadRequest, r)	
	}

	_,err = insertCoordinates(uId,username,phone,email,uuid,longitude,latitude)
	if err != nil {
		fmt.Printf("coordinatehandler.go::UserDeviceCoordinate()::failed to insert coordinates pushed due to : %s\n", err)
		r := h.Response {
			Status: "error",
			Message: err.Error(),//"error occured, please try again",//
		}
		return c.JSON(http.StatusBadRequest, r)
	}
	r := h.Response {
		Status: "success",
		Message: "coordinates inserted successfully",
	}
	return c.JSON(http.StatusOK, r)
}

func insertCoordinates(uId,username,phone,email,uuid,longitude,latitude string) (string,error) {
	con, err := h.OpenConnection()
	if err != nil {
		return "", err
		//return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()
	var insertedId string
	insertQuery := `INSERT INTO "user_device_coordinates"("Id","UserId","Username","PhoneNumber","Email","DeviceUUID","Longitude","Latitude","DateEntered") 
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "Id"`
	err = con.Db.QueryRow(insertQuery,h.GenerateUuid(),uId,username,phone,email,uuid,longitude,latitude,time.Now()).Scan(&insertedId)
	if err != nil {
		fmt.Println("coordinatehandler.go::UserDeviceCoordinate()::error encountered while inserting into user_device_coordinates; response is ", err)
		return "",err
	}
	//check if the row was inserted successfully
	if insertedId == "" {
		return "", errors.New("insertion into user_device_coordinates failed")
	}
	return insertedId, nil
}