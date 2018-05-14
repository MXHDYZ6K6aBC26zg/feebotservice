package email 

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

func SendMail(recipient,message string) error {
	subject := "test Reset Passwork Confirmation Code"
	m := gomail.NewMessage()
	m.SetHeader("From", "ipaytsa1@gmail.com")
	//m.SetHeader("To", "nduson2k@gmail.com")
	m.SetHeader("To", recipient)
	//m.SetHeader("Subject", "Testing Mail From Golang!")
	m.SetHeader("Subject", subject)
	//m.SetBody("text/html", "Hello <b>Bob</b>!")
	m.SetBody("text/plain", message)

	// Send the email to Bob
	d := gomail.NewDialer("smtp.gmail.com", 587, "ipaytsa1@gmail.com", "ikechukwu")
	if err := d.DialAndSend(m); err != nil {
		fmt.Println("email.go:sendmail: Sending mail to", recipient, " Faild Due to", err)
		return err
	}
	return nil
}