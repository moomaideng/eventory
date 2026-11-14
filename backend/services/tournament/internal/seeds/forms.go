package seeds

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/seedids"
	"github.com/moomaideng/eventory/services/tournament/internal/forms"
	"github.com/moomaideng/eventory/services/tournament/internal/repositories"
	"github.com/moomaideng/eventory/services/tournament/models"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// SeedForms ensures MongoDB indexes and upserts custom registration forms and
// member answers for showcase tournaments (A5, A6, B1, B3, B4).
func SeedForms(ctx context.Context, mongoDB *mongo.Database) error {
	repo := repositories.NewRegistrationFormRepository(mongoDB)
	if err := repo.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("ensure registration form indexes: %w", err)
	}

	formDocs, responses, err := BuildFormFixtures()
	if err != nil {
		return err
	}

	for i := range formDocs {
		doc := formDocs[i]
		if _, err := repo.UpsertForm(ctx, &doc); err != nil {
			return fmt.Errorf("upsert form %s: %w", doc.TournamentID, err)
		}
	}

	for i := range responses {
		resp := responses[i]
		if _, err := repo.UpsertResponse(ctx, &resp); err != nil {
			return fmt.Errorf("upsert response %s/%s: %w", resp.TournamentID, resp.AccountID, err)
		}
	}

	return nil
}

// BuildFormFixtures constructs and validates all seeded MongoDB registration
// forms and participant answers against forms.NewDefaultRegistry().
func BuildFormFixtures() ([]models.RegistrationForm, []models.FormResponse, error) {
	registry := forms.NewDefaultRegistry()

	formDocs := []models.RegistrationForm{
		// A5: [TEST · Review] Queue, Forms & Refunds (all 5 field types)
		{
			TournamentID: seedids.TournTestReviewQueue,
			Fields: []models.FormField{
				{
					ID:          "riot_id",
					Type:        models.FieldTypeShortText,
					Label:       "Riot ID (Name#TAG)",
					Description: "Exact in-game Riot ID used for lobby invites",
					Required:    true,
					Order:       0,
					Config:      map[string]any{},
				},
				{
					ID:          "comp_bio",
					Type:        models.FieldTypeLongText,
					Label:       "Competitive Background",
					Description: "Share recent scrim or premier experience",
					Required:    true,
					Order:       1,
					Config:      map[string]any{},
				},
				{
					ID:          "peak_rr",
					Type:        models.FieldTypeNumber,
					Label:       "Current Peak RR / Rating",
					Description: "Numeric rating between 0 and 1500",
					Required:    true,
					Order:       2,
					Config: map[string]any{
						"min":           float64(0),
						"max":           float64(1500),
						"allowDecimals": false,
					},
				},
				{
					ID:       "server_region",
					Type:     models.FieldTypeSelectOne,
					Label:    "Preferred Server",
					Required: true,
					Order:    3,
					Config: map[string]any{
						"options": []models.FormFieldOption{
							{ID: "bkk", Label: "Bangkok (TH)"},
							{ID: "sg", Label: "Singapore (SG)"},
							{ID: "hk", Label: "Hong Kong (HK)"},
						},
					},
				},
				{
					ID:          "agent_roles",
					Type:        models.FieldTypeSelectMultiple,
					Label:       "Playable Roles (Pick 1-3)",
					Description: "Select 1 to 3 roles you can fill in tournament play",
					Required:    true,
					Order:       4,
					Config: map[string]any{
						"options": []models.FormFieldOption{
							{ID: "duelist", Label: "Duelist"},
							{ID: "initiator", Label: "Initiator"},
							{ID: "controller", Label: "Controller"},
							{ID: "sentinel", Label: "Sentinel"},
						},
						"minSelections": 1,
						"maxSelections": 3,
					},
				},
			},
		},

		// A6: [TEST · Solo] Paid Entry & Form Playground (all 5 field types)
		{
			TournamentID: seedids.TournTestSoloPaid,
			Fields: []models.FormField{
				{
					ID:          "cfn_id",
					Type:        models.FieldTypeShortText,
					Label:       "Capcom Fighters Network (CFN) ID",
					Description: "Used by bracket marshals to verify match rooms",
					Required:    true,
					Order:       0,
					Config:      map[string]any{},
				},
				{
					ID:       "fgc_notes",
					Type:     models.FieldTypeLongText,
					Label:    "Tournament History / Seeding Notes",
					Required: false,
					Order:    1,
					Config:   map[string]any{},
				},
				{
					ID:          "mr_points",
					Type:        models.FieldTypeNumber,
					Label:       "Master Rating (MR)",
					Description: "Enter 0 if below Master rank",
					Required:    true,
					Order:       2,
					Config: map[string]any{
						"min":           float64(0),
						"max":           float64(3000),
						"allowDecimals": false,
					},
				},
				{
					ID:       "platform",
					Type:     models.FieldTypeSelectOne,
					Label:    "Controller Setup",
					Required: true,
					Order:    3,
					Config: map[string]any{
						"options": []models.FormFieldOption{
							{ID: "pad", Label: "PS5 / Xbox Pad"},
							{ID: "stick", Label: "Arcade Fightstick"},
							{ID: "leverless", Label: "Leverless / Hitbox"},
						},
					},
				},
				{
					ID:       "avail_slots",
					Type:     models.FieldTypeSelectMultiple,
					Label:    "Available Check-in Windows",
					Required: true,
					Order:    4,
					Config: map[string]any{
						"options": []models.FormFieldOption{
							{ID: "sat_am", Label: "Saturday Morning"},
							{ID: "sat_pm", Label: "Saturday Afternoon"},
							{ID: "sun_pm", Label: "Sunday Finals"},
						},
						"minSelections": 1,
						"maxSelections": 3,
					},
				},
			},
		},

		// B1: Chula Inter-Faculty Valorant Cup 2026 (DRAFT pre-configured form)
		{
			TournamentID: seedids.TournRealDraftCU,
			Fields: []models.FormField{
				{
					ID:       "cu_faculty",
					Type:     models.FieldTypeSelectOne,
					Label:    "Chulalongkorn Faculty",
					Required: true,
					Order:    0,
					Config: map[string]any{
						"options": []models.FormFieldOption{
							{ID: "eng", Label: "Engineering"},
							{ID: "sci", Label: "Science"},
							{ID: "cbs", Label: "Commerce & Accountancy"},
							{ID: "arts", Label: "Arts / Communication Arts"},
						},
					},
				},
				{
					ID:       "riot_id",
					Type:     models.FieldTypeShortText,
					Label:    "Riot ID (Name#TH1)",
					Required: true,
					Order:    1,
					Config:   map[string]any{},
				},
				{
					ID:       "student_year",
					Type:     models.FieldTypeNumber,
					Label:    "Academic Year (1-4)",
					Required: true,
					Order:    2,
					Config: map[string]any{
						"min":           float64(1),
						"max":           float64(6),
						"allowDecimals": false,
					},
				},
				{
					ID:       "roles",
					Type:     models.FieldTypeSelectMultiple,
					Label:    "Preferred Team Roles",
					Required: true,
					Order:    3,
					Config: map[string]any{
						"options": []models.FormFieldOption{
							{ID: "duelist", Label: "Duelist"},
							{ID: "initiator", Label: "Initiator"},
							{ID: "controller", Label: "Controller"},
							{ID: "sentinel", Label: "Sentinel"},
							{ID: "igl", Label: "In-Game Leader (IGL)"},
						},
						"minSelections": 1,
						"maxSelections": 3,
					},
				},
				{
					ID:       "jersey_note",
					Type:     models.FieldTypeLongText,
					Label:    "Additional Notes for Organizers",
					Required: false,
					Order:    4,
					Config:   map[string]any{},
				},
			},
		},

		// B3: SEA Valorant Challengers Open
		{
			TournamentID: seedids.TournRealRegValorant,
			Fields: []models.FormField{
				{
					ID:       "riot_id",
					Type:     models.FieldTypeShortText,
					Label:    "Official Riot ID",
					Required: true,
					Order:    0,
					Config:   map[string]any{},
				},
				{
					ID:       "peak_rank",
					Type:     models.FieldTypeSelectOne,
					Label:    "Current / Peak Competitive Rank",
					Required: true,
					Order:    1,
					Config: map[string]any{
						"options": []models.FormFieldOption{
							{ID: "ascendant", Label: "Ascendant"},
							{ID: "immortal", Label: "Immortal"},
							{ID: "radiant", Label: "Radiant"},
						},
					},
				},
				{
					ID:       "ping_ms",
					Type:     models.FieldTypeNumber,
					Label:    "Average Singapore Server Ping (ms)",
					Required: true,
					Order:    2,
					Config: map[string]any{
						"min":           float64(1),
						"max":           float64(200),
						"allowDecimals": false,
					},
				},
				{
					ID:       "roles",
					Type:     models.FieldTypeSelectMultiple,
					Label:    "Agent Roles (Pick 1-2)",
					Required: true,
					Order:    3,
					Config: map[string]any{
						"options": []models.FormFieldOption{
							{ID: "duelist", Label: "Duelist"},
							{ID: "initiator", Label: "Initiator"},
							{ID: "controller", Label: "Controller"},
							{ID: "sentinel", Label: "Sentinel"},
						},
						"minSelections": 1,
						"maxSelections": 2,
					},
				},
				{
					ID:       "team_bio",
					Type:     models.FieldTypeLongText,
					Label:    "Roster Accomplishments",
					Required: false,
					Order:    4,
					Config:   map[string]any{},
				},
			},
		},

		// B4: TH Street Fighter 6 Community Clash (Free SOLO)
		{
			TournamentID: seedids.TournRealRegSF6Solo,
			Fields: []models.FormField{
				{
					ID:       "cfn_id",
					Type:     models.FieldTypeShortText,
					Label:    "CFN Player Handle",
					Required: true,
					Order:    0,
					Config:   map[string]any{},
				},
				{
					ID:       "main_char",
					Type:     models.FieldTypeSelectOne,
					Label:    "Primary Character",
					Required: true,
					Order:    1,
					Config: map[string]any{
						"options": []models.FormFieldOption{
							{ID: "luke", Label: "Luke"},
							{ID: "ken", Label: "Ken"},
							{ID: "cammy", Label: "Cammy"},
							{ID: "akuma", Label: "Akuma"},
							{ID: "chunli", Label: "Chun-Li"},
						},
					},
				},
				{
					ID:       "controller",
					Type:     models.FieldTypeSelectOne,
					Label:    "Controller Type",
					Required: true,
					Order:    2,
					Config: map[string]any{
						"options": []models.FormFieldOption{
							{ID: "pad", Label: "Pad"},
							{ID: "stick", Label: "Arcade Stick"},
							{ID: "leverless", Label: "Leverless"},
						},
					},
				},
				{
					ID:       "seed_rating",
					Type:     models.FieldTypeNumber,
					Label:    "Self-Rated Seed Tier (1-10)",
					Required: true,
					Order:    3,
					Config: map[string]any{
						"min":           float64(1),
						"max":           float64(10),
						"allowDecimals": false,
					},
				},
			},
		},
	}

	// Build responses for every participant on A5 (excluding PlayerID(2) intentionally!), A6, B3, and B4.
	var responses []models.FormResponse

	// A5 accounts (all members in TM050001..TM050010 except PlayerID(2))
	a5Accounts := []uuid.UUID{
		seedids.AlexID, seedids.PlayerID(16), // TM050001
		seedids.PlengID,                      // TM050002 (PlayerID(2) omitted intentionally!)
		seedids.TonklaID, seedids.PlayerID(3), // TM050003
		seedids.BeamID, seedids.PlayerID(4), // TM050004
		seedids.PlayerID(5), seedids.PlayerID(6), // TM050005
		seedids.PlayerID(7), seedids.PlayerID(8), // TM050006
		seedids.PlayerID(9), seedids.PlayerID(10), // TM050007
		seedids.PlayerID(11), seedids.PlayerID(12), // TM050008
		seedids.PlayerID(13), seedids.PlayerID(14), // TM050009
		seedids.PlayerID(15), seedids.PlayerID(17), // TM050010
	}
	servers := []string{"bkk", "sg", "hk"}
	roleChoices := [][]string{{"duelist", "initiator"}, {"controller"}, {"sentinel", "controller"}, {"initiator", "duelist"}}
	for i, acc := range a5Accounts {
		responses = append(responses, models.FormResponse{
			TournamentID: seedids.TournTestReviewQueue,
			AccountID:    acc,
			Answers: []models.FieldAnswer{
				{FieldID: "riot_id", Value: fmt.Sprintf("THPlayer%02d#BKK", i+1)},
				{FieldID: "comp_bio", Value: "Active ranked & community tournament competitor across SEA servers."},
				{FieldID: "peak_rr", Value: 250 + i*35},
				{FieldID: "server_region", Value: servers[i%len(servers)]},
				{FieldID: "agent_roles", Value: roleChoices[i%len(roleChoices)]},
			},
		})
	}

	// A6 accounts (Alex omitted intentionally so dev can test solo join + form submit + checkout)
	a6Accounts := []uuid.UUID{
		seedids.KaitoID, seedids.PlengID, seedids.TonklaID, seedids.BeamID,
		seedids.FahID, seedids.PlayerID(19), seedids.PlayerID(20),
	}
	ctrlOpts := []string{"pad", "stick", "leverless"}
	for i, acc := range a6Accounts {
		responses = append(responses, models.FormResponse{
			TournamentID: seedids.TournTestSoloPaid,
			AccountID:    acc,
			Answers: []models.FieldAnswer{
				{FieldID: "cfn_id", Value: fmt.Sprintf("BKK_FGC_%02d", i+1)},
				{FieldID: "fgc_notes", Value: "Regular competitor at Bangkok weekly FGC locals."},
				{FieldID: "mr_points", Value: 1550 + i*45},
				{FieldID: "platform", Value: ctrlOpts[i%len(ctrlOpts)]},
				{FieldID: "avail_slots", Value: []string{"sat_pm", "sun_pm"}},
			},
		})
	}

	// B3 accounts (TM130001..TM130008, PlayerID(2) is referee so absent)
	b3Accounts := []uuid.UUID{
		seedids.KaitoID, seedids.PlayerID(3), seedids.PlayerID(4), seedids.PlayerID(5), seedids.PlayerID(6),
		seedids.PlengID, seedids.PlayerID(7), seedids.PlayerID(8), seedids.PlayerID(9), seedids.PlayerID(10),
		seedids.MekID, seedids.PlayerID(11), seedids.PlayerID(12), seedids.PlayerID(14), seedids.PlayerID(15),
		seedids.PlayerID(13), seedids.PlayerID(16), seedids.PlayerID(17), seedids.PlayerID(18), seedids.PlayerID(19),
		seedids.FahID, seedids.PlayerID(20), seedids.PlayerID(21), seedids.PlayerID(23), seedids.PlayerID(24),
		seedids.PlayerID(22), seedids.PlayerID(25), seedids.PlayerID(26), seedids.PlayerID(27), seedids.PlayerID(28),
		seedids.TonklaID, seedids.PlayerID(29), seedids.PlayerID(30), seedids.PlayerID(31),
		seedids.BeamID, seedids.PlayerID(32),
	}
	ranks := []string{"ascendant", "immortal", "radiant"}
	for i, acc := range b3Accounts {
		responses = append(responses, models.FormResponse{
			TournamentID: seedids.TournRealRegValorant,
			AccountID:    acc,
			Answers: []models.FieldAnswer{
				{FieldID: "riot_id", Value: fmt.Sprintf("SEAChal%02d#SEA", i+1)},
				{FieldID: "peak_rank", Value: ranks[i%len(ranks)]},
				{FieldID: "ping_ms", Value: 18 + (i%7)*3},
				{FieldID: "roles", Value: roleChoices[i%len(roleChoices)]},
				{FieldID: "team_bio", Value: "Top 16 qualifier participant in SEA Premier Division."},
			},
		})
	}

	// B4 accounts (SF6 Solo; Alex omitted intentionally)
	b4Accounts := []uuid.UUID{
		seedids.KaitoID, seedids.PlengID, seedids.TonklaID, seedids.BeamID,
		seedids.FahID, seedids.MekID,
		seedids.PlayerID(1), seedids.PlayerID(2), seedids.PlayerID(3),
		seedids.PlayerID(4), seedids.PlayerID(5), seedids.PlayerID(6),
		seedids.PlayerID(7), seedids.PlayerID(8), seedids.PlayerID(9),
	}
	chars := []string{"luke", "ken", "cammy", "akuma", "chunli"}
	for i, acc := range b4Accounts {
		responses = append(responses, models.FormResponse{
			TournamentID: seedids.TournRealRegSF6Solo,
			AccountID:    acc,
			Answers: []models.FieldAnswer{
				{FieldID: "cfn_id", Value: fmt.Sprintf("TH_Clash_%02d", i+1)},
				{FieldID: "main_char", Value: chars[i%len(chars)]},
				{FieldID: "controller", Value: ctrlOpts[i%len(ctrlOpts)]},
				{FieldID: "seed_rating", Value: 4 + (i % 6)},
			},
		})
	}

	// Validate every form and answer at fixture construction time so bugs fail loudly.
	formByTournament := make(map[uuid.UUID]models.RegistrationForm, len(formDocs))
	for _, f := range formDocs {
		for _, field := range f.Fields {
			if err := registry.ValidateFieldConfig(field); err != nil {
				return nil, nil, fmt.Errorf("invalid seeded form config (%s/%s): %w", f.TournamentID, field.ID, err)
			}
		}
		formByTournament[f.TournamentID] = f
	}
	for _, r := range responses {
		f := formByTournament[r.TournamentID]
		if err := registry.ValidateFormAnswers(f.Fields, r.Answers); err != nil {
			return nil, nil, fmt.Errorf("invalid seeded form response (%s/%s): %w", r.TournamentID, r.AccountID, err)
		}
	}

	return formDocs, responses, nil
}
