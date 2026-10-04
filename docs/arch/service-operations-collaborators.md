# Service–Operations–Collaborators

Matching diagram: [architecture-diagram.md](./architecture-diagram.md).

| Service | Operations | Collaborators |
| :---- | :---- | :---- |
| Account Service | • CreateOrganizerProfile()<br>• CreateSponsorProfile()<br>• GetAccount()<br>• GetOrganizerProfile()<br>• GetSponsorProfile() | Supabase Auth<br>PostgreSQL (account) |
| Tournament Service | • CreateTournamentDraft()<br>• ConfigureCrowdfunding()<br>• ConfigureRegistrationForm()<br>• UpdateDraftSettings()<br>• OpenTournament()<br>• GetTournament()<br>• GetFundingProgress()<br>• CreateTeamLobby()<br>• JoinTeamLobby()<br>• LockTeamRoster()<br>• SubmitRegistration()<br>• RecordSponsorship()<br>• RecordMatchResult() | Account Service<br>• GetAccount()<br>Payment Service<br>• CreateCheckout()<br>• GetPaymentStatus()<br>Object Storage<br>Message Broker<br>• publish registration.submitted<br>• publish match_result.recorded<br>• publish sponsorship.confirmed<br>PostgreSQL (tournament)<br>MongoDB (registration forms) |
| Payment Service | • CreateCheckout()<br>• GetPaymentStatus()<br>• HandleGatewayCallback()<br>• RecordLedgerEntry() | Payment Gateway<br>Message Broker<br>• publish payment.succeeded<br>PostgreSQL (payment) |
| Notification Service | • SendEmailNotification()<br>• PublishInAppNotification() | Message Broker<br>• consume payment.succeeded<br>• consume registration.submitted<br>• consume match_result.recorded<br>• consume sponsorship.confirmed<br>Email Provider<br>PostgreSQL (notification) |
