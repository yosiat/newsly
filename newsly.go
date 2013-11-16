package main

import (
	"encoding/json"

	"os"
	"regexp"
	"strings"
)

var (
	UserEmail   = ""
	AccessToken = ""

	GmailServer = "imap.gmail.com:993"
)

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	gmail, err := Connect(GmailServer, UserEmail, AccessToken)
	HandleError(err)

	defer gmail.Close()

	messages := make(chan MailMessage)
	go gmail.GetAllMail("INBOX", messages)

	urlRegexp := regexp.MustCompile("<https?://.*>")

	var fromToRemoveSubscriptionUrl map[string]string = make(map[string]string)

	for mail := range messages {
		_, exists := fromToRemoveSubscriptionUrl[mail.From]

		if !exists && mail.UnsubscribeList != "" {
			if url := urlRegexp.FindString(mail.UnsubscribeList); url != "" {
				fromToRemoveSubscriptionUrl[mail.From] = strings.Trim(url, "<> ")
			}
		}

	}

	b, err := json.Marshal(fromToRemoveSubscriptionUrl)
	HandleError(err)

	os.Stdout.Write(b)

}
