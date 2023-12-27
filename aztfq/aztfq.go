package aztfq

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/magodo/azure-rest-api-bridge/ctrl"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger"
)

type Option struct {
	// ImplicitArrayIndex makes the array index to be implicit, e.g. "foos.*.id" -> "foos.id".
	ImplicitArrayIndex bool
}

func BuildLookupTable(input []byte, opt *Option) (LookupTable, error) {
	if opt == nil {
		opt = &Option{
			ImplicitArrayIndex: false,
		}
	}

	var output map[string]ctrl.ModelMap
	if err := json.Unmarshal(input, &output); err != nil {
		return LookupTable{}, err
	}
	return buildLookupTable(output, opt)
}

func buildLookupTable(output map[string]ctrl.ModelMap, opt *Option) (LookupTable, error) {
	t := LookupTable{}
	for tfRT, mm := range output {
		for tfPropAddr, apiPoses := range mm {
			for _, apiPos := range apiPoses {
				jsptr := apiPos.RootModel.PathRef.GetPointer()
				if jsptr == nil {
					return nil, fmt.Errorf("nil JSON pointer for %s: %s", tfRT, tfPropAddr)
				}
				tks := jsptr.DecodedTokens()
				if len(tks) != 2 {
					return nil, fmt.Errorf("the length of JSON pointer for %s: %s expects to be 2, got=%d", tfRT, tfPropAddr, len(tks))
				}
				azureRT, ok := azureResourceTypeFromPath(tks[1])
				if !ok {
					continue
				}
				tt, ok := t[azureRT]
				if !ok {
					tt = map[string]map[string][]TFResult{}
					t[azureRT] = tt
				}

				apiVersion := apiPos.RootModel.Version
				ttt, ok := tt[apiVersion]
				if !ok {
					ttt = map[string][]TFResult{}
					tt[apiVersion] = ttt
				}
				tttAny, ok := tt[""]
				if !ok {
					tttAny = map[string][]TFResult{}
					tt[""] = tttAny
				}

				apiAddr := apiPos.Addr
				if opt.ImplicitArrayIndex {
					apiAddr = removeArrayIndex(apiAddr)
				}
				apiPropAddr := apiAddr.String()

				ttt[apiPropAddr] = append(ttt[apiPropAddr], TFResult{
					ResourceType: tfRT,
					PropertyAddr: tfPropAddr,
				})
				tttAny[apiPropAddr] = append(tttAny[apiPropAddr], TFResult{
					ResourceType: tfRT,
					PropertyAddr: tfPropAddr,
				})
			}
		}
	}

	for _, tt := range t {
		for _, ttt := range tt {
			for azurePropAddr, tfresults := range ttt {
				m := map[TFResult]bool{}
				var newTFResults []TFResult
				for _, tfr := range tfresults {
					if !m[tfr] {
						m[tfr] = true
						newTFResults = append(newTFResults, tfr)
					}
				}
				sort.Slice(newTFResults, func(i, j int) bool {
					r1, r2 := newTFResults[i], newTFResults[j]
					if r1.ResourceType != r2.ResourceType {
						return r1.ResourceType < r2.ResourceType
					}
					return r1.PropertyAddr < r2.PropertyAddr
				})
				ttt[azurePropAddr] = newTFResults
			}
		}
	}
	return t, nil
}

func removeArrayIndex(apiAddr swagger.PropertyAddr) swagger.PropertyAddr {
	newApiAddr := make(swagger.PropertyAddr, 0)
	for _, step := range apiAddr {
		if step.Type != swagger.PropertyAddrStepTypeIndex {
			newApiAddr = append(newApiAddr, step)
		}
	}

	return newApiAddr
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

	rt := strings.ToUpper(strings.Join(rts, "/"))
	return rt, rt != ""
}

type TFResult struct {
	ResourceType string
	PropertyAddr string
}

// LookupTable is the main lookup table used for querying.
// key1: Azure resource type in upper case (e.g. MICROSOFT.COMPUTE/VIRTUALMACHINES)
// key2: API version. Especially, there is always an empty string ("") key represents any api version.
// key3: Azure resource property address (e.g. properties.object.key, values.*.id)
type LookupTable map[string]map[string]map[string][]TFResult
