package handlers

import (
	"github.com/labstack/echo"
	"net/http"
	"fmt"
	"encoding/json"
	
	s "strings"
	h "github.com/kenmobility/feezbot/helper"
)

type merchantsSummary struct {
	Info []merchantInfo `json:"merchantsInfo"`
}

type merchantInfo struct {
	MerchantID    	string 		`json:"merchantId"`
	Title 			string 		`json:"title"`
	Description    	string 		`json:"description"`
	Photo 			string 		`json:"photo"`
}

type merchantFees struct {
	MerchantCode  string `json:"merchant_code"`
	MerchantID    string `json:"merchant_id"`
	MerchantPhoto string `json:"merchant_photo"`
	MerchantTitle string `json:"merchant_title"`
	Fees []fees `json:"merchant_fees"`		
}

type fees struct {
	FeeAmount 					float64    		`json:"fee_amount"`
	FeeID     					string 			`json:"fee_id"`
	FeeTitle 	 				string 			`json:"fee_title"`
	FeeType   					string 			`json:"fee_type"`
	InstallmentEnabled 			bool			`json:"installment_allowed"`		
	Installments				int				`json:"number_of_installments"`
	MinimumAmount				float64			`json:"minimum_amount"`
	FirstInstallmentPercentage	float64			`json:"first_installment_percentage"`
	MerchantFeeId				string			`json:"merchant_fee_id"`
	FeeBearer					string			`json:"fee_bearer"`
}

func ShowMerchantsSummary(c echo.Context) error {
	con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()

	merchant := make([]merchantInfo,0)
	var id,title,description,photo interface{}
	var mId,mTitle,mDescription,mPhoto string
	q := `SELECT "Id","Title","Description","PhotoId" FROM "merchants" WHERE "Enabled" = $1`
	mRows,err := con.Db.Query(q,true)
	defer mRows.Close()
	if err != nil{
		if s.Contains(fmt.Sprintf("%v", err), "no records") == true {
			res := h.Response {
				Status: "error",
				Message:"No record found!",
			}
			return c.JSON(http.StatusOK, res)	
		}else{
			fmt.Println("merchanthandlers.go::ShowMerchantSummary()::Select From merchants table Failed due to:", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	for mRows.Next() {
		err = mRows.Scan(&id,&title,&description,&photo)
		if err != nil {
			fmt.Println("merchanthandlers.go::ShowMerchantSummary()::scanning merchants columns Failed due to:", err)
		}
		if id != nil {
			mId = id.(string)
		}
		if title != nil {
			mTitle = title.(string)
		}
		if description != nil {
			mDescription = description.(string)
		}
		if photo != nil {
			mPhoto = photo.(string)
		}
		m := merchantInfo {
			MerchantID: mId,
			Title : mTitle,
			Description: mDescription,
			Photo : mPhoto,
		}
		merchant = append(merchant, m)
	}
	ms := merchantsSummary {
		Info : merchant,
	}
	bs,_:= json.Marshal(ms)
	res := h.Response {
		Status: "success",
		Message: "Merchants fetched successfully",
		Data: bs,
	}
	return c.JSON(http.StatusOK,res)
}

func ShowMerchantFees(c echo.Context) error {
	merchantId := c.QueryParam("merchantId")

	con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()
	feesSlice := make([]fees,0)
	var imId,imTitle,imCode,imPhoto,ifTitle,ifType,imfId,imfAmount,installmentEnabled,howManyInstallment,feeBearer,firstInstPerc,imerchFeesId interface{}
	var smId,smTitle,smCode,smPhoto,sfTitle,sfType,sfId,merchfeeBearer,smerchFeesId string 
	var enabledInstallment bool
	var numberOfInstallment int64
	var mfAmount,firstInstallmentPerc float64

	q := `SELECT "merchants"."Id","merchants"."Title","merchants"."Code","merchants"."PhotoId","fees"."Title","fees"."Type","merchant_fees"."Id",
	"merchant_fees"."FeeId","merchant_fees"."Amount","merchant_fees"."EnabledForInstallment","merchant_fees"."HowManyInstallment",
	"merchant_fees"."FeeBearer","merchant_fees"."FirstInstallmentPercentage" FROM "merchants" INNER JOIN "merchant_fees" ON "merchant_fees"."MerchantId" = "merchants"."Id" INNER JOIN "fees" ON 
	 "merchant_fees"."FeeId" = "fees"."Id" WHERE "merchants"."Id" = $1 AND "merchant_fees"."Enabled" = $2`

	mFeeRows,err := con.Db.Query(q,merchantId,true)
	defer mFeeRows.Close()
	if err != nil{
		if s.Contains(fmt.Sprintf("%v", err), "no records") == true {
			res := h.Response {
				Status: "error",
				Message:"No record found!",
			}
			return c.JSON(http.StatusOK, res)	
		}else{
			fmt.Println("merchanthandlers.go::ShowMerchantFees()::ShowMerchantFees sql query Failed due to:", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	for mFeeRows.Next() {
		err = mFeeRows.Scan(&imId,&imTitle,&imCode,&imPhoto,&ifTitle,&ifType,&imerchFeesId,&imfId,&mfAmount,&installmentEnabled,&howManyInstallment,&feeBearer,&firstInstPerc)
		if err != nil {
			fmt.Println("merchanthandlers.go::ShowMerchantFees()::scanning query Failed due to:", err)
		}
		if imId != nil {
			smId = imId.(string)
		}
		if imTitle != nil {
			smTitle = imTitle.(string)
		}
		if imCode != nil {
			smCode = imCode.(string)
		}
		if imPhoto != nil {
			smPhoto = imPhoto.(string)
		}
		if ifTitle != nil {
			sfTitle = ifTitle.(string)
		}
		if ifType != nil {
			sfType = ifType.(string)
		}
		if imfId != nil {
			sfId = imfId.(string)
		}
		if imerchFeesId != nil {
			smerchFeesId = imerchFeesId.(string)
		}
		if imfAmount != nil {
			mfAmount = imfAmount.(float64)
		}
		if installmentEnabled != nil {
			enabledInstallment = installmentEnabled.(bool)
		}
		if howManyInstallment != nil {
			numberOfInstallment = howManyInstallment.(int64)
		}
		if feeBearer != nil {
			merchfeeBearer = feeBearer.(string)
		}
		if firstInstPerc != nil {
			firstInstallmentPerc = firstInstPerc.(float64)
		}

		//Calculate the minimum amount payable using the firstinstallment percentage and the fee amount
		minAmount := h.GetAmountByPercentage(firstInstallmentPerc, mfAmount)//minAmount := ((firstInstallmentPerc / 100) * mfAmount)

		feeRes := fees {
			FeeAmount: mfAmount,
			FeeID: sfId,
			FeeTitle: sfTitle,
			FeeType: sfType,
			InstallmentEnabled: enabledInstallment,
			Installments: int(numberOfInstallment),
			MinimumAmount: minAmount,
			FirstInstallmentPercentage: firstInstallmentPerc,
			MerchantFeeId: smerchFeesId,
			FeeBearer: merchfeeBearer,
		}

		feesSlice = append(feesSlice,feeRes)
	}
	ms := merchantFees {
		MerchantID: smId,
		MerchantCode: smCode,
		MerchantPhoto: smPhoto,
		MerchantTitle: smTitle,
		Fees: feesSlice,
	}
	bs,_:= json.Marshal(ms)
	res := h.Response {
		Status: "success",
		Message: "Fees fetched successfully",
		Data: bs,
	}
	return c.JSON(http.StatusOK, res)
}