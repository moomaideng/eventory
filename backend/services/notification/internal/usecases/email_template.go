package usecases

import (
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"

	"github.com/moomaideng/eventory/internal/events"
	"github.com/moomaideng/eventory/internal/money"
)

const (
	registrationTemplateKey         = "registration-submitted-v1"
	registrationApprovedTemplateKey = "registration-approved-v1"
	registrationRejectedTemplateKey = "registration-rejected-v1"
	paymentConfirmedTemplateKey     = "payment-confirmed-v1"
)

type EmailMessage struct {
	To        string
	Subject   string
	TextBody  string
	HTMLBody  string
	MessageID string
}

type templateData struct {
	DisplayName      string
	TournamentName   string
	RegistrationID   string
	RegistrationMode string
	TeamLine         string
	TournamentURL    string
}

var registrationTextTemplate = texttemplate.Must(texttemplate.New("registration-text").Parse(`Hello {{.DisplayName}},

We received your registration for {{.TournamentName}}.
Registration ID: {{.RegistrationID}}
Registration type: {{.RegistrationMode}}{{.TeamLine}}

Your registration is waiting for organizer review. This email does not mean that the registration has been accepted yet.

View tournament: {{.TournamentURL}}

Eventory
`))

var registrationHTMLTemplate = htmltemplate.Must(htmltemplate.New("registration-html").Parse(`<!doctype html>
<html><body style="font-family:Arial,sans-serif;color:#1f2937;line-height:1.5">
<h1 style="font-size:22px">Registration received</h1>
<p>Hello {{.DisplayName}},</p>
<p>We received your registration for <strong>{{.TournamentName}}</strong>.</p>
<ul><li>Registration ID: {{.RegistrationID}}</li><li>Registration type: {{.RegistrationMode}}{{.TeamLine}}</li></ul>
<p>Your registration is waiting for organizer review. This email does not mean that the registration has been accepted yet.</p>
<p><a href="{{.TournamentURL}}">View tournament</a></p>
<p>Eventory</p>
</body></html>`))

func RenderRegistrationEmail(event events.RegistrationSubmitted, publicWebURL string) (EmailMessage, error) {
	teamLine := ""
	if event.Data.RegistrationMode == events.RegistrationModeTeam && event.Data.TeamName != nil {
		teamLine = " - Team: " + *event.Data.TeamName
	}
	data := templateData{
		DisplayName: event.Data.Recipient.DisplayName, TournamentName: event.Data.TournamentName,
		RegistrationID: event.Data.RegistrationID, RegistrationMode: string(event.Data.RegistrationMode),
		TeamLine:      teamLine,
		TournamentURL: strings.TrimRight(publicWebURL, "/") + "/tournaments/" + event.Data.TournamentID.String(),
	}
	var textBody, htmlBody strings.Builder
	if err := registrationTextTemplate.Execute(&textBody, data); err != nil {
		return EmailMessage{}, err
	}
	if err := registrationHTMLTemplate.Execute(&htmlBody, data); err != nil {
		return EmailMessage{}, err
	}
	return EmailMessage{
		To:       event.Data.Recipient.Email,
		Subject:  headerValue("Registration received - " + event.Data.TournamentName),
		TextBody: textBody.String(), HTMLBody: htmlBody.String(),
		MessageID: fmt.Sprintf("<%s-%s@eventory>", event.ID, event.Data.Recipient.AccountID),
	}, nil
}

func headerValue(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r", " "), "\n", " ")
}

var approvedTextTemplate = texttemplate.Must(texttemplate.New("approved-text").Parse(`Hello {{.DisplayName}},

Great news! Your team {{.TeamName}} has been approved for {{.TournamentName}}.

View tournament: {{.TournamentURL}}

Eventory
`))

var approvedHTMLTemplate = htmltemplate.Must(htmltemplate.New("approved-html").Parse(`<!doctype html>
<html><body style="font-family:Arial,sans-serif;color:#1f2937;line-height:1.5">
<h1 style="font-size:22px">Registration Approved</h1>
<p>Hello {{.DisplayName}},</p>
<p>Great news! Your team <strong>{{.TeamName}}</strong> has been approved for <strong>{{.TournamentName}}</strong>.</p>
<p><a href="{{.TournamentURL}}">View tournament</a></p>
<p>Eventory</p>
</body></html>`))

func RenderRegistrationApprovedEmail(event events.RegistrationReviewed, publicWebURL string) (EmailMessage, error) {
	data := struct {
		DisplayName    string
		TeamName       string
		TournamentName string
		TournamentURL  string
	}{
		DisplayName:    event.Data.Recipient.DisplayName,
		TeamName:       event.Data.TeamName,
		TournamentName: event.Data.TournamentName,
		TournamentURL:  strings.TrimRight(publicWebURL, "/") + "/tournaments/" + event.Data.TournamentID.String(),
	}
	var textBody, htmlBody strings.Builder
	if err := approvedTextTemplate.Execute(&textBody, data); err != nil {
		return EmailMessage{}, err
	}
	if err := approvedHTMLTemplate.Execute(&htmlBody, data); err != nil {
		return EmailMessage{}, err
	}
	return EmailMessage{
		To:        event.Data.Recipient.Email,
		Subject:   headerValue("Registration Approved - " + event.Data.TournamentName),
		TextBody:  textBody.String(),
		HTMLBody:  htmlBody.String(),
		MessageID: fmt.Sprintf("<%s-%s@eventory>", event.ID, event.Data.Recipient.AccountID),
	}, nil
}

var rejectedTextTemplate = texttemplate.Must(texttemplate.New("rejected-text").Parse(`Hello {{.DisplayName}},

Your registration application for {{.TournamentName}} was not accepted.
Reason: {{.Reason}}

View tournament: {{.TournamentURL}}

Eventory
`))

var rejectedHTMLTemplate = htmltemplate.Must(htmltemplate.New("rejected-html").Parse(`<!doctype html>
<html><body style="font-family:Arial,sans-serif;color:#1f2937;line-height:1.5">
<h1 style="font-size:22px">Registration Rejected</h1>
<p>Hello {{.DisplayName}},</p>
<p>Your registration application for <strong>{{.TournamentName}}</strong> was not accepted.</p>
<p>Reason: {{.Reason}}</p>
<p><a href="{{.TournamentURL}}">View tournament</a></p>
<p>Eventory</p>
</body></html>`))

func RenderRegistrationRejectedEmail(event events.RegistrationReviewed, publicWebURL string) (EmailMessage, error) {
	reason := event.Data.Reason
	if reason == "" {
		reason = "No specific reason provided."
	}
	data := struct {
		DisplayName    string
		TournamentName string
		Reason         string
		TournamentURL  string
	}{
		DisplayName:    event.Data.Recipient.DisplayName,
		TournamentName: event.Data.TournamentName,
		Reason:         reason,
		TournamentURL:  strings.TrimRight(publicWebURL, "/") + "/tournaments/" + event.Data.TournamentID.String(),
	}
	var textBody, htmlBody strings.Builder
	if err := rejectedTextTemplate.Execute(&textBody, data); err != nil {
		return EmailMessage{}, err
	}
	if err := rejectedHTMLTemplate.Execute(&htmlBody, data); err != nil {
		return EmailMessage{}, err
	}
	return EmailMessage{
		To:        event.Data.Recipient.Email,
		Subject:   headerValue("Registration Rejected - " + event.Data.TournamentName),
		TextBody:  textBody.String(),
		HTMLBody:  htmlBody.String(),
		MessageID: fmt.Sprintf("<%s-%s@eventory>", event.ID, event.Data.Recipient.AccountID),
	}, nil
}

var paymentConfirmedTextTemplate = texttemplate.Must(texttemplate.New("payment-text").Parse(`Hello {{.DisplayName}},

Your payment of {{.Amount}} for {{.Description}} has been confirmed.

Eventory
`))

var paymentConfirmedHTMLTemplate = htmltemplate.Must(htmltemplate.New("payment-html").Parse(`<!doctype html>
<html><body style="font-family:Arial,sans-serif;color:#1f2937;line-height:1.5">
<h1 style="font-size:22px">Payment Confirmed</h1>
<p>Hello {{.DisplayName}},</p>
<p>Your payment of <strong>{{.Amount}}</strong> for <strong>{{.Description}}</strong> has been confirmed.</p>
<p>Eventory</p>
</body></html>`))

func RenderPaymentConfirmedEmail(event events.PaymentConfirmed, _ string) (EmailMessage, error) {
	data := struct {
		DisplayName string
		Amount      string
		Description string
	}{
		DisplayName: event.Data.Recipient.DisplayName,
		Amount:      money.Format(event.Data.AmountMinor, event.Data.Currency),
		Description: event.Data.Description,
	}
	var textBody, htmlBody strings.Builder
	if err := paymentConfirmedTextTemplate.Execute(&textBody, data); err != nil {
		return EmailMessage{}, err
	}
	if err := paymentConfirmedHTMLTemplate.Execute(&htmlBody, data); err != nil {
		return EmailMessage{}, err
	}
	return EmailMessage{
		To:        event.Data.Recipient.Email,
		Subject:   headerValue("Payment Confirmed - " + event.Data.Description),
		TextBody:  textBody.String(),
		HTMLBody:  htmlBody.String(),
		MessageID: fmt.Sprintf("<%s-%s@eventory>", event.ID, event.Data.Recipient.AccountID),
	}, nil
}

const announcementTemplateKey = "announcement-created-v1"

var announcementTextTemplate = texttemplate.Must(texttemplate.New("announcement-text").Parse(`Hello {{.DisplayName}},

{{.AuthorName}} posted a new announcement for {{.TournamentName}}.

{{.Title}}

{{.Body}}

View tournament: {{.TournamentURL}}

Eventory
`))

var announcementHTMLTemplate = htmltemplate.Must(htmltemplate.New("announcement-html").Parse(`<!doctype html>
<html><body style="font-family:Arial,sans-serif;color:#1f2937;line-height:1.5">
<h1 style="font-size:22px">{{.Title}}</h1>
<p>Hello {{.DisplayName}},</p>
<p><strong>{{.AuthorName}}</strong> posted a new announcement for <strong>{{.TournamentName}}</strong>.</p>
<p style="white-space:pre-wrap">{{.Body}}</p>
<p><a href="{{.TournamentURL}}">View tournament</a></p>
<p>Eventory</p>
</body></html>`))

func RenderAnnouncementEmail(event events.AnnouncementCreated, publicWebURL string) (EmailMessage, error) {
	displayName := event.Data.Recipient.DisplayName
	if displayName == "" {
		displayName = "there"
	}
	authorName := event.Data.AuthorName
	if authorName == "" {
		authorName = "The organizer"
	}
	data := struct {
		DisplayName    string
		AuthorName     string
		TournamentName string
		Title          string
		Body           string
		TournamentURL  string
	}{
		DisplayName:    displayName,
		AuthorName:     authorName,
		TournamentName: event.Data.TournamentName,
		Title:          event.Data.Title,
		Body:           event.Data.Body,
		TournamentURL:  strings.TrimRight(publicWebURL, "/") + "/tournaments/" + event.Data.TournamentID.String(),
	}
	var textBody, htmlBody strings.Builder
	if err := announcementTextTemplate.Execute(&textBody, data); err != nil {
		return EmailMessage{}, err
	}
	if err := announcementHTMLTemplate.Execute(&htmlBody, data); err != nil {
		return EmailMessage{}, err
	}
	return EmailMessage{
		To:        event.Data.Recipient.Email,
		Subject:   headerValue("[" + event.Data.TournamentName + "] " + event.Data.Title),
		TextBody:  textBody.String(),
		HTMLBody:  htmlBody.String(),
		MessageID: fmt.Sprintf("<%s-%s@eventory>", event.ID, event.Data.Recipient.AccountID),
	}, nil
}
