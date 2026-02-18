// internal/usecase/wise_phrase.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"github.com/hensybex/soulwi_go_back/internal/model"
	"github.com/hensybex/soulwi_go_back/internal/repository"

	sim "github.com/xrash/smetrics"
)

const (
	maxRecentlyViewed = 50 // or whatever
	likeWeightFactor  = 0.1
)

// In-memory store for recently viewed phrases
type recentlyViewedStore struct {
	sync.RWMutex
	store map[string]map[uint]struct{} // map[userUID]map[phraseID]struct{}
}

func newRecentlyViewedStore() *recentlyViewedStore {
	return &recentlyViewedStore{
		store: make(map[string]map[uint]struct{}),
	}
}

func (s *recentlyViewedStore) get(userUID string) map[uint]struct{} {
	s.RLock()
	defer s.RUnlock()
	seen := make(map[uint]struct{})
	if userSeen, ok := s.store[userUID]; ok {
		for id := range userSeen {
			seen[id] = struct{}{}
		}
	}
	return seen
}

func (s *recentlyViewedStore) add(userUID string, phraseID uint) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.store[userUID]; !ok {
		s.store[userUID] = make(map[uint]struct{})
	}
	s.store[userUID][phraseID] = struct{}{}

	if len(s.store[userUID]) > maxRecentlyViewed {
		// Simple trim strategy: delete a "random" key.
		for idToDelete := range s.store[userUID] {
			delete(s.store[userUID], idToDelete)
			break // delete only one
		}
	}
}

func (s *recentlyViewedStore) clear(userUID string) {
	s.Lock()
	defer s.Unlock()
	delete(s.store, userUID)
}

type WisePhraseUsecase interface {
	GenerateBatch(ctx context.Context, promptIDOrText string, count int) error
	GetRandomPhrase(ctx context.Context, userUID string) (*model.WisePhrase, error)
	ToggleLikePhrase(ctx context.Context, userUID string, phraseID uint) error
	ListUserLikes(ctx context.Context, userUID string) ([]model.LikedPhraseResponse, error)
	RecordShare(ctx context.Context, userUID string, phraseID uint) error // <-- ADD THIS
	ListAdminPhrases(ctx context.Context, limit, offset int) ([]model.WisePhrase, int64, error)
	DeleteWisePhrase(ctx context.Context, phraseID uint) error
}

type wisePhraseUsecase struct {
	wpRepo         repository.WisePhraseRepository
	wpLikeRepo     repository.WisePhraseLikeRepository
	wpShareRepo    repository.WisePhraseShareRepository // <-- ADD THIS
	promptRepo     repository.PromptRepository
	aiUsecase      AIUsecase
	recentlyViewed *recentlyViewedStore
}

func NewWisePhraseUsecase(
	wpRepo repository.WisePhraseRepository,
	wpLikeRepo repository.WisePhraseLikeRepository,
	wpShareRepo repository.WisePhraseShareRepository, // <-- ADD THIS
	promptRepo repository.PromptRepository,
	aiUsecase AIUsecase,
) WisePhraseUsecase {
	return &wisePhraseUsecase{
		wpRepo:         wpRepo,
		wpLikeRepo:     wpLikeRepo,
		wpShareRepo:    wpShareRepo, // <-- ADD THIS
		promptRepo:     promptRepo,
		aiUsecase:      aiUsecase,
		recentlyViewed: newRecentlyViewedStore(),
	}
}

func (uc *wisePhraseUsecase) GenerateBatch(ctx context.Context, promptIDOrText string, count int) error {
	prompt, err := uc.promptRepo.GetByName(ctx, promptIDOrText)
	if err != nil {
		return fmt.Errorf("failed to fetch prompt by name: %w", err)
	}

	existing, err := uc.wpRepo.ListAll(ctx)
	if err != nil {
		return err
	}

	existingTexts := make([]string, 0, len(existing))
	for _, phrase := range existing {
		existingTexts = append(existingTexts, phrase.Text)
	}

	generatedTexts, err := uc.aiUsecase.GenerateWisePhrases(ctx, prompt.ID, count, existingTexts)
	if err != nil {
		return err
	}

	finalBatch := []model.WisePhrase{}
	existingForSimilarity := make([]model.WisePhrase, len(existing))
	copy(existingForSimilarity, existing)

	for _, candidate := range generatedTexts {
		if isTooSimilar(candidate, existingForSimilarity) {
			continue
		}
		phrase := model.WisePhrase{Text: candidate}
		finalBatch = append(finalBatch, phrase)
		existingForSimilarity = append(existingForSimilarity, phrase)
	}

	if len(finalBatch) == 0 {
		return errors.New("all generated phrases were too similar, none added")
	}

	if err := uc.wpRepo.CreateBatch(ctx, finalBatch); err != nil {
		return err
	}

	return nil
}

func isTooSimilar(candidate string, existing []model.WisePhrase) bool {
	candidateClean := strings.ToLower(strings.TrimSpace(candidate))
	for _, ex := range existing {
		exClean := strings.ToLower(strings.TrimSpace(ex.Text))
		if candidateClean == exClean {
			return true
		}
		dist := sim.WagnerFischer(candidateClean, exClean, 1, 1, 2)
		maxLen := float64(len(candidateClean))
		if maxLen == 0 {
			return true
		}
		simRatio := 1 - (float64(dist) / maxLen)
		if simRatio > 0.8 {
			return true
		}
	}
	return false
}

func (uc *wisePhraseUsecase) GetRandomPhrase(ctx context.Context, userUID string) (*model.WisePhrase, error) {
	all, err := uc.wpRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, errors.New("no wise phrases in DB")
	}

	seenIDs := uc.recentlyViewed.get(userUID)
	candidates := make([]model.WisePhrase, 0)
	for _, p := range all {
		if _, seen := seenIDs[p.ID]; !seen {
			candidates = append(candidates, p)
		}
	}

	if len(candidates) == 0 {
		uc.recentlyViewed.clear(userUID)
		candidates = all
	}

	var totalWeight float64
	weights := make([]float64, len(candidates))
	for i, c := range candidates {
		weight := 1.0 + likeWeightFactor*float64(c.LikeCount)
		weights[i] = weight
		totalWeight += weight
	}

	r := rand.Float64() * totalWeight

	var chosen model.WisePhrase
	if len(candidates) > 0 {
		chosen = candidates[0] // Default choice
	}

	for i, c := range candidates {
		if r < weights[i] {
			chosen = c
			break
		}
		r -= weights[i]
	}

	if chosen.ID != 0 {
		uc.recentlyViewed.add(userUID, chosen.ID)
	}

	return &chosen, nil
}

func (uc *wisePhraseUsecase) ToggleLikePhrase(ctx context.Context, userUID string, phraseID uint) error {
	exists, err := uc.wpLikeRepo.IsAlreadyLiked(ctx, userUID, phraseID)
	if err != nil {
		return err
	}

	if exists {
		if err := uc.wpLikeRepo.DeleteLike(ctx, userUID, phraseID); err != nil {
			return err
		}
		return uc.wpRepo.DecrementLikeCount(ctx, phraseID)
	} else {
		if err := uc.wpLikeRepo.CreateLike(ctx, userUID, phraseID); err != nil {
			return err
		}
		return uc.wpRepo.IncrementLikeCount(ctx, phraseID)
	}
}

func (uc *wisePhraseUsecase) ListUserLikes(ctx context.Context, userUID string) ([]model.LikedPhraseResponse, error) {
	return uc.wpLikeRepo.ListLikedPhrases(ctx, userUID)
}

func (uc *wisePhraseUsecase) RecordShare(ctx context.Context, userUID string, phraseID uint) error {
	// We record the share event for analytics, even if the user shares multiple times.
	if err := uc.wpShareRepo.CreateShare(ctx, userUID, phraseID); err != nil {
		return err
	}
	// And we increment the public counter.
	return uc.wpRepo.IncrementShareCount(ctx, phraseID)
}

func (uc *wisePhraseUsecase) ListAdminPhrases(ctx context.Context, limit, offset int) ([]model.WisePhrase, int64, error) {
	total, err := uc.wpRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	phrases, err := uc.wpRepo.ListPaged(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return phrases, total, nil
}

func (uc *wisePhraseUsecase) DeleteWisePhrase(ctx context.Context, phraseID uint) error {
	if err := uc.wpLikeRepo.DeleteByPhraseID(ctx, phraseID); err != nil {
		return err
	}
	if err := uc.wpShareRepo.DeleteByPhraseID(ctx, phraseID); err != nil {
		return err
	}
	return uc.wpRepo.DeleteByID(ctx, phraseID)
}
