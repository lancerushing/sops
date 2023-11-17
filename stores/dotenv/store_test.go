package dotenv

import (
	"strings"
	"testing"

	"github.com/getsops/sops/v3"
	"github.com/stretchr/testify/assert"
)

var PLAIN = []byte(strings.TrimLeft(`
VAR1=val1
VAR2=val2
#comment
VAR3_unencrypted=val3
VAR4=val4\nval4
`, "\n"))

var BRANCH = sops.TreeBranch{
	sops.TreeItem{
		Key:   "VAR1",
		Value: "val1",
	},
	sops.TreeItem{
		Key:   "VAR2",
		Value: "val2",
	},
	sops.TreeItem{
		Key:   sops.Comment{Value: "comment"},
		Value: nil,
	},
	sops.TreeItem{
		Key:   "VAR3_unencrypted",
		Value: "val3",
	},
	sops.TreeItem{
		Key:   "VAR4",
		Value: "val4\nval4",
	},
}

func TestLoadPlainFile(t *testing.T) {
	branches, err := (&Store{}).LoadPlainFile(PLAIN)
	assert.Nil(t, err)
	assert.Equal(t, BRANCH, branches[0])
}
func TestEmitPlainFile(t *testing.T) {
	branches := sops.TreeBranches{
		BRANCH,
	}
	bytes, err := (&Store{}).EmitPlainFile(branches)
	assert.Nil(t, err)
	assert.Equal(t, PLAIN, bytes)
}

func TestEmitValueString(t *testing.T) {
	bytes, err := (&Store{}).EmitValue("hello")
	assert.Nil(t, err)
	assert.Equal(t, []byte("hello"), bytes)
}

func TestEmitValueNonstring(t *testing.T) {
	_, err := (&Store{}).EmitValue(BRANCH)
	assert.NotNil(t, err)
}

func TestEmitEncryptedFileStability(t *testing.T) {
	// emit the same tree multiple times to ensure the output is stable
	// i.e. emitting the same tree always yields exactly the same output
	var previous []byte
	for i := 0; i < 10; i += 1 {
		bytes, err := (&Store{}).EmitEncryptedFile(sops.Tree{
			Branches: []sops.TreeBranch{{}},
		})
		assert.Nil(t, err)
		assert.NotEmpty(t, bytes)
		if previous != nil {
			assert.Equal(t, previous, bytes)
		}
		previous = bytes
	}
}

// MACOnlyEncrypted metadata wants a bool
// But dotenv wants all strings
func TestCastForBoolMetaData(t *testing.T) {
	store := &Store{}
	bytes, err := store.EmitEncryptedFile(sops.Tree{
		Branches: []sops.TreeBranch{{}},
		Metadata: sops.Metadata{MACOnlyEncrypted: true},
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, bytes)

	_, err2 := store.LoadEncryptedFile(bytes)
	// "No keys found in file" err will occur after the mac-only-encrypted was loaded correctly
	assert.ErrorContains(t, err2, "No keys found in file")

}

func TestMapToMetadata(t *testing.T) {
	m1 := map[string]interface{}{
		"mac_only_encrypted": "false",
	}
	m2 := map[string]interface{}{
		"mac_only_encrypted": "true",
	}
	m3 := map[string]interface{}{
		"mac_only_encrypted": "bad-value",
	}

	metaData1, err1 := mapToMetadata(m1)
	assert.Nil(t, err1)
	assert.False(t, metaData1.MACOnlyEncrypted)

	metaData2, err2 := mapToMetadata(m2)
	assert.Nil(t, err2)
	assert.True(t, metaData2.MACOnlyEncrypted)

	_, err3 := mapToMetadata(m3)
	assert.ErrorContains(t, err3, "unrecognized value 'bad-value'")

}
