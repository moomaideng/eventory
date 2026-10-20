package usecases

import (
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"

	"github.com/moomaideng/eventory/internal/events"
)

const registrationTemplateKey = "registration-submitted-v1"

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
