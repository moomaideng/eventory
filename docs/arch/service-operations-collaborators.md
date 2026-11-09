# Service–Operations–Collaborators

Matching diagram: [architecture-diagram.md](./architecture-diagram.md).

| Service | Operations | Collaborators |
| :---- | :---- | :---- |
| Account Service | • CreateOrganizerProfile()<br>• CreateSponsorProfile()<br>• GetAccount()<br>• GetOrganizerProfile()<br>• GetSponsorProfile() | Supabase Auth<br>PostgreSQL (account) |
| Tournament Service | • CreateTournamentDraft()<br>• ConfigureCrowdfunding()<br>• ConfigureRegistrationForm()<br>• UpdateDraftSettings()<br>• OpenTournament()<br>• GetTournament()<br>• GetFundingProgress()<br>• CreateTeamLobby()<br>• JoinTeamLobby()<br>• LockTeamRoster()<br>• SubmitRegistration()<br>• RecordSponsorship()<br>• RecordMatchResult() | Account Service<br>• BatchGetAccounts()<br>Payment Service (gRPC)<br>• CreatePayment()<br>• RefundRegistration()<br>Object Storage (S3-compatible)<br>• store/fetch sponsor logo<br>Message Broker<br>• publish registration.submitted<br>• publish announcement.created<br>PostgreSQL (tournament)<br>MongoDB (registration forms) |
| Payment Service | • CreatePayment()<br>• GetPayment()<br>• RefundRegistration()<br>• CompleteMock()<br>• RecordLedgerEntry() | Message Broker<br>• publish payment.confirmed<br>PostgreSQL (payment) |
| Notification Service | • SendEmailNotification()<br>• PublishInAppNotification() | Message Broker<br>• consume payment.succeeded<br>• consume registration.submitted<br>• consume match_result.recorded<br>• consume sponsorship.confirmed<br>Email Provider<br>PostgreSQL (notification) |
