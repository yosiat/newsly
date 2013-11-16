package main

import (
	"bytes"
	"code.google.com/p/go-imap/go1/imap"
	"net/mail"
	"strings"

	"fmt"
	"io"
)

// TODO: Maybe refactor the headers to Map[string, string]
// TODO: Why we need "To" and "DeliveredTo"
type MailMessage struct {
	To              string
	DeliveredTo     string
	From            string
	ContentType     string
	Subject         string
	Body            io.Reader
	UnsubscribeList string
}

type Gmail struct {
	connection *imap.Client
	userEmail  string
}

// Connects to the given mail server (example - some.mailserver.com:993)
// And returns a gmail connection and error
func Connect(mailServer string, userEmail string, accessToken string) (gmail *Gmail, err error) {
	connection, err := imap.DialTLS(mailServer, nil)
	HandleError(err)

	// TODO: Extract XoAuth
	_, err = connection.Auth(XoAuth(UserEmail, AccessToken))
	HandleError(err)

	gmail = &Gmail{
		connection: connection,
		userEmail:  userEmail,
	}

	return gmail, err
}

// Closes the connection see imap.Client.Close
func (gmail *Gmail) Close() {
	gmail.connection.Close(true)
}

// Trys to parse the given address key from the header, if it fails
// it returns the address from the header
func GetAddress(headers mail.Header, key string) (address string) {
	addresses, _ := headers.AddressList(key)

	if len(addresses) > 0 {
		return addresses[0].Address
	}

	return headers.Get(key)
}

// Given a mail.Message and body returns a MailMessage
// TODO: put body in msg
func ParseMessage(msg *mail.Message) (mail MailMessage) {
	to := GetAddress(msg.Header, "To")
	delivered_to := GetAddress(msg.Header, "Delivered-To")
	from := GetAddress(msg.Header, "From")
	content_type := strings.Trim(msg.Header.Get("Content-Type"), " ")
	subject := msg.Header.Get("Subject")
	unnsubscribe_list := msg.Header.Get("List-Unsubscribe")

	return MailMessage{To: to, DeliveredTo: delivered_to, From: from, ContentType: content_type, Body: msg.Body, Subject: subject, UnsubscribeList: unnsubscribe_list}

}

// Get all mails from the given mailbox and puts it on the channel, when it finished the channel
// will get closed
func (gmail *Gmail) GetAllMail(mailbox string, messages chan<- MailMessage) {
	command, err := gmail.connection.Select(mailbox, true)
	HandleError(err)

	// Search for mail that sender to the sender, and their body contains the unsubscribe

	command, err = imap.Wait(gmail.connection.UIDSearch([]imap.Field{"TO", gmail.userEmail},
		[]imap.Field{"BODY", "unsubscribe"},
		[]imap.Field{"BODY", "Unsubscribe"},
	))
	HandleError(err)

	matchedUids := command.Data[0].SearchResults()

	if len(matchedUids) > 0 {

		set, _ := imap.NewSeqSet("")
		set.AddNum(matchedUids...)

		gmail.connection.Data = nil

		command, err = gmail.connection.UIDFetch(set, "RFC822.HEADER", "BODY[]")
		HandleError(err)

		var (
			response     *imap.Response
			message_info *imap.MessageInfo
			header       []byte
			body         []byte
			message      []byte
		)

		// While we can read, receive data from the connection
		for command.InProgress() {

			gmail.connection.Recv(-1)

			// parse each response and send the parsed mail in the chanel..
			for _, response = range command.Data {
				message_info = response.MessageInfo()

				// get the header and the body and concat them
				header = imap.AsBytes(message_info.Attrs["RFC822.HEADER"])
				body = imap.AsBytes(message_info.Attrs["BODY[]"])
				message = append(header, body...)

				if msg, _ := mail.ReadMessage(bytes.NewReader(message)); msg != nil {
					messages <- ParseMessage(msg)
				}

			}

			command.Data = nil

			// Process unilateral server data
			for _, response = range gmail.connection.Data {
				fmt.Println("Server data:", response)
			}
			gmail.connection.Data = nil
		}

	}

	close(messages)
}
