package main

import (
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

const (
	ACCESS_KEY_ID     = "ACCESS"
	SECRET_ACCESS_KEY = "SECRET"
	REGION            = "us-east-1"
)

var sess *session.Session

type Mail struct {
	From    string
	To      []string
	CC      []string
	BCC     []string
	Subject string
	Body    []byte
}

func (m *Mail) Validate() error {
	if m.From == "" {
		return errors.New("from field is required")
	}
	if len(m.To) == 0 {
		return errors.New("at least one recipient email address is required in To field")
	}
	if m.Subject == "" {
		return errors.New("subject field is required")
	}
	if len(m.Body) == 0 {
		return errors.New("body field is required")
	}
	return nil
}

func main() {
	var err error
	// create new AWS session
	sess, err = session.NewSession(&aws.Config{
		Region:      aws.String(REGION),
		Credentials: credentials.NewStaticCredentials(ACCESS_KEY_ID, SECRET_ACCESS_KEY, "")},
	)
	if err != nil {
		log.Println("Error occurred while creating aws session", err)
		return
	}
	mail := &Mail{
		From:    "hello@gmail.com",
		To:      []string{"<toEmails_1>", "<toEmails_2>"},
		CC:      []string{"<ccEmails_1>", "<ccEmails_2>"},   //optional
		BCC:     []string{"<bccEmails_1>", "<bccEmails_2>"}, //optional
		Subject: "This is AWS Email System",
		Body:    []byte(`This is body of email`),
	}

	//mail validation
	if err := mail.Validate(); err != nil {
		log.Println("Validation error:", err)
		return
	}

	SendEmailSES(mail)

}

// SendEmailSES sends email to specified email IDs
func SendEmailSES(mail *Mail) {

	// set to section
	var recipients []*string
	for _, r := range mail.To {
		recipients = append(recipients, &r)
	}

	// set cc section
	var ccRecipients []*string
	if len(mail.CC) > 0 {
		for _, r := range mail.CC {
			ccRecipients = append(ccRecipients, &r)
		}
	}

	// set bcc section
	var bccRecipients []*string
	if len(mail.BCC) > 0 {
		for _, r := range mail.BCC {
			bccRecipients = append(bccRecipients, &r)
		}
	}

	// create an SES session.
	svc := ses.New(sess)

	// Assemble the email.
	input := &ses.SendEmailInput{

		// Set destination emails
		Destination: &ses.Destination{
			ToAddresses:  recipients,
			CcAddresses:  ccRecipients,
			BccAddresses: bccRecipients,
		},

		// Set email message and subject
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(string(mail.Body)),
				},
			},

			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(mail.Subject),
			},
		},

		// send from email
		Source: aws.String(mail.From),
	}

	// Call AWS send email function which internally calls to SES API
	_, err := svc.SendEmail(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				log.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				log.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println("Error sending mail - ", err.Error())
		}
		return

	}
	log.Println("Email sent successfully to: ", mail.To)
}
