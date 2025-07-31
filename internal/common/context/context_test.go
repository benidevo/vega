package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithRole(t *testing.T) {
	t.Run("should_add_role_to_context_when_role_provided", func(t *testing.T) {
		ctx := context.Background()
		role := "admin"

		newCtx := WithRole(ctx, role)

		require.NotNil(t, newCtx)
		assert.NotEqual(t, ctx, newCtx, "should return new context")

		// Verify role is in context
		value := newCtx.Value(UserRoleKey)
		assert.Equal(t, role, value)
	})

	t.Run("should_add_empty_role_to_context_when_empty_string_provided", func(t *testing.T) {
		ctx := context.Background()
		role := ""

		newCtx := WithRole(ctx, role)

		require.NotNil(t, newCtx)
		value := newCtx.Value(UserRoleKey)
		assert.Equal(t, "", value)
	})

	t.Run("should_override_existing_role_when_called_multiple_times", func(t *testing.T) {
		ctx := context.Background()

		ctx1 := WithRole(ctx, "user")
		ctx2 := WithRole(ctx1, "admin")

		value1 := ctx1.Value(UserRoleKey)
		value2 := ctx2.Value(UserRoleKey)

		assert.Equal(t, "user", value1)
		assert.Equal(t, "admin", value2)
	})

	t.Run("should_preserve_other_context_values_when_adding_role", func(t *testing.T) {
		type otherKey string
		const testKey otherKey = "testKey"

		ctx := context.Background()
		ctx = context.WithValue(ctx, testKey, "testValue")

		newCtx := WithRole(ctx, "admin")

		// Check both values are present
		assert.Equal(t, "testValue", newCtx.Value(testKey))
		assert.Equal(t, "admin", newCtx.Value(UserRoleKey))
	})
}

func TestGetRole(t *testing.T) {
	t.Run("should_return_role_and_true_when_role_exists", func(t *testing.T) {
		ctx := context.Background()
		expectedRole := "admin"
		ctx = WithRole(ctx, expectedRole)

		role, ok := GetRole(ctx)

		assert.True(t, ok)
		assert.Equal(t, expectedRole, role)
	})

	t.Run("should_return_empty_string_and_false_when_role_not_exists", func(t *testing.T) {
		ctx := context.Background()

		role, ok := GetRole(ctx)

		assert.False(t, ok)
		assert.Equal(t, "", role)
	})

	t.Run("should_return_empty_string_and_true_when_empty_role_exists", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithRole(ctx, "")

		role, ok := GetRole(ctx)

		assert.True(t, ok)
		assert.Equal(t, "", role)
	})

	t.Run("should_return_empty_string_and_false_when_value_is_not_string", func(t *testing.T) {
		ctx := context.Background()
		// Directly set a non-string value
		ctx = context.WithValue(ctx, UserRoleKey, 123)

		role, ok := GetRole(ctx)

		assert.False(t, ok)
		assert.Equal(t, "", role)
	})

	// Note: GetRole with nil context will panic, which is expected Go behavior
	// Context should never be nil in practice
}

func TestContextKeyUniqueness(t *testing.T) {
	t.Run("should_not_collide_with_string_key_of_same_value", func(t *testing.T) {
		ctx := context.Background()

		// Add value with string key
		ctx = context.WithValue(ctx, "userRole", "string-value")

		// Add value with contextKey
		ctx = WithRole(ctx, "context-value")

		// Verify they don't collide
		stringValue := ctx.Value("userRole")
		contextValue, ok := GetRole(ctx)

		assert.Equal(t, "string-value", stringValue)
		assert.True(t, ok)
		assert.Equal(t, "context-value", contextValue)
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("should_handle_concurrent_reads_safely", func(t *testing.T) {
		ctx := WithRole(context.Background(), "admin")
		done := make(chan bool, 10)

		// Launch multiple goroutines to read concurrently
		for i := 0; i < 10; i++ {
			go func() {
				role, ok := GetRole(ctx)
				assert.True(t, ok)
				assert.Equal(t, "admin", role)
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func BenchmarkWithRole(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = WithRole(ctx, "admin")
	}
}

func BenchmarkGetRole(b *testing.B) {
	ctx := WithRole(context.Background(), "admin")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = GetRole(ctx)
	}
}
