package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/davecgh/go-spew/spew"
)

func Test_parseStatusHTML(t *testing.T) {
	require := require.New(t)
	content, err := ioutil.ReadFile("example_DocsisStatus.htm")
	require.NoError(err)

	res, err := parseStatusHTML(string(content))
	require.NoError(err)

	spew.Dump(res)

	require.Equal("OK", res.Init.ConnectivityStateStatus)
	require.EqualValues(38700000, res.Upstream[0].FrequencyHz)

}
