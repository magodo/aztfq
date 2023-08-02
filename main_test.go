package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAzureResourceTypeFromPath(t *testing.T) {
	cases := []struct {
		path     string
		expectRT string
		expectOK bool
	}{
		{
			path:     "/providers/Foo.Bar",
			expectRT: "FOO.BAR",
			expectOK: true,
		},
		{
			path:     "/providers/Foo.Bar/foos",
			expectOK: false,
		},
		{
			path:     "/providers/Foo.Bar/foos/{fooName}",
			expectRT: "FOO.BAR/FOOS",
			expectOK: true,
		},
		{
			path:     "/providers/Foo.Bar/foos/{fooName}/bars/{barName}",
			expectRT: "FOO.BAR/FOOS/BARS",
			expectOK: true,
		},
		{
			path:     "/providers/ABABAB/ababas/{ababaName}/providers/Foo.Bar/foos/{fooName}",
			expectRT: "FOO.BAR/FOOS",
			expectOK: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.path, func(t *testing.T) {
			rt, ok := azureResourceTypeFromPath(tt.path)
			require.Equal(t, tt.expectOK, ok)
			if ok {
				require.Equal(t, tt.expectRT, rt)
			}
		})
	}
}
