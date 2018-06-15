package email 

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	//"strings"
)

type mail struct {
	senderId 		string
	senderPassword  string
	toId     		string
	subject  		string
	body     		string
}

type smtpServer struct {
	host string
	port string
}

func (s *smtpServer) ServerName() string {
	return s.host + ":" + s.port
}

func MailConfig(sender,senderPassword,receipient, subject, body string) *mail {
	return &mail {
		senderId: sender,
		senderPassword: senderPassword,
		toId: receipient,
		subject: subject,
		body: body,
	}
}

 func buildMessage(mail *mail) string {
	message := ""
	message += fmt.Sprintf("From: %s\r\n", mail.senderId)
	message += fmt.Sprintf("To: %s\r\n", mail.toId)
	message += fmt.Sprintf("Subject: %s\r\n", mail.subject)
	message += "\r\n" + mail.body

	return message
}

func SendMail(mail *mail) error {
	messageBody := buildMessage(mail)

	smtpServer := smtpServer{host: "smtp.gmail.com", port: "465"}

	//log.Println("server host is :",smtpServer.host)
	//build an auth
	auth := smtp.PlainAuth("", mail.senderId, mail.senderPassword, smtpServer.host)

	// Gmail will reject connection if it's not secure
	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer.host,
	}

	conn, err := tls.Dial("tcp", smtpServer.ServerName(), tlsconfig)
	if err != nil {
		log.Println("email.go::SendMail():: tls.Dial error due to: ",err)
	}

	client, err := smtp.NewClient(conn, smtpServer.host)
	if err != nil {
		log.Println("email.go::SendMail():: creating smtp new client failed due to :",err)
		return err//log.Panic(err)
	}

	// step 1: Use Auth
	if err = client.Auth(auth); err != nil {
		log.Println("email.go::SendMail()::  client.Auth error due to: ",err)
	}

	// step 2: add all from and to
	if err = client.Mail(mail.senderId); err != nil {
		log.Println("email.go::SendMail()::  client.Mail error due to: ",err)
	}
	/* for _, k := range mail.toId {
		if err = client.Rcpt(k); err != nil {
			log.Panic(err)
		}
	} */
	if err = client.Rcpt(mail.toId); err != nil {
		log.Println("creating a client reciepient failed due to ", err)
		return err
	}
	// Data
	w, err := client.Data()
	if err != nil {
		log.Println("creating the mail data failed due to ", err)
		return err
	}

	_, err = w.Write([]byte(messageBody))
	if err != nil {
		log.Println("writing the built message body to byte array failed du to ", err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Println("closing the written message body failed due to :", err)
		return err
	}

	client.Quit()

	//log.Println("Mail sent successfully")
	return nil
}

/* import (
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
} */





