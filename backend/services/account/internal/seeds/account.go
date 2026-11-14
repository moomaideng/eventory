package seeds

import (
	"fmt"
	"time"

	"github.com/moomaideng/eventory/internal/seedids"
	"github.com/moomaideng/eventory/services/account/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedAccounts(db *gorm.DB) error {
	accounts := []models.Account{
		// --- Dev Quick Login Accounts ---
		{
			ID:          seedids.AlexID,
			Email:       "alex.dev@eventory.gg",
			Handle:      "alex_dev",
			DisplayName: "Alex Dev",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.BobID,
			Email:       "bob.dev@eventory.gg",
			Handle:      "bob_dev",
			DisplayName: "Bob Dev",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.CarterID,
			Email:       "carter.dev@eventory.gg",
			Handle:      "carter_dev",
			DisplayName: "Carter Dev",
			Status:      models.AccountStatusActive,
		},

		// --- Multi-role Partner ---
		{
			ID:          seedids.MayaID,
			Email:       "maya@example.com",
			Handle:      "maya_works",
			DisplayName: "Maya Works",
			Status:      models.AccountStatusActive,
		},

		// --- Named Competitors / Captains ---
		{
			ID:          seedids.KaitoID,
			Email:       "kaito@example.com",
			Handle:      "kaito_th",
			DisplayName: "Kaito Tanaka",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.PlengID,
			Email:       "pleng@example.com",
			Handle:      "pleng_gg",
			DisplayName: "Pleng Nattapong",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.TonklaID,
			Email:       "tonkla@example.com",
			Handle:      "tonkla_bkk",
			DisplayName: "Tonkla S.",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.BeamID,
			Email:       "beam@example.com",
			Handle:      "beam_val",
			DisplayName: "Beam W.",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.FahID,
			Email:       "fah@example.com",
			Handle:      "fah_dota",
			DisplayName: "Fah Chanya",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.MekID,
			Email:       "mek@example.com",
			Handle:      "mek_sf6",
			DisplayName: "Mek Pakorn",
			Status:      models.AccountStatusActive,
		},

		// --- Sponsor Directory Showcase Brands ---
		{
			ID:          seedids.BrandHyperVoltID,
			Email:       "hypervolt@example.com",
			Handle:      "hypervolt_th",
			DisplayName: "HyperVolt Energy",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.BrandApexGearID,
			Email:       "apexgear@example.com",
			Handle:      "apex_gear",
			DisplayName: "Apex Peripherals",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.BrandCloud9THID,
			Email:       "cloud9th@example.com",
			Handle:      "cloud9_th",
			DisplayName: "Cloud9TH Hosting",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.BrandSiamFiberID,
			Email:       "siamfiber@example.com",
			Handle:      "siam_fiber",
			DisplayName: "Siam Fiber Net",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.BrandKraitLabID,
			Email:       "kraitlab@example.com",
			Handle:      "krait_lab",
			DisplayName: "Krait Gear Lab",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.BrandCUAlumniID,
			Email:       "cualumni@example.com",
			Handle:      "cu_esports_alumni",
			DisplayName: "CU Alumni Esports",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.BrandPulsePayID,
			Email:       "pulsepay@example.com",
			Handle:      "pulsepay_th",
			DisplayName: "PulsePay Wallet",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.BrandPixelStreamID,
			Email:       "pixelstream@example.com",
			Handle:      "pixelstream",
			DisplayName: "PixelStream Studio",
			Status:      models.AccountStatusActive,
		},

		// --- Edge-Case Accounts ---
		{
			ID:          seedids.StealthSponsorID,
			Email:       "stealth@example.com",
			Handle:      "stealth_brand",
			DisplayName: "Stealth Brand (Hidden)",
			Status:      models.AccountStatusActive,
		},
		{
			ID:          seedids.OnboardingUserID,
			Email:       "newcomer@example.com",
			Handle:      "newcomer_pending",
			DisplayName: "Newcomer (Onboarding)",
			Status:      models.AccountStatusOnboarding,
		},
	}

	for number := 1; number <= 40; number++ {
		accounts = append(accounts, models.Account{
			ID:          seedids.PlayerID(number),
			Email:       fmt.Sprintf("player%02d@example.com", number),
			Handle:      fmt.Sprintf("player_%02d", number),
			DisplayName: fmt.Sprintf("Player %02d", number),
			Status:      models.AccountStatusActive,
		})
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"email", "handle", "display_name", "status"}),
	}).Create(&accounts).Error; err != nil {
		return err
	}

	seededConsentAt := time.Now().UTC()

	organizerProfiles := []models.OrganizerProfile{
		// Alex Dev: Organizer + Sponsor
		{
			AccountID:          seedids.AlexID,
			OrganizerName:      "Alex Events",
			OrganizerEmail:     "alex.events@eventory.gg",
			PrivacyConsentedAt: &seededConsentAt,
		},
		// Bob Dev: Organizer ONLY ("Bob Events")
		{
			AccountID:          seedids.BobID,
			OrganizerName:      "Bob Events",
			OrganizerEmail:     "bob.events@eventory.gg",
			PrivacyConsentedAt: &seededConsentAt,
		},
		// Maya: Organizer + Sponsor
		{
			AccountID:          seedids.MayaID,
			OrganizerName:      "Maya Works",
			OrganizerEmail:     "maya.events@example.com",
			PrivacyConsentedAt: &seededConsentAt,
		},
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"organizer_name", "organizer_email", "privacy_consented_at"}),
	}).Create(&organizerProfiles).Error; err != nil {
		return err
	}

	sponsorProfiles := []models.SponsorProfile{
		// Alex Dev: Organizer + Sponsor
		{
			AccountID:          seedids.AlexID,
			SponsorName:        "Alex Ventures",
			SponsorEmail:       "alex.ventures@eventory.gg",
			WebsiteURL:         "https://alexventures.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=AV",
			Tags:               []string{"gaming", "esports", "hardware"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		// Carter Dev: Sponsor ONLY ("Carter Fundings")
		{
			AccountID:          seedids.CarterID,
			SponsorName:        "Carter Fundings",
			SponsorEmail:       "carter.fundings@eventory.gg",
			WebsiteURL:         "https://carterfundings.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=CF",
			Tags:               []string{"fintech", "esports", "thai-community"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		// Maya: Organizer + Sponsor
		{
			AccountID:          seedids.MayaID,
			SponsorName:        "Maya Works Studio",
			SponsorEmail:       "studio@example.com",
			WebsiteURL:         "https://mayaworks.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=MW",
			Tags:               []string{"gaming", "mobile-gaming", "streaming"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		// 8 Directory Showcase Brands
		{
			AccountID:          seedids.BrandHyperVoltID,
			SponsorName:        "HyperVolt Energy",
			SponsorEmail:       "partners@hypervolt.example.com",
			WebsiteURL:         "https://hypervolt.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=HV",
			Tags:               []string{"energy-drink", "esports", "thai-community"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		{
			AccountID:          seedids.BrandApexGearID,
			SponsorName:        "Apex Peripherals",
			SponsorEmail:       "sponsorships@apexperipherals.example.com",
			WebsiteURL:         "https://apexperipherals.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=AP",
			Tags:               []string{"hardware", "peripherals", "esports"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		{
			AccountID:          seedids.BrandCloud9THID,
			SponsorName:        "Cloud9TH Hosting",
			SponsorEmail:       "events@cloud9th.example.com",
			WebsiteURL:         "https://cloud9th.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=C9",
			Tags:               []string{"cloud", "esports", "streaming"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		{
			AccountID:          seedids.BrandSiamFiberID,
			SponsorName:        "Siam Fiber Net",
			SponsorEmail:       "marketing@siamfiber.example.com",
			WebsiteURL:         "https://siamfiber.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=SF",
			Tags:               []string{"telecom", "streaming", "thai-community"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		{
			AccountID:          seedids.BrandKraitLabID,
			SponsorName:        "Krait Gear Lab",
			SponsorEmail:       "hello@kraitgear.example.com",
			WebsiteURL:         "https://kraitgear.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=KG",
			Tags:               []string{"hardware", "peripherals", "gaming"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		{
			AccountID:          seedids.BrandCUAlumniID,
			SponsorName:        "CU Alumni Esports",
			SponsorEmail:       "alumni@cuesports.example.com",
			WebsiteURL:         "https://cuesports.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=CU",
			Tags:               []string{"university", "thai-community", "esports"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		{
			AccountID:          seedids.BrandPulsePayID,
			SponsorName:        "PulsePay Wallet",
			SponsorEmail:       "growth@pulsepay.example.com",
			WebsiteURL:         "https://pulsepay.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=PP",
			Tags:               []string{"fintech", "mobile-gaming", "thai-community"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		{
			AccountID:          seedids.BrandPixelStreamID,
			SponsorName:        "PixelStream Studio",
			SponsorEmail:       "brand@pixelstream.example.com",
			WebsiteURL:         "https://pixelstream.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=PS",
			Tags:               []string{"streaming", "gaming", "university"},
			PrivacyConsentedAt: &seededConsentAt,
		},
		// Stealth sponsor: empty SponsorEmail -> excluded from public organizer sponsor directory
		{
			AccountID:          seedids.StealthSponsorID,
			SponsorName:        "Stealth Brand",
			SponsorEmail:       "",
			WebsiteURL:         "https://stealth.example.com",
			LogoURL:            "https://placehold.co/128x128/png?text=SB",
			Tags:               []string{"esports"},
			PrivacyConsentedAt: &seededConsentAt,
		},
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"sponsor_name", "sponsor_email", "website_url", "logo_url", "tags", "privacy_consented_at"}),
	}).Create(&sponsorProfiles).Error
}
