package usecase_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/hextec/yolyolakay-backend/internal/notification/domain"
	"github.com/hextec/yolyolakay-backend/internal/notification/usecase"
	userPb "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
	notifMocks "github.com/hextec/yolyolakay-backend/pkg/mocks/notification"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type MockUserServiceClient struct {
	userPb.UserServiceClient
}

func TestNotificationUseCase_SendMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := notifMocks.NewMockNotificationRepository(ctrl)
	tokenRepo := notifMocks.NewMockDeviceTokenRepository(ctrl)

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	var userClient userPb.UserServiceClient // can be nil for SendMessage if not used

	uc := usecase.NewNotificationUseCase(repo, tokenRepo, redisClient, nil, nil, userClient)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		senderID := uuid.New()
		receiverID := uuid.New()
		rideID := uuid.New()
		content := "Hello"

		// Since SendMessage doesn't call repo.CreateMessage but saves asynchronously, we only check token repo
		tokenRepo.EXPECT().GetDeviceTokens(ctx, receiverID).Return([]*domain.DeviceToken{}, nil)

		replyToID := uuid.New()
		msg, err := uc.SendMessage(ctx, senderID, receiverID, &rideID, &replyToID, content)
		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, content, msg.Content)
		assert.Equal(t, &replyToID, msg.ReplyToID)
		assert.False(t, msg.CreatedAt.IsZero(), "CreatedAt should not be zero")
		assert.False(t, msg.UpdatedAt.IsZero(), "UpdatedAt should not be zero")
	})
}
