package helper 

import (
	"crypto/sha512"
	//"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io"
	"strings"
)

//Hash will hash the provided input string (which is the none, apiKey and apiSecret) using sha512 and return the string
func Hash256(input string) string {
	h512 := sha512.New()
	io.WriteString(h512, input)
	h := h512.Sum(nil)
	hashedString := fmt.Sprintf("%x", h)
	return strings.ToUpper(hashedString)
}

func BcryptHashPassword(plain string) (string, error) {
	hashedBytes, err :=  bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "",err
	}
	return string(hashedBytes),nil
}

func BcryptValidatePassword(plainPassword, passwordHash string) bool {
	//fmt.Println("plain password is ", plainPassword, "password hash is ", passwordHash)
	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(plainPassword))
		if err != nil {
			switch err {
			case bcrypt.ErrMismatchedHashAndPassword:
				//fmt.Println("entered the ErrMismatchedHashAndPassword of compare hash and password")
				return false
			default:
				//fmt.Println("entering the default of compare hash and password with err as ", err)
				return false
			}
		}
	return true
}

