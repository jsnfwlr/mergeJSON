package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	merger "github.com/wundergraph/graphql-go-tools/v2/pkg/astjson"
)

func main() {
	Execute()
}

type jsonFile struct {
	name     string
	contents []byte
}

func doMerge(path, base, output string) {
	var start jsonFile
	var src []jsonFile

	df, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, f := range df {
		switch {
		case f.IsDir():
			continue
		case f.Name() == output:
			continue
		case f.Name() == base:
			c, err := os.ReadFile(filepath.Join(path, f.Name()))
			if err != nil {
				panic(err)
			}
			start = jsonFile{name: f.Name(), contents: c}
		case strings.HasSuffix(f.Name(), ".json"):
			c, err := os.ReadFile(filepath.Join(path, f.Name()))
			if err != nil {
				panic(err)
			}

			src = append(src, jsonFile{name: f.Name(), contents: c})
		}
	}

	merged := &merger.JSON{}
	err = merged.ParseObject(start.contents)
	if err != nil {
		fmt.Printf("could not parse json: %v\n%s\n", err, start.contents)
		os.Exit(1)
	}

	out := &bytes.Buffer{}
	_ = merged.PrintNode(merged.Nodes[merged.RootNode], out)

	for _, s := range src {

		add, err := merged.AppendObject(s.contents)
		if err != nil {
			fmt.Printf("%s: '%s'\n", s.name, string(s.contents))
			panic(err)
		}

		_ = merged.MergeNodes(merged.RootNode, add)

		out = &bytes.Buffer{}
		_ = merged.PrintNode(merged.Nodes[merged.RootNode], out)
	}

	out = &bytes.Buffer{}
	_ = merged.PrintNode(merged.Nodes[merged.RootNode], out)

	_ = os.WriteFile(filepath.Join(path, output), out.Bytes(), 0o644)
}

var (
	version  = "dev"
	mergeCmd = &cobra.Command{
		Use:     "mergejson -b <base> -p <path> -o <output>",
		Short:   "",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			path, _ := cmd.Flags().GetString("path")
			base, _ := cmd.Flags().GetString("base")
			output, _ := cmd.Flags().GetString("output")

			doMerge(path, base, output)
		},
	}
)

func Execute() {
	err := mergeCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	mergeCmd.Flags().StringP("base", "b", "", "the name of the file to use as the base for the merge")
	mergeCmd.Flags().StringP("path", "p", "", "path to a directory containing the files to merge")
	mergeCmd.Flags().StringP("output", "o", "", "the name of the file to write the merged json to")
}
