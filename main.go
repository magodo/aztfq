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
	rt := flag.String("rt", "", "Azure resource type (e.g. Microsoft.Compute/virtualMachines)")
	prop := flag.String("prop", "", "Azure property address (e.g. properties.osProfile.computerName)")
	version := flag.String("version", "", "Azure API version")
	flag.Parse()
	if err := realMain(*rt, *version, *prop); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(rt, version, prop string) error {
	t, err := aztfq.BuildLookupTable()
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
