package redis_test

import (
	"context"
	"testing"

	"github.com/hextec/yolyolakay-backend/internal/auth/repository/redis"
	redisClient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestOTPRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	host, err := redisContainer.Host(ctx)
	require.NoError(t, err)

	port, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	addr := host + ":" + port.Port()

	client := redisClient.NewClient(&redisClient.Options{
		Addr: addr,
	})
	defer client.Close()

	repo := redis.NewOTPRepository(client)

	t.Run("Save and Verify OTP", func(t *testing.T) {
		phone := "+998901234567"
		code := "123456"

		err := repo.SaveOTP(ctx, phone, code)
		require.NoError(t, err)

		// Verify should succeed and then delete it
		ok, err := repo.VerifyOTP(ctx, phone, code)
		require.NoError(t, err)
		assert.True(t, ok)

		// Verify again should fail because it was deleted
		ok, err = repo.VerifyOTP(ctx, phone, code)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("Verify Wrong OTP", func(t *testing.T) {
		phone := "+998901111111"
		code := "654321"

		err := repo.SaveOTP(ctx, phone, code)
		require.NoError(t, err)

		ok, err := repo.VerifyOTP(ctx, phone, "111111")
		require.NoError(t, err)
		assert.False(t, ok)
	})
}
