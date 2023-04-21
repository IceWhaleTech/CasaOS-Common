package jwt_test

import (
	"encoding/base64"
	"testing"

	"github.com/IceWhaleTech/CasaOS-Common/utils/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestGenerateSecret(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("Test valid secret generation", func(t *testing.T) {
		secret, err := jwt.GenerateSecret()
		require.NoError(t, err, "Expected no error for valid length")

		decoded, err := base64.StdEncoding.DecodeString(secret)
		require.NoError(t, err, "Expected no error decoding base64 secret")

		assert.Equal(t, 6, len(decoded), "Expected the decoded secret to have the specified length")
	})

	t.Run("Test secret randomness", func(t *testing.T) {
		secret1, err := jwt.GenerateSecret()
		require.NoError(t, err, "Expected no error generating first secret")

		secret2, err := jwt.GenerateSecret()
		require.NoError(t, err, "Expected no error generating second secret")

		assert.NotEqual(t, secret1, secret2, "Expected two generated secrets to be different")
	})
}
