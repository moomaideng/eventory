package seedids

import (
	"fmt"

	"github.com/google/uuid"
)

// Dev Accounts (mapped in internal/middlewares/auth.go)
var (
	AlexID   = uuid.MustParse("99999999-0000-4000-8000-000000000001") // alex-dev (All-rounder: Org + Sponsor)
	BobID    = uuid.MustParse("99999999-0000-4000-8000-000000000002") // bob-dev (Organizer only: Bob Events)
	CarterID = uuid.MustParse("99999999-0000-4000-8000-000000000003") // carter-dev (Sponsor only: Carter Fundings)
)

// Partner, Brand & Edge Accounts
var (
	MayaID = uuid.MustParse("00000000-0000-4000-8000-000000000003")

	// Named Competitors / Captains
	KaitoID  = uuid.MustParse("00000000-0000-4000-8000-000000000010")
	PlengID  = uuid.MustParse("00000000-0000-4000-8000-000000000011")
	TonklaID = uuid.MustParse("00000000-0000-4000-8000-000000000012")
	BeamID   = uuid.MustParse("00000000-0000-4000-8000-000000000013")
	FahID    = uuid.MustParse("00000000-0000-4000-8000-000000000014")
	MekID    = uuid.MustParse("00000000-0000-4000-8000-000000000015")

	// Sponsor Directory Showcase Brands
	BrandHyperVoltID   = uuid.MustParse("00000000-0000-4000-8000-000000000021")
	BrandApexGearID    = uuid.MustParse("00000000-0000-4000-8000-000000000022")
	BrandCloud9THID    = uuid.MustParse("00000000-0000-4000-8000-000000000023")
	BrandSiamFiberID   = uuid.MustParse("00000000-0000-4000-8000-000000000024")
	BrandKraitLabID    = uuid.MustParse("00000000-0000-4000-8000-000000000025")
	BrandCUAlumniID    = uuid.MustParse("00000000-0000-4000-8000-000000000026")
	BrandPulsePayID    = uuid.MustParse("00000000-0000-4000-8000-000000000027")
	BrandPixelStreamID = uuid.MustParse("00000000-0000-4000-8000-000000000028")

	// Edge-case accounts
	StealthSponsorID = uuid.MustParse("00000000-0000-4000-8000-000000000029")
	OnboardingUserID = uuid.MustParse("00000000-0000-4000-8000-000000000030")
)

func PlayerID(n int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-4000-8000-%012d", 100+n))
}

// Tournaments: Group A (Feature Tests) & Group B (Real-looking Tournaments)
var (
	TournTestDraft            = Fixture(0x33333333, 1)
	TournTestCrowdGoalReached = Fixture(0x33333333, 2)
	TournTestCrowdUnderfunded = Fixture(0x33333333, 3)
	TournTestSponsorCheckout  = Fixture(0x33333333, 4)
	TournTestReviewQueue      = Fixture(0x33333333, 5)
	TournTestSoloPaid         = Fixture(0x33333333, 6)
	TournTestBracketSetup     = Fixture(0x33333333, 7)
	TournTestPermAdmin        = Fixture(0x33333333, 8)
	TournTestPermReferee      = Fixture(0x33333333, 9)
	TournTestBracketLive      = Fixture(0x33333333, 10)
	TournTestBracketCompleted = Fixture(0x33333333, 11)

	TournRealDraftCU       = Fixture(0x33333333, 101)
	TournRealCrowdBCA      = Fixture(0x33333333, 102)
	TournRealRegValorant   = Fixture(0x33333333, 103)
	TournRealRegSF6Solo    = Fixture(0x33333333, 104)
	TournRealOngoingDota2  = Fixture(0x33333333, 105)
	TournRealCompletedMLBB = Fixture(0x33333333, 106)
)

func Fixture(prefix uint32, suffix int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("%08x-0000-4000-8000-%012d", prefix, suffix))
}

func TierID(n int) uuid.UUID         { return Fixture(0x55555555, n) }
func PledgeID(n int) uuid.UUID       { return Fixture(0x66666666, n) }
func InvoiceID(n int) uuid.UUID      { return Fixture(0x66666667, n) }
func ReceiptID(n int) uuid.UUID      { return Fixture(0x66666668, n) }
func AttemptID(n int) uuid.UUID      { return Fixture(0x66666669, n) }
func MatchID(n int) uuid.UUID        { return Fixture(0x77777777, n) }
func MatchEntryID(n int) uuid.UUID   { return Fixture(0x88888888, n) }
func AnnouncementID(n int) uuid.UUID { return Fixture(0xaaaaaaaa, n) }
func ReviewID(n int) uuid.UUID       { return Fixture(0xbbbbbbbb, n) }
func NotificationID(n int) uuid.UUID { return Fixture(0xcccccccc, n) }

// Cross-service invoice fixtures linking Tournament (pledges/teams) <-> Payment (invoices/receipts)
type SponsorshipInvoiceFixture struct {
	PledgeID       uuid.UUID
	InvoiceID      uuid.UUID
	ReceiptID      uuid.UUID
	PayerAccountID uuid.UUID
	TournamentName string
	TierName       string
	AmountMinor    int64
	Status         string // PAID, UNPAID, EXPIRED, CANCELLED
}

type RegistrationInvoiceFixture struct {
	TournamentID   uuid.UUID
	TournamentName string
	TeamID         string
	AttemptID      uuid.UUID
	InvoiceID      uuid.UUID
	ReceiptID      uuid.UUID
	PayerAccountID uuid.UUID
	AmountMinor    int64
	Status         string // PAID, UNPAID, REFUNDED
	RefundReason   string
}

func SponsorshipInvoices() []SponsorshipInvoiceFixture {
	return []SponsorshipInvoiceFixture{
		// A2: Goal Reached (35,000 / 30,000 THB)
		{PledgeID: PledgeID(1), InvoiceID: InvoiceID(1), ReceiptID: ReceiptID(1), PayerAccountID: CarterID, TournamentName: "[TEST · Crowdfunding] Goal Reached (Normal Open)", TierName: "Platinum", AmountMinor: 2000000, Status: "PAID"},
		{PledgeID: PledgeID(2), InvoiceID: InvoiceID(2), ReceiptID: ReceiptID(2), PayerAccountID: BrandHyperVoltID, TournamentName: "[TEST · Crowdfunding] Goal Reached (Normal Open)", TierName: "Gold", AmountMinor: 1500000, Status: "PAID"},

		// A3: Underfunded & Tier Locks (15,000 confirmed + 5,000 pending + expired + cancelled)
		{PledgeID: PledgeID(3), InvoiceID: InvoiceID(3), ReceiptID: ReceiptID(3), PayerAccountID: CarterID, TournamentName: "[TEST · Crowdfunding] Underfunded & Tier Locks", TierName: "Gold (Sold Out)", AmountMinor: 1500000, Status: "PAID"},
		{PledgeID: PledgeID(4), InvoiceID: InvoiceID(4), ReceiptID: ReceiptID(4), PayerAccountID: MayaID, TournamentName: "[TEST · Crowdfunding] Underfunded & Tier Locks", TierName: "Silver (Locked)", AmountMinor: 500000, Status: "UNPAID"},
		{PledgeID: PledgeID(5), InvoiceID: InvoiceID(5), ReceiptID: ReceiptID(5), PayerAccountID: BrandKraitLabID, TournamentName: "[TEST · Crowdfunding] Underfunded & Tier Locks", TierName: "Community (Editable)", AmountMinor: 100000, Status: "EXPIRED"},
		{PledgeID: PledgeID(6), InvoiceID: InvoiceID(6), ReceiptID: ReceiptID(6), PayerAccountID: BrandPulsePayID, TournamentName: "[TEST · Crowdfunding] Underfunded & Tier Locks", TierName: "Community (Editable)", AmountMinor: 100000, Status: "CANCELLED"},

		// A4: Sponsor Checkout Playground (Carter confirmed Bronze; Alex free to test checkout)
		{PledgeID: PledgeID(7), InvoiceID: InvoiceID(7), ReceiptID: ReceiptID(7), PayerAccountID: CarterID, TournamentName: "[TEST · Sponsor] Checkout Playground", TierName: "Bronze", AmountMinor: 100000, Status: "PAID"},

		// A6: Solo Paid (Carter confirmed)
		{PledgeID: PledgeID(8), InvoiceID: InvoiceID(8), ReceiptID: ReceiptID(8), PayerAccountID: CarterID, TournamentName: "[TEST · Solo] Paid Entry & Form Playground", TierName: "Community Supporter", AmountMinor: 300000, Status: "PAID"},

		// B2: Bangkok Cyber Arena 2026 (65,000 / 100,000 THB)
		{PledgeID: PledgeID(9), InvoiceID: InvoiceID(9), ReceiptID: ReceiptID(9), PayerAccountID: BrandHyperVoltID, TournamentName: "Bangkok Cyber Arena 2026", TierName: "Title Partner", AmountMinor: 3500000, Status: "PAID"},
		{PledgeID: PledgeID(10), InvoiceID: InvoiceID(10), ReceiptID: ReceiptID(10), PayerAccountID: BrandApexGearID, TournamentName: "Bangkok Cyber Arena 2026", TierName: "Hardware Partner", AmountMinor: 2000000, Status: "PAID"},
		{PledgeID: PledgeID(11), InvoiceID: InvoiceID(11), ReceiptID: ReceiptID(11), PayerAccountID: BrandCloud9THID, TournamentName: "Bangkok Cyber Arena 2026", TierName: "Server Partner", AmountMinor: 1000000, Status: "PAID"},

		// B3: SEA Valorant Challengers Open (Alex owns; Carter & Apex Gear confirmed)
		{PledgeID: PledgeID(12), InvoiceID: InvoiceID(12), ReceiptID: ReceiptID(12), PayerAccountID: CarterID, TournamentName: "SEA Valorant Challengers Open", TierName: "Gold", AmountMinor: 1500000, Status: "PAID"},
		{PledgeID: PledgeID(13), InvoiceID: InvoiceID(13), ReceiptID: ReceiptID(13), PayerAccountID: BrandApexGearID, TournamentName: "SEA Valorant Challengers Open", TierName: "Silver", AmountMinor: 750000, Status: "PAID"},

		// B5: ASEAN Dota 2 Invitational 2026 (Bob owns; Alex, Carter, Siam Fiber confirmed)
		{PledgeID: PledgeID(14), InvoiceID: InvoiceID(14), ReceiptID: ReceiptID(14), PayerAccountID: AlexID, TournamentName: "ASEAN Dota 2 Invitational 2026", TierName: "Platinum", AmountMinor: 4000000, Status: "PAID"},
		{PledgeID: PledgeID(15), InvoiceID: InvoiceID(15), ReceiptID: ReceiptID(15), PayerAccountID: CarterID, TournamentName: "ASEAN Dota 2 Invitational 2026", TierName: "Gold", AmountMinor: 3000000, Status: "PAID"},
		{PledgeID: PledgeID(16), InvoiceID: InvoiceID(16), ReceiptID: ReceiptID(16), PayerAccountID: BrandSiamFiberID, TournamentName: "ASEAN Dota 2 Invitational 2026", TierName: "Broadcast Partner", AmountMinor: 2000000, Status: "PAID"},

		// B6: Thailand Mobile Legends Championship S4 (Maya owns; Carter, PulsePay, PixelStream confirmed)
		{PledgeID: PledgeID(17), InvoiceID: InvoiceID(17), ReceiptID: ReceiptID(17), PayerAccountID: CarterID, TournamentName: "Thailand Mobile Legends Championship S4", TierName: "Finals Partner", AmountMinor: 2500000, Status: "PAID"},
		{PledgeID: PledgeID(18), InvoiceID: InvoiceID(18), ReceiptID: ReceiptID(18), PayerAccountID: BrandPulsePayID, TournamentName: "Thailand Mobile Legends Championship S4", TierName: "Payment Partner", AmountMinor: 1500000, Status: "PAID"},
		{PledgeID: PledgeID(19), InvoiceID: InvoiceID(19), ReceiptID: ReceiptID(19), PayerAccountID: BrandPixelStreamID, TournamentName: "Thailand Mobile Legends Championship S4", TierName: "Stream Partner", AmountMinor: 1000000, Status: "PAID"},
	}
}

func RegistrationInvoices() []RegistrationInvoiceFixture {
	return []RegistrationInvoiceFixture{
		// A5: [TEST · Review] Queue, Forms & Refunds (500 THB = 50000 satang)
		{TournamentID: TournTestReviewQueue, TournamentName: "[TEST · Review] Queue, Forms & Refunds", TeamID: "TM050003", AttemptID: AttemptID(1), InvoiceID: InvoiceID(101), ReceiptID: ReceiptID(101), PayerAccountID: TonklaID, AmountMinor: 50000, Status: "UNPAID"},
		{TournamentID: TournTestReviewQueue, TournamentName: "[TEST · Review] Queue, Forms & Refunds", TeamID: "TM050004", AttemptID: AttemptID(2), InvoiceID: InvoiceID(102), ReceiptID: ReceiptID(102), PayerAccountID: BeamID, AmountMinor: 50000, Status: "PAID"},
		{TournamentID: TournTestReviewQueue, TournamentName: "[TEST · Review] Queue, Forms & Refunds", TeamID: "TM050005", AttemptID: AttemptID(3), InvoiceID: InvoiceID(103), ReceiptID: ReceiptID(103), PayerAccountID: PlayerID(5), AmountMinor: 50000, Status: "PAID"},
		{TournamentID: TournTestReviewQueue, TournamentName: "[TEST · Review] Queue, Forms & Refunds", TeamID: "TM050006", AttemptID: AttemptID(4), InvoiceID: InvoiceID(104), ReceiptID: ReceiptID(104), PayerAccountID: PlayerID(7), AmountMinor: 50000, Status: "PAID"},
		{TournamentID: TournTestReviewQueue, TournamentName: "[TEST · Review] Queue, Forms & Refunds", TeamID: "TM050007", AttemptID: AttemptID(5), InvoiceID: InvoiceID(105), ReceiptID: ReceiptID(105), PayerAccountID: PlayerID(9), AmountMinor: 50000, Status: "PAID"},
		{TournamentID: TournTestReviewQueue, TournamentName: "[TEST · Review] Queue, Forms & Refunds", TeamID: "TM050008", AttemptID: AttemptID(6), InvoiceID: InvoiceID(106), ReceiptID: ReceiptID(106), PayerAccountID: PlayerID(11), AmountMinor: 50000, Status: "PAID"},
		{TournamentID: TournTestReviewQueue, TournamentName: "[TEST · Review] Queue, Forms & Refunds", TeamID: "TM050009", AttemptID: AttemptID(7), InvoiceID: InvoiceID(107), ReceiptID: ReceiptID(107), PayerAccountID: PlayerID(13), AmountMinor: 50000, Status: "REFUNDED", RefundReason: "Registration rejected by organizer"},
		{TournamentID: TournTestReviewQueue, TournamentName: "[TEST · Review] Queue, Forms & Refunds", TeamID: "TM050010", AttemptID: AttemptID(8), InvoiceID: InvoiceID(108), ReceiptID: ReceiptID(108), PayerAccountID: PlayerID(15), AmountMinor: 50000, Status: "REFUNDED", RefundReason: "Registration rejected by organizer"},

		// A6: [TEST · Solo] Paid Entry & Form Playground (250 THB = 25000 satang)
		{TournamentID: TournTestSoloPaid, TournamentName: "[TEST · Solo] Paid Entry & Form Playground", TeamID: "SL060001", AttemptID: AttemptID(9), InvoiceID: InvoiceID(109), ReceiptID: ReceiptID(109), PayerAccountID: KaitoID, AmountMinor: 25000, Status: "PAID"},
		{TournamentID: TournTestSoloPaid, TournamentName: "[TEST · Solo] Paid Entry & Form Playground", TeamID: "SL060002", AttemptID: AttemptID(10), InvoiceID: InvoiceID(110), ReceiptID: ReceiptID(110), PayerAccountID: PlengID, AmountMinor: 25000, Status: "PAID"},
		{TournamentID: TournTestSoloPaid, TournamentName: "[TEST · Solo] Paid Entry & Form Playground", TeamID: "SL060003", AttemptID: AttemptID(11), InvoiceID: InvoiceID(111), ReceiptID: ReceiptID(111), PayerAccountID: TonklaID, AmountMinor: 25000, Status: "PAID"},
		{TournamentID: TournTestSoloPaid, TournamentName: "[TEST · Solo] Paid Entry & Form Playground", TeamID: "SL060004", AttemptID: AttemptID(12), InvoiceID: InvoiceID(112), ReceiptID: ReceiptID(112), PayerAccountID: BeamID, AmountMinor: 25000, Status: "PAID"},
		{TournamentID: TournTestSoloPaid, TournamentName: "[TEST · Solo] Paid Entry & Form Playground", TeamID: "SL060005", AttemptID: AttemptID(13), InvoiceID: InvoiceID(113), ReceiptID: ReceiptID(113), PayerAccountID: FahID, AmountMinor: 25000, Status: "PAID"},
		{TournamentID: TournTestSoloPaid, TournamentName: "[TEST · Solo] Paid Entry & Form Playground", TeamID: "SL060006", AttemptID: AttemptID(14), InvoiceID: InvoiceID(114), ReceiptID: ReceiptID(114), PayerAccountID: PlayerID(19), AmountMinor: 25000, Status: "PAID"},
		{TournamentID: TournTestSoloPaid, TournamentName: "[TEST · Solo] Paid Entry & Form Playground", TeamID: "SL060007", AttemptID: AttemptID(15), InvoiceID: InvoiceID(115), ReceiptID: ReceiptID(115), PayerAccountID: PlayerID(20), AmountMinor: 25000, Status: "REFUNDED", RefundReason: "Registration rejected by organizer"},

		// B3: SEA Valorant Challengers Open (800 THB = 80000 satang)
		{TournamentID: TournRealRegValorant, TournamentName: "SEA Valorant Challengers Open", TeamID: "TM130001", AttemptID: AttemptID(16), InvoiceID: InvoiceID(116), ReceiptID: ReceiptID(116), PayerAccountID: KaitoID, AmountMinor: 80000, Status: "PAID"},
		{TournamentID: TournRealRegValorant, TournamentName: "SEA Valorant Challengers Open", TeamID: "TM130002", AttemptID: AttemptID(17), InvoiceID: InvoiceID(117), ReceiptID: ReceiptID(117), PayerAccountID: PlengID, AmountMinor: 80000, Status: "PAID"},
		{TournamentID: TournRealRegValorant, TournamentName: "SEA Valorant Challengers Open", TeamID: "TM130003", AttemptID: AttemptID(18), InvoiceID: InvoiceID(118), ReceiptID: ReceiptID(118), PayerAccountID: MekID, AmountMinor: 80000, Status: "PAID"},
		{TournamentID: TournRealRegValorant, TournamentName: "SEA Valorant Challengers Open", TeamID: "TM130004", AttemptID: AttemptID(19), InvoiceID: InvoiceID(119), ReceiptID: ReceiptID(119), PayerAccountID: PlayerID(13), AmountMinor: 80000, Status: "PAID"},
		{TournamentID: TournRealRegValorant, TournamentName: "SEA Valorant Challengers Open", TeamID: "TM130005", AttemptID: AttemptID(20), InvoiceID: InvoiceID(120), ReceiptID: ReceiptID(120), PayerAccountID: FahID, AmountMinor: 80000, Status: "PAID"},
		{TournamentID: TournRealRegValorant, TournamentName: "SEA Valorant Challengers Open", TeamID: "TM130006", AttemptID: AttemptID(21), InvoiceID: InvoiceID(121), ReceiptID: ReceiptID(121), PayerAccountID: PlayerID(22), AmountMinor: 80000, Status: "PAID"},
		{TournamentID: TournRealRegValorant, TournamentName: "SEA Valorant Challengers Open", TeamID: "TM130008", AttemptID: AttemptID(22), InvoiceID: InvoiceID(122), ReceiptID: ReceiptID(122), PayerAccountID: BeamID, AmountMinor: 80000, Status: "REFUNDED", RefundReason: "Registration rejected by organizer"},
	}
}
