package usecases

import (
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"

	"github.com/moomaideng/eventory/internal/events"
)

const (
	registrationTemplateKey        = "registration-submitted-v1"
	registrationApprovedTemplateKey = "registration-approved-v1"
	registrationRejectedTemplateKey = "registration-rejected-v1"
	paymentConfirmedTemplateKey    = "payment-confirmed-v1"
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
		RegistrationID: event.Data.RegistrationID.String(), RegistrationMode: string(event.Data.RegistrationMode),
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
		DisplayName   string
		TeamName      string
		TournamentName string
		TournamentURL  string
	}{
		DisplayName:   event.Data.Recipient.DisplayName,
		TeamName:      event.Data.TeamName,
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
		DisplayName   string
		TournamentName string
		Reason         string
		TournamentURL  string
	}{
		DisplayName:   event.Data.Recipient.DisplayName,
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

Your payment of {{.Amount}} THB for {{.Description}} has been confirmed.

Eventory
`))

var paymentConfirmedHTMLTemplate = htmltemplate.Must(htmltemplate.New("payment-html").Parse(`<!doctype html>
<html><body style="font-family:Arial,sans-serif;color:#1f2937;line-height:1.5">
<h1 style="font-size:22px">Payment Confirmed</h1>
<p>Hello {{.DisplayName}},</p>
<p>Your payment of <strong>{{.Amount}} THB</strong> for <strong>{{.Description}}</strong> has been confirmed.</p>
<p>Eventory</p>
</body></html>`))

func RenderPaymentConfirmedEmail(event events.PaymentConfirmed, _ string) (EmailMessage, error) {
	amountStr := fmt.Sprintf("%.2f", float64(event.Data.AmountMinor)/100.0)
	data := struct {
		DisplayName string
		Amount      string
		Description string
	}{
		DisplayName: event.Data.Recipient.DisplayName,
		Amount:      amountStr,
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
