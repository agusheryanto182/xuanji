package usecase_test

import (
	"context"
	"testing"

	"github.com/agusheryanto182/ecommerce-monolith/internal/entity"
	"github.com/agusheryanto182/ecommerce-monolith/internal/usecase/product"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProductUseCase(t *testing.T) (*product.UseCase, *MockProductRepo, *MockInterface) {
	t.Helper()

	ctrl := gomock.NewController(t)

	repo := NewMockProductRepo(ctrl)
	l := NewMockInterface(ctrl)
	useCase := product.New(repo, l)

	return useCase, repo, l
}

func TestStore(t *testing.T) {
	t.Parallel()

	t.Run("store success", func(t *testing.T) {
		t.Parallel()

		uc, repo, _ := newProductUseCase(t)
		repo.EXPECT().Store(context.Background(), gomock.Any()).Return(nil)

		p, err := uc.Store(context.Background(), &entity.Product{
			Name:        "Test Product",
			Description: "This is a test product",
			Price:       99.99,
			Stock:       10,
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		require.NoError(t, err)
		assert.NotEmpty(t, p.ID)
		assert.Equal(t, "Test Product", p.Name)
		assert.Equal(t, "This is a test product", p.Description)
		assert.Equal(t, 99.99, p.Price)
		assert.Equal(t, 10, p.Stock)
	})

	t.Run("store failed", func(t *testing.T) {
		t.Parallel()

		uc, repo, logger := newProductUseCase(t)

		logger.EXPECT().
			Error(gomock.Any()).
			Times(1)

		repo.EXPECT().
			Store(gomock.Any(), gomock.Any()).
			Return(entity.ErrInvalidProductCreate)

		_, err := uc.Store(context.Background(), &entity.Product{
			Name:        "Test Product",
			Description: "This is a test product",
			Price:       999,
			Stock:       10,
		})

		require.ErrorIs(t, err, entity.ErrInvalidProductCreate)
	})
}
