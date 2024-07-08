package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/magodo/aztfq/aztfq"
)

func main() {
	input := flag.String("i", "", "The output file of azure-rest-api-bridge")
	rt := flag.String("rt", "", "Azure resource type (e.g. Microsoft.Compute/virtualMachines)")
	prop := flag.String("prop", "", "Azure property address (e.g. properties/osProfile/computerName)")
	version := flag.String("version", "", "Azure API version")
	flag.Parse()
	if err := realMain(*input, *rt, *version, *prop); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(input, rt, version, prop string) error {
	b, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	t, err := aztfq.BuildLookupTable(b, nil)
	if err != nil {
		return err
	}
	if tt, ok := t[strings.ToUpper(rt)]; ok {
		if ttt, ok := tt[version]; ok {
			if results, ok := ttt[prop]; ok {
				fmt.Println(results)
			}
		}
	}
	return nil
}
