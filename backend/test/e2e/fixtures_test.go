package e2e

import (
	"context"
	"testing"
	"time"

	mongoRepo "github.com/holdennekt/sgame/backend/internal/infrastructure/database/mongo"
	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// insertTestPack inserts a minimal pack into the test Mongo database and returns its ID.
// The pack has:
//   - 1 round with 3 questions: Regular, CatInBag, Auction
//   - 1 final round with 2 categories
//
// This covers every RoomState in TestFullGame.
func insertTestPack(t *testing.T, mongoURI string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })

	repo := mongoRepo.NewPackRepository(client.Database("sgame_test"))

	text := "What is the answer?"
	pack := &domain.Pack{
		Name: "E2E Test Pack",
		Type: domain.Public,
		Rounds: []domain.Round{
			{
				Name: "Round 1",
				Categories: []domain.Category{
					{
						Name: "General",
						Questions: []domain.Question{
							{
								HiddenQuestion: domain.HiddenQuestion{
									Round: "Round 1", Category: "General", Index: 0, Value: 100,
								},
								Type:    domain.Regular,
								Text:    &text,
								Answers: []string{"42"},
							},
							{
								HiddenQuestion: domain.HiddenQuestion{
									Round: "Round 1", Category: "General", Index: 1, Value: 200,
								},
								Type:    domain.CatInBag,
								Text:    &text,
								Answers: []string{"Cat"},
							},
							{
								HiddenQuestion: domain.HiddenQuestion{
									Round: "Round 1", Category: "General", Index: 2, Value: 300,
								},
								Type:    domain.Auction,
								Text:    &text,
								Answers: []string{"Auction"},
							},
						},
					},
				},
			},
		},
		FinalRound: domain.FinalRound{
			Categories: []domain.FinalRoundCategory{
				{
					HiddenFinalRoundCategory: domain.HiddenFinalRoundCategory{Name: "Final A"},
					Question: domain.FinalRoundQuestion{
						HiddenFinalRoundQuestion: domain.HiddenFinalRoundQuestion{
							Category: "Final A",
							Text:     &text,
						},
						Answers: []string{"Final Answer A"},
					},
				},
				{
					HiddenFinalRoundCategory: domain.HiddenFinalRoundCategory{Name: "Final B"},
					Question: domain.FinalRoundQuestion{
						HiddenFinalRoundQuestion: domain.HiddenFinalRoundQuestion{
							Category: "Final B",
							Text:     &text,
						},
						Answers: []string{"Final Answer B"},
					},
				},
			},
		},
	}

	id, err := repo.Create(ctx, pack)
	require.NoError(t, err)
	return id
}
