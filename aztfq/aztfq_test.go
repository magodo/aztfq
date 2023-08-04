package aztfq

import (
	"encoding/json"
	"testing"

	"github.com/magodo/azure-rest-api-bridge/ctrl"
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

func TestBuildLookupTable(t *testing.T) {
	input := `{
	"azurerm_foo": {
	  "/p1": [{
	   	"addr": "properties.p1",
		"root_model": {
		  "path_ref": "foo.json#/paths/~1%7BresourceId%7D~1providers~1Microsoft.Foo~1foos~1%7BfooName%7D",
		  "version": "2020-01-01"
		}
	  }],
	  "/p2": [{
	   	"addr": "properties.p2",
		"root_model": {
		  "path_ref": "foo.json#/paths/~1%7BresourceId%7D~1providers~1Microsoft.Foo~1foos~1%7BfooName%7D~1settings~1%7BsettingName%7D",
		  "version": "2020-01-02"
		}
	  }]
	},
	"azurerm_bar": {
	  "/p1": [{
	   	"addr": "properties.p1",
		"root_model": {
		  "path_ref": "bar.json#/paths/~1%7BresourceId%7D~1providers~1Microsoft.Bar~1bars~1%7BbarName%7D",
		  "version": "2020-02-01"
		}
	  }],
	  "/p2": [{
	   	"addr": "properties.p2",
		"root_model": {
		  "path_ref": "bar.json#/paths/~1%7BresourceId%7D~1providers~1Microsoft.Bar~1bars~1%7BbarName%7D~1settings~1%7BsettingName%7D",
		  "version": "2020-02-02"
		}
	  }]
	}
}`
	var output map[string]ctrl.ModelMap
	require.NoError(t, json.Unmarshal([]byte(input), &output))
	table, err := buildLookupTable(output)
	require.NoError(t, err)
	require.Equal(t, LookupTable{
		"MICROSOFT.FOO/FOOS": map[string]map[string][]TFResult{
			"": {
				"properties.p1": {
					{
						ResourceType: "azurerm_foo",
						PropertyAddr: "/p1",
					},
				},
			},
			"2020-01-01": {
				"properties.p1": {
					{
						ResourceType: "azurerm_foo",
						PropertyAddr: "/p1",
					},
				},
			},
		},
		"MICROSOFT.FOO/FOOS/SETTINGS": map[string]map[string][]TFResult{
			"": {
				"properties.p2": {
					{
						ResourceType: "azurerm_foo",
						PropertyAddr: "/p2",
					},
				},
			},
			"2020-01-02": {
				"properties.p2": {
					{
						ResourceType: "azurerm_foo",
						PropertyAddr: "/p2",
					},
				},
			},
		},
		"MICROSOFT.BAR/BARS": map[string]map[string][]TFResult{
			"": {
				"properties.p1": {
					{
						ResourceType: "azurerm_bar",
						PropertyAddr: "/p1",
					},
				},
			},
			"2020-02-01": {
				"properties.p1": {
					{
						ResourceType: "azurerm_bar",
						PropertyAddr: "/p1",
					},
				},
			},
		},
		"MICROSOFT.BAR/BARS/SETTINGS": map[string]map[string][]TFResult{
			"": {
				"properties.p2": {
					{
						ResourceType: "azurerm_bar",
						PropertyAddr: "/p2",
					},
				},
			},
			"2020-02-02": {
				"properties.p2": {
					{
						ResourceType: "azurerm_bar",
						PropertyAddr: "/p2",
					},
				},
			},
		},
	}, table)
}
