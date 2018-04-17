package rand

import(
	"crypto/rand"
	"encoding/base64"
)

const TokenBytes = 32

func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

//String will generate a byte slice of size nBytes and then return 
//a string that is base64 URL encoded version of that byte slice
func String(nBytes int) (string,error) {
	b,err := Bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

//ReturnToken is a helper function designed to generate tokens of a predetermined byte size
func ReturnToken() (string, error) {
	return String(TokenBytes)
}

//RandStr is used to get the secret to be used for qrcoding generation
func RandStr(strSize int, randType string) string {
	var dictionary string

	if randType == "alphanum" {
			dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "alpha" {
			dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "number" {
			dictionary = "0123456789"
	}

	var bytes = make([]byte, strSize)
	rand.Read(bytes)
	for k, v := range bytes {
			bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}



