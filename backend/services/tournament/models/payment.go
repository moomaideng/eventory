package models

// PaymentStatus is the status of an invoice as the Payment service reports it.
type PaymentStatus string

const (
	PaymentUnpaid    PaymentStatus = "UNPAID"
	PaymentPaid      PaymentStatus = "PAID"
	PaymentExpired   PaymentStatus = "EXPIRED"
	PaymentCancelled PaymentStatus = "CANCELLED"
	PaymentRefunded  PaymentStatus = "REFUNDED"
)

// Invoice sources Tournament creates. Payment echoes them back, so Tournament
// can tell which of its records an invoice or payment event belongs to.
const (
	PaymentSourceTeamRegistration = "TEAM_REGISTRATION"
	PaymentSourceSponsorship      = "SPONSORSHIP"
)
