package paystack

import (
	//"fmt"
)

type tokenizeCardResponse struct {
	Status 				string
	Message 			string 
	AuthorizationCode	string	
}

type paystackInitializeTransactionResponse struct {
	Data struct {
		AccessCode       string `json:"access_code"`
		AuthorizationURL string `json:"authorization_url"`
		Reference        string `json:"reference"`
	} `json:"data"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

type paystackChargeResponse struct {
	Data struct {
		DisplayText string `json:"display_text,omitempty"`
		Message string `json:"message"`
		Reference   string `json:"reference"`
		Status      string `json:"status"`
	} `json:"data"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

type paystackSuccessChargeResponse struct {
	Data struct {
		Amount        int `json:"amount"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bank              string `json:"bank"`
			Bin               string `json:"bin"`
			Brand             string `json:"brand"`
			CardType          string `json:"card_type"`
			Channel           string `json:"channel"`
			CountryCode       string `json:"country_code"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Last4             string `json:"last4"`
			Reusable          bool   `json:"reusable"`
			Signature         string `json:"signature"`
		} `json:"authorization"`
		Channel  string `json:"channel"`
		Currency string `json:"currency"`
		Customer struct {
			CustomerCode string      `json:"customer_code"`
			Email        string      `json:"email"`
			FirstName    interface{} `json:"first_name"`
			ID           int         `json:"id"`
			LastName     interface{} `json:"last_name"`
			Metadata     interface{} `json:"metadata"`
			Phone        interface{} `json:"phone"`
			RiskAction   string      `json:"risk_action"`
		} `json:"customer"`
		Domain          string      `json:"domain"`
		Fees            int         `json:"fees"`
		FeesSplit struct {
			Integration int `json:"integration"`
			Params      struct {
				Bearer            string `json:"bearer"`
				PercentageCharge  string `json:"percentage_charge"`
				TransactionCharge string `json:"transaction_charge"`
			} `json:"params"`
			Paystack   int `json:"paystack"`
			Subaccount int `json:"subaccount"`
		} `json:"fees_split"`
		GatewayResponse string      `json:"gateway_response"`
		IPAddress       string      `json:"ip_address"`
		Log             interface{} `json:"log"`
		Message         interface{} `json:"message"`
		Metadata        struct {
			CustomFields []struct {
				DisplayName  string `json:"display_name"`
				Value        string `json:"value"`
				VariableName string `json:"variable_name"`
			} `json:"custom_fields"`
		} `json:"metadata,omitempty"`
		Plan            interface{} `json:"plan"`
		Reference       string      `json:"reference"`
		Status          string      `json:"status"`
		TransactionDate string      `json:"transaction_date"`
	} `json:"data"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
}


type paystackTokenizeCardResponse struct {
	Data struct {
		AuthorizationCode string `json:"authorization_code"`
		Bank              string `json:"bank"`
		Bin               string `json:"bin"`
		Brand             string `json:"brand"`
		CardType          string `json:"card_type"`
		Channel           string `json:"channel"`
		CountryCode       string `json:"country_code"`
		Customer          struct {
			CustomerCode string      `json:"customer_code"`
			Email        string      `json:"email"`
			FirstName    interface{} `json:"first_name"`
			ID           int         `json:"id"`
			LastName     interface{} `json:"last_name"`
			Metadata     interface{} `json:"metadata"`
			Phone        interface{} `json:"phone"`
			RiskAction   string      `json:"risk_action"`
		} `json:"customer"`
		ExpMonth  string `json:"exp_month"`
		ExpYear   string `json:"exp_year"`
		Last4     string `json:"last4"`
		Reusable  bool   `json:"reusable"`
		Signature string `json:"signature"`
	} `json:"data"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
}


type createSubAccountResponse struct {
	Data struct {
		AccountNumber       string      `json:"account_number"`
		Active              bool        `json:"active"`
		BusinessName        string      `json:"business_name"`
		CreatedAt           string      `json:"createdAt"`
		Description         interface{} `json:"description"`
		Domain              string      `json:"domain"`
		ID                  int         `json:"id"`
		Integration         int         `json:"integration"`
		IsVerified          bool        `json:"is_verified"`
		Metadata            interface{} `json:"metadata"`
		Migrate             bool        `json:"migrate"`
		PercentageCharge    float64     `json:"percentage_charge"`
		PrimaryContactEmail interface{} `json:"primary_contact_email"`
		PrimaryContactName  interface{} `json:"primary_contact_name"`
		PrimaryContactPhone interface{} `json:"primary_contact_phone"`
		SettlementBank      string      `json:"settlement_bank"`
		SettlementSchedule  string      `json:"settlement_schedule"`
		SubaccountCode      string      `json:"subaccount_code"`
		UpdatedAt           string      `json:"updatedAt"`
	} `json:"data"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

type verifyTransactionResponse struct {
	Data struct {
		Amount        int `json:"amount"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			Bank              string `json:"bank"`
			Bin               string `json:"bin"`
			Brand             string `json:"brand"`
			CardType          string `json:"card_type"`
			Channel           string `json:"channel"`
			CountryCode       string `json:"country_code"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Last4             string `json:"last4"`
			Reusable          bool   `json:"reusable"`
			Signature         string `json:"signature"`
		} `json:"authorization"`
		Channel   string `json:"channel"`
		CreatedAt string `json:"createdAt"`
		Created_At string `json:"created_at"`
		Currency  string `json:"currency"`
		Customer  struct {
			CustomerCode string      `json:"customer_code"`
			Email        string      `json:"email"`
			FirstName    interface{} `json:"first_name"`
			ID           int         `json:"id"`
			LastName     interface{} `json:"last_name"`
			Metadata     interface{} `json:"metadata"`
			Phone        interface{} `json:"phone"`
			RiskAction   string      `json:"risk_action"`
		} `json:"customer"`
		Domain          string      `json:"domain"`
		Fees            interface{} `json:"fees"`
		FeesSplit struct {
			Integration int `json:"integration"`
			Params      struct {
				Bearer            string `json:"bearer"`
				PercentageCharge  string `json:"percentage_charge"`
				TransactionCharge string `json:"transaction_charge"`
			} `json:"params"`
			Paystack   int `json:"paystack"`
			Subaccount int `json:"subaccount"`
		} `json:"fees_split"`
		GatewayResponse string      `json:"gateway_response"`
		ID              int         `json:"id"`
		IPAddress       string      `json:"ip_address"`
		Log             struct {
			Attempts       int    `json:"attempts"`
			Authentication string `json:"authentication"`
			Channel        string `json:"channel"`
			Errors         int    `json:"errors"`
			History        []struct {
				Message string `json:"message"`
				Time    int    `json:"time"`
				Type    string `json:"type"`
			} `json:"history"`
			Input     []interface{} `json:"input"`
			Mobile    bool          `json:"mobile"`
			Success   bool          `json:"success"`
			TimeSpent int           `json:"time_spent"`
		} `json:"log"`
		Message  interface{} `json:"message"`
		Metadata struct {
			CustomFields []struct {
				DisplayName string `json:"display_name"`
				Value       string `json:"value"`
			} `json:"custom_fields"`
		} `json:"metadata"`
		PaidAt          string      `json:"paidAt"`
		Paid_At          string      `json:"paid_at"`
		Plan            interface{} `json:"plan"`
		PlanObject      struct{}    `json:"plan_object"`
		Reference       string      `json:"reference"`
		Status          string      `json:"status"`
		TransactionDate string      `json:"transaction_date"`
	} `json:"data"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
}