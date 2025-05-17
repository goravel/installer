package envfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/assert"
)

func TestReplaceValues(t *testing.T) {
	t.Run("replace failed", func(t *testing.T) {
		assert.ErrorIs(t, ReplaceValues("notfound", nil), os.ErrNotExist)
	})

	t.Run("replace success", func(t *testing.T) {
		envFile := filepath.Join(t.TempDir(), "test.env")
		assert.NoError(t, file.PutContent(envFile, "FOO=bar\n#comment\nBAZ=qux"))
		assert.True(t, file.Contain(envFile, "FOO=bar\n#comment\nBAZ=qux"))

		assert.NoError(t, ReplaceValues(envFile, map[string]string{
			"FOO": "newbar",
			"BAZ": "newqux",
			"NEW": "new value",
		}))

		assert.False(t, file.Contain(envFile, "FOO=bar\n#comment\nBAZ=qux"))
		assert.True(t, file.Contain(envFile, "FOO=newbar\n#comment\nBAZ=newqux\nNEW=\"new value\""))
	})

}
