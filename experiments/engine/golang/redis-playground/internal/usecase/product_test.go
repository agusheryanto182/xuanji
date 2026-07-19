package usecase_test

import (
	"context"
	"testing"

	"github.com/agusheryanto182/redis-playground/internal/entity"
	"github.com/agusheryanto182/redis-playground/internal/usecase/product"
	gomock "github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newProductUseCase(t *testing.T) (*product.UseCase, *MockProductRepo, *MockInterface) {
	t.Helper()

	ctrl := gomock.NewController(t)
	repo := NewMockProductRepo(ctrl)
	l := NewMockInterface(ctrl)
	useCase := product.New(repo, nil, l)

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

func TestGetByID(t *testing.T) {
	productID := uuid.New()
	t.Parallel()

	t.Run("get by id success", func(t *testing.T) {
		t.Parallel()

		uc, repo, _ := newProductUseCase(t)

		repo.EXPECT().
			GetByID(context.Background(), productID).
			Return(&entity.Product{
				Name:        "Test Product",
				Description: "This is a test product",
				Price:       999.99,
				Stock:       10,
			}, nil)

		p, err := uc.GetByID(context.Background(), productID)

		require.NoError(t, err)
		assert.Equal(t, "Test Product", p.Name)
		assert.Equal(t, "This is a test product", p.Description)
		assert.Equal(t, 999.99, p.Price)
		assert.Equal(t, 10, p.Stock)
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("full update success", func(t *testing.T) {
		t.Parallel()

		uc, repo, _ := newProductUseCase(t)

		repo.EXPECT().Update(context.Background(), gomock.Any()).Return(nil)

		updated, err := uc.Update(context.Background(), &entity.Product{
			ID:          uuid.New(),
			Name:        "Test Product Update",
			Description: "This is a test product update",
			Price:       999.98,
			Stock:       18,
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		require.NoError(t, err)
		assert.Equal(t, "Test Product Update", updated.Name)
		assert.Equal(t, "This is a test product update", updated.Description)
		assert.Equal(t, 999.98, updated.Price)
		assert.Equal(t, 18, updated.Stock)
	})

	t.Run("error - product not found", func(t *testing.T) {
		t.Parallel()

		uc, repo, _ := newProductUseCase(t)
		expectedError := entity.ErrProductNotFound

		repo.EXPECT().Update(context.Background(), gomock.Any()).Return(expectedError)

		updated, err := uc.Update(context.Background(), &entity.Product{
			ID:          uuid.New(),
			Name:        "Test Product Update",
			Description: "This is a test product update",
			Price:       999.98,
			Stock:       18,
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, expectedError)
		assert.Nil(t, updated)
	})
}

func TestGetAll(t *testing.T) {
	t.Parallel()

	t.Run("get success", func(t *testing.T) {
		t.Parallel()

		uc, repo, _ := newProductUseCase(t)

		repo.EXPECT().GetAll(context.Background(), 15, 0).Return([]*entity.Product{
			{
				Name:        "Test Product 1",
				Description: "This is a test product 1",
				Price:       999.99,
				Stock:       10,
			},
			{
				Name:        "Test Product 2",
				Description: "This is a test product 2",
				Price:       999.99,
				Stock:       10,
			},
			{
				Name:        "Test Product 3",
				Description: "This is a test product 3",
				Price:       999.99,
				Stock:       10,
			},
		}, nil)

		repo.EXPECT().CountProducts(context.Background()).Return(3, nil)

		products, meta, err := uc.GetAll(context.Background(), 15, 0)

		require.NoError(t, err)

		assert.Len(t, products, 3)
		assert.Equal(t, meta, 3)
		assert.Equal(t, "Test Product 1", products[0].Name)
		assert.Equal(t, "This is a test product 1", products[0].Description)
		assert.Equal(t, 999.99, products[0].Price)
		assert.Equal(t, 10, products[0].Stock)
	})
}
