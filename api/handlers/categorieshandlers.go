package handlers

import (
	"github.com/labstack/echo"
	"net/http"
	"fmt"
	"encoding/json"
	
	s "strings"
	h "github.com/kenmobility/feezbot/helper"
)

type AllCategories struct {
	Categories []CategoriesInfo `json:"categories"`
}

type CategoriesInfo struct {
	CategoryName        string `json:"category_Name"`
	CategoryDescription string `json:"category_description"`
	CategoryID          string `json:"category_id"`
}

func ShowCategories(c echo.Context) error {
	con, err := h.OpenConnection()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error in connecting to database")
	}
	defer con.Close()

	catSlice := make([]CategoriesInfo,0)

	var catName, catDesc, catId interface{}
	var sCatName, sCatDesc, sCatId string

	q := `SELECT "Id","Category","Description" FROM "_categories"`
	catRows,err := con.Db.Query(q)
	defer catRows.Close()
	if err != nil{
		if s.Contains(fmt.Sprintf("%v", err), "no records") == true {
			res := h.Response {
				Status: "error",
				Message:"No record found!",
			}
			return c.JSON(http.StatusOK, res)	
		}else{
			fmt.Println("categorieshandlers.go::ShowCategories()::ShowCategories sql query Failed due to:", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	for catRows.Next() {
		err = catRows.Scan(&catId,&catName,&catDesc)
		if err != nil {
			fmt.Println("categorieshandlers.go::ShowCategories()::scanning query Failed due to:", err)
		}
		if catId != nil {
			sCatId = catId.(string)
		}
		if catName != nil {
			sCatName = catName.(string)
		}
		if catDesc != nil {
			sCatDesc = catDesc.(string)
		}

		c := CategoriesInfo {
			CategoryID : sCatId,
			CategoryName : sCatName,
			CategoryDescription : sCatDesc,
		}
		catSlice = append(catSlice, c)
	}
	ac := AllCategories {
		Categories : catSlice,
	}
	bs,_:= json.Marshal(ac)
	res := h.Response {
		Status: "success",
		Message: "Categories fetched successfully",
		Data: bs,
	}
	return c.JSON(http.StatusOK,res)
}