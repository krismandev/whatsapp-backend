package lib

import (
	"crypto/tls"
	"strings"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

const(
	smtpHost = "smtp.gmail.com"
	smtpPort = 587
)
//Email struct for email
type Email struct {
	SenderEmail		string
	SenderPassword 	string
}

//SendEmail for send email
func (email Email) SendEmail(to,cc,bcc string, subject, body string) error{
	mailer := gomail.NewMessage()
	toList:=strings.Split(to,",")
    ccList:=strings.Split(cc,",")
    bccList:=strings.Split(bcc,",")
    mailer.SetHeader("From", email.SenderEmail)
    mailer.SetHeader("To", toList...)
    if len(cc)>0{
        mailer.SetHeader("Cc", ccList...)
    }
    if len(bcc)>0{
        mailer.SetHeader("Bcc", bccList...)
    }
    //mailer.SetAddressHeader("Cc", "tralalala@gmail.com", "Tra Lala La")
    mailer.SetHeader("Subject", subject) 
    mailer.SetBody("text/html", body)

    dialer := gomail.NewDialer(
        smtpHost,
        smtpPort,
        email.SenderEmail,
        email.SenderPassword,
    )
	dialer.TLSConfig=&tls.Config{InsecureSkipVerify: true}

    err := dialer.DialAndSend(mailer)
    if err != nil {
		log.Errorf("Error Send Email : &s",err.Error())
		return err
    }
	return nil
}

//SendEmailWithAtt for send email with attachment
func (email Email) SendEmailWithAtt(to,cc,bcc string, subject, body,attFilePath string ) error {
	mailer := gomail.NewMessage()
    toList:=strings.Split(to,",")
    ccList:=strings.Split(cc,",")
    bccList:=strings.Split(bcc,",")
    mailer.SetHeader("From", email.SenderEmail)
    mailer.SetHeader("To", toList...)
    if len(cc)>0{
        mailer.SetHeader("Cc", ccList...)
    }
    if len(bcc)>0{
        mailer.SetHeader("Bcc", bccList...)
    }

    mailer.SetHeader("Subject", subject) 
    mailer.SetBody("text/html", body)
    //mailer.Attach("./sample.png")
	mailer.Attach(attFilePath)
    dialer := gomail.NewDialer(
        smtpHost,
		smtpPort,
        email.SenderEmail,
        email.SenderPassword,
    )
	dialer.TLSConfig=&tls.Config{InsecureSkipVerify: true}

	err := dialer.DialAndSend(mailer)
    if err != nil {
		log.Errorf("Error Send Email : &s",err.Error())
		return err
    }
	return nil
}