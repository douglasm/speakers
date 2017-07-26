package users

import (
	"fmt"
	"gopkg.in/mailgun/mailgun-go.v1"
	"log"
)

func sendActivate(email, code string) {
	theLink := getBaseAddr()

	messageText := fmt.Sprintf("Please go to %s/activate/%s", theLink, code)
	messageText += " to activate your profile, or go to "
	messageText += fmt.Sprintf("%s/activate and enter the code %s.", theLink, code)

	sendAddress := fmt.Sprintf("noreply@%s", domain)
	fmt.Println(sendAddress)

	mg := mailgun.NewMailgun(mailGunDomain, mailKey, "")
	m := mg.NewMessage(
		sendAddress,
		fmt.Sprintf("Welcome to %s", siteName),
		messageText,
		email,
	)

	_, _, err := mg.Send(m)
	if err != nil {
		log.Println("Error: send activate", err)
	}
	return
}

func sendReset(email, code string) {
	theLink := getBaseAddr()
	messageText := fmt.Sprintf("Please go to %s/newpassword/%s", theLink, code)
	messageText += " to reset your password, or go to "
	messageText += fmt.Sprintf("%s/newpassword and enter the code %s.", theLink, code)

	sendAddress := fmt.Sprintf("passwordreset@%s", domain)

	mg := mailgun.NewMailgun(mailGunDomain, mailKey, "")
	m := mg.NewMessage(
		sendAddress,
		fmt.Sprintf("%s password reset", siteName),
		messageText,
		email,
	)

	_, _, err := mg.Send(m)
	if err != nil {
		log.Println("Error: send activate", err)
	}
	return
}

func sendMessage(email, theMessage, sender string) {
	sendAddress := fmt.Sprintf("noreply@%s", domain)

	mg := mailgun.NewMailgun(mailGunDomain, mailKey, "")
	m := mg.NewMessage(
		sendAddress,
		fmt.Sprintf("%s message from %s", siteName, sender),
		theMessage,
		email,
	)
	_, _, err := mg.Send(m)
	if err != nil {
		log.Println("Error: sendMessage", err)
	}
	return
}

func getBaseAddr() string {
	theLink := "http"
	if isSecure {
		theLink += "s"
	}
	theLink += "://" + domain
	return theLink
}
