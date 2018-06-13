package paystack

import (
	"io/ioutil"
	"net/http"
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
	"strings"
	"encoding/base64"
	"time"
	"crypto/sha512"
	"io"
)

const(
	contentType = "application/json"
	TEST_SECRET_KEY =  "sk_test_3136a7307d138cc584153b9151f6278938badd19"
	TEST_PUBLIC_KEY = "pk_test_8b053df8eabd3a65e6de5aa2f46e08a78c7df152"
)

var auth = fmt.Sprintf("Bearer %s", TEST_SECRET_KEY)

// sendHttpRequest is use to send Get http request
func sendHTTPRequest(url, requestMethod string) (string, int) {
	req, _ := http.NewRequest(requestMethod, url, nil)

	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", contentType)

	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return err.Error(), -1
	}
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return string(body), resp.StatusCode
}

// sendHttpPostRequest send post request with post data
func sendHTTPPostRequest(url, requestMethod, postData string) (string,int) {
	req, err := http.NewRequest(requestMethod, url, strings.NewReader(postData))
	if err != nil {
		return err.Error(),-1
	}
	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", contentType)
	//fmt.Println("the req is ", req)
	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		//fmt.Println("http post", resp, "error:", err)
		return err.Error(), -1
	}
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return string(body), resp.StatusCode
}

func generateTimestamp() int64 {

	return time.Now().Unix()
}

func generateNonce() string {
	uid, _ := uuid.NewV4()
	nonce := fmt.Sprintf("%v", uid)
	nonce = strings.Replace(nonce, "-", "", -1)

	return nonce
}

func generateRequestReference(requestReferencePrefix string) string {
	time := time.Now().Unix()
	//requestReferencePrefix := "1609" // live
	//requestReferencePrefix := "1456" // test
	return fmt.Sprintf("%v%v", requestReferencePrefix, time)
}

func percentEncode(encodeMe string) string {
	if encodeMe == "" {
		return ""
	}
	encodeMe = strings.Replace(encodeMe, "%", "%25", -1)
	encodeMe = strings.Replace(encodeMe, "!", "%21", -1)
	encodeMe = strings.Replace(encodeMe, "#", "%23", -1)
	encodeMe = strings.Replace(encodeMe, "$", "%24", -1)
	encodeMe = strings.Replace(encodeMe, "&", "%26", -1)
	encodeMe = strings.Replace(encodeMe, "'", "%27", -1)
	encodeMe = strings.Replace(encodeMe, "(", "%28", -1)
	encodeMe = strings.Replace(encodeMe, ")", "%29", -1)
	encodeMe = strings.Replace(encodeMe, "*", "%2A", -1)
	encodeMe = strings.Replace(encodeMe, "+", "%2B", -1)
	encodeMe = strings.Replace(encodeMe, ",", "%2C", -1)
	encodeMe = strings.Replace(encodeMe, "/", "%2F", -1)
	encodeMe = strings.Replace(encodeMe, ":", "%3A", -1)
	encodeMe = strings.Replace(encodeMe, ";", "%3B", -1)
	encodeMe = strings.Replace(encodeMe, "=", "%3D", -1)
	encodeMe = strings.Replace(encodeMe, "?", "%3F", -1)
	encodeMe = strings.Replace(encodeMe, "@", "%40", -1)
	encodeMe = strings.Replace(encodeMe, "[", "%5B", -1)
	encodeMe = strings.Replace(encodeMe, "]", "%5D", -1)

	return encodeMe
}

func generateAuthHeader(clientid string) string {
	clientID := base64.StdEncoding.EncodeToString([]byte(clientid))
	auth := fmt.Sprintf("InterswitchAuth %v", clientID)
	return auth
}

func generateCypher(clientID, secretKey, urlEncoded, httpVerb, timestamp, nonce string) string {
	baseStringToBeSigned := httpVerb + "&" +
		urlEncoded + "&" +
		timestamp + "&" +
		nonce + "&" +
		clientID + "&" +
		secretKey

	return baseStringToBeSigned
}

func generateSignature(baseStringToBeSigned string) string {
	s := sha512.New()
	io.WriteString(s, baseStringToBeSigned)
	signatureR := s.Sum(nil)
	signature := base64.StdEncoding.EncodeToString(signatureR)
	return signature
}
