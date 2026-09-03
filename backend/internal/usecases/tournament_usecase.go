package usecases

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/moomaideng/eventory/internal/models"
	"github.com/moomaideng/eventory/internal/repositories"
)

var ErrInvalidTournamentFilters = errors.New("invalid tournament filters")

var catalogLocation = time.FixedZone("Asia/Bangkok", 7*60*60)

type SearchTournamentsInput struct {
	Query       string
	StartFrom   string
	StartTo     string
	MinEntryFee *int64
	MaxEntryFee *int64
	Status      string
	Sort        string
	Page        int
	PageSize    int
}

type TournamentSearchResult struct {
	Items    []models.Tournament
	Total    int64
	Page     int
	PageSize int
}

type TournamentUseCase struct {
	tournamentRepo repositories.TournamentRepository
}

func NewTournamentUseCase(tournamentRepo repositories.TournamentRepository) *TournamentUseCase {
	return &TournamentUseCase{tournamentRepo: tournamentRepo}
}

// validate query
func (u *TournamentUseCase) Search(
	ctx context.Context,
	input SearchTournamentsInput,
) (*TournamentSearchResult, error) {
	query := strings.TrimSpace(input.Query)
	if len(query) > 100 {
		return nil, ErrInvalidTournamentFilters
	}

	startFrom, err := parseCatalogDate(input.StartFrom, false)
	if err != nil {
		return nil, ErrInvalidTournamentFilters
	}
	startTo, err := parseCatalogDate(input.StartTo, true)
	if err != nil {
		return nil, ErrInvalidTournamentFilters
	}
	if startFrom != nil && startTo != nil && startFrom.After(*startTo) {
		return nil, ErrInvalidTournamentFilters
	}
	if input.MinEntryFee != nil && *input.MinEntryFee < 0 ||
		input.MaxEntryFee != nil && *input.MaxEntryFee < 0 ||
		input.MinEntryFee != nil && input.MaxEntryFee != nil && *input.MinEntryFee > *input.MaxEntryFee {
		return nil, ErrInvalidTournamentFilters
	}

	status := strings.ToUpper(strings.TrimSpace(input.Status))
	validStatuses := map[string]bool{
		"":                                      true,
		models.TournamentStatusRegistrationOpen: true,
		models.TournamentStatusRegistrationClosed: true,
		models.TournamentStatusOngoing:            true,
		models.TournamentStatusCompleted:          true,
	}
	if !validStatuses[status] {
		return nil, ErrInvalidTournamentFilters
	}

	sort := strings.ToLower(strings.TrimSpace(input.Sort))
	if sort == "" {
		sort = "start_asc"
	}
	validSorts := map[string]bool{
		"start_asc": true, "start_desc": true, "fee_asc": true, "fee_desc": true,
	}
	if !validSorts[sort] {
		return nil, ErrInvalidTournamentFilters
	}

	page := input.Page
	if page == 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 12
	}
	if page < 1 || pageSize < 1 || pageSize > 100 {
		return nil, ErrInvalidTournamentFilters
	}

	items, total, err := u.tournamentRepo.Search(ctx, repositories.TournamentFilters{
		Query: query, StartFrom: startFrom, StartTo: startTo,
		MinEntryFee: input.MinEntryFee, MaxEntryFee: input.MaxEntryFee,
		Status: status, Sort: sort, Page: page, PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	return &TournamentSearchResult{
		Items: items, Total: total, Page: page, PageSize: pageSize,
	}, nil
}

func parseCatalogDate(value string, endOfDay bool) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, catalogLocation)
	if err != nil {
		return nil, err
	}
	if endOfDay {
		parsed = parsed.Add(24*time.Hour - time.Nanosecond)
	}
	return &parsed, nil
}
