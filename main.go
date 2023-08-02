package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/magodo/azure-rest-api-bridge/ctrl"
)

//go:embed output.json
var OutputRaw []byte

func main() {
	rt := flag.String("rt", "", "Azure resource type (e.g. Microsoft.Compute/virtualMachines)")
	prop := flag.String("prop", "", "Azure property address (e.g. properties.osProfile.computerName)")
	flag.Parse()
	if err := realMain(*rt, *prop); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(rt, prop string) error {
	var output map[string]ctrl.ModelMap
	if err := json.Unmarshal(OutputRaw, &output); err != nil {
		return err
	}
	t := buildLookupTable(output)
	results := t.Lookup(rt, prop)
	fmt.Println(results)
	return nil
}

func buildLookupTable(output map[string]ctrl.ModelMap) LookupTable {
	t := LookupTable{}
	for tfRT, mm := range output {
		for tfPropAddr, apiPoses := range mm {
			for _, apiPos := range apiPoses {
				azureRT, ok := azureResourceTypeFromPath(apiPos.APIPath)
				if !ok {
					continue
				}
				tt, ok := t[azureRT]
				if !ok {
					tt = map[string][]TFResult{}
					t[azureRT] = tt
				}
				apiPropAddr := apiPos.Addr.String()
				tt[apiPropAddr] = append(tt[apiPropAddr], TFResult{
					ResourceType: tfRT,
					PropertyAddr: tfPropAddr,
				})
			}
		}
	}
	return t
}

func azureResourceTypeFromPath(path string) (string, bool) {
	idx := strings.LastIndex(path, "/providers/")
	if idx == -1 {
		return "", false
	}
	path = path[idx+1:]
	segs := strings.Split(path, "/")

	rtSegs := segs[2:]

	if len(rtSegs)%2 != 0 {
		return "", false
	}

	rts := []string{segs[1]}
	for i := 0; i < len(rtSegs); i += 2 {
		rts = append(rts, rtSegs[i])
	}

	return strings.ToUpper(strings.Join(rts, "/")), true
}

type TFResult struct {
	ResourceType string
	PropertyAddr string
}

// LookupTable is the main lookup table used for querying.
// key1: Azure resource type in upper case (e.g. MICROSOFT.COMPUTE/VIRTUALMACHINES)
// key2: Azure resource property address (e.g. properties.foo.bar)
type LookupTable map[string]map[string][]TFResult

func (t LookupTable) Lookup(rt, propAddr string) []TFResult {
	m, ok := t[strings.ToUpper(rt)]
	if !ok {
		return nil
	}
	return m[propAddr]
}
