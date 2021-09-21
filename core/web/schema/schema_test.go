package schema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Schema(t *testing.T) {
	s, err := GetRootSchema()
	require.NoError(t, err)

	fmt.Println(s)

	t.Fail()
}
