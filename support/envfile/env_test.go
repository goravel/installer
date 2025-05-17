package envfile

import (
	"path/filepath"
	"testing"

	"github.com/gookit/goutil/testutil/assert"

	"github.com/goravel/framework/support/file"
)

func TestReplaceValues(t *testing.T) {
	t.Run("replace failed", func(t *testing.T) {
		assert.ErrSubMsg(t, ReplaceValues("notfound", nil), "no such file or directory")
	})

	t.Run("replace success", func(t *testing.T) {
		envFile := filepath.Join(t.TempDir(), "test.env")
		assert.NoErr(t, file.PutContent(envFile, "FOO=bar\n#comment\nBAZ=qux"))
		assert.True(t, file.Contain(envFile, "FOO=bar\n#comment\nBAZ=qux"))

		assert.NoErr(t, ReplaceValues(envFile, map[string]string{
			"FOO": "newbar",
			"BAZ": "newqux",
			"NEW": "new value",
		}))

		assert.False(t, file.Contain(envFile, "FOO=bar\n#comment\nBAZ=qux"))
		assert.True(t, file.Contain(envFile, "FOO=newbar\n#comment\nBAZ=newqux\nNEW=\"new value\""))
	})

}
