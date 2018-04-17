package helper 

import (
	"crypto/sha256"
	"encoding/base64"
	//"fmt"
	"golang.org/x/crypto/bcrypt"
)

//Hash will hash the provided input string using sha256 and return the encoded base64 string
func Hash256(input string) string {
	h := sha256.New()
	h.Write([]byte(input))
	b := h.Sum(nil)
	return base64.URLEncoding.EncodeToString(b)
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

