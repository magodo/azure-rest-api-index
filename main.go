package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/magodo/azure-rest-api-index/azidx"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"text/scanner"

	"github.com/go-openapi/jsonpointer"
	"github.com/hashicorp/go-hclog"
	"github.com/urfave/cli/v2"
)

var (
	flagVerbose bool

	flagOutput string
	flagDedup  string

	flagIndex  string
	flagMethod string
	flagURL    string
)

func main() {
	app := &cli.App{
		Name:      "azure-rest-api-index",
		Version:   getVersion(),
		Usage:     "Index of azure-rest-api-specs",
		UsageText: "azure-rest-api-index <command> [option]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       `Show debug logs`,
				Destination: &flagVerbose,
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "build",
				Usage:     `Building the index`,
				UsageText: "azure-rest-api-index build [option] <specdir>",
				Before: func(ctx *cli.Context) error {
					initLogger()
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "output",
						Aliases:     []string{"o"},
						Usage:       `Output file`,
						Destination: &flagOutput,
					},
					&cli.StringFlag{
						Name:        "dedup",
						Usage:       `Deduplicate file`,
						Destination: &flagDedup,
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("The swagger spec dir not specified")
					}
					if c.NArg() > 1 {
						return fmt.Errorf("More than one arguments specified")
					}
					specdir := c.Args().First()
					index, err := azidx.BuildIndex(specdir, flagDedup)
					if err != nil {
						return err
					}
					b, err := json.MarshalIndent(index, "", "  ")
					if err != nil {
						log.Fatal(err)
					}
					if flagOutput == "" {
						fmt.Println(string(b))
						return nil
					}
					return os.WriteFile(flagOutput, b, 0644)
				},
			},
			{
				Name:      "lookup",
				Usage:     `Lookup a request's swagger definition based on the index`,
				UsageText: "azure-rest-api-index lookup [option]",
				Before: func(ctx *cli.Context) error {
					initLogger()
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "index",
						Usage:       `Use the pre-built index file by the "build" subcommand`,
						Destination: &flagIndex,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "method",
						Usage:       `The request method (e.g. GET)`,
						Destination: &flagMethod,
						Required:    true,
					},
					&cli.StringFlag{
						Name:        "url",
						Usage:       `The request URL`,
						Destination: &flagURL,
						Required:    true,
					},
				},
				Action: func(c *cli.Context) error {
					b, err := os.ReadFile(flagIndex)
					if err != nil {
						return fmt.Errorf("reading index file %s: %v", flagIndex, err)
					}
					var index azidx.Index
					if err := json.Unmarshal(b, &index); err != nil {
						return fmt.Errorf("unmarshal index file: %v", err)
					}
					uRL, err := url.Parse(flagURL)
					if err != nil {
						return fmt.Errorf("parsing URL %s: %v", flagURL, err)
					}
					ref, err := index.Lookup(flagMethod, *uRL)
					if err != nil {
						return err
					}
					pointerTokens := ref.GetPointer().DecodedTokens()
					out := fmt.Sprintf(`
Path	: %s
Method  : %s
File	: %s
`, ref.GetURL().Path, pointerTokens[1], pointerTokens[2])
					if index.Commit != "" {
						link, err := buildGithubLink(*ref.GetPointer(), index.Commit, index.Specdir, ref.GetURL().Path)
						if err != nil {
							return err
						}
						out += "Link    : " + link + "\n"
					}
					out += "Raw     : " + ref.String()
					fmt.Println(out)
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initLogger() {
	lvl := "INFO"
	if flagVerbose {
		lvl = "DEBUG"
	}
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "azure-rest-api-index",
		Level: hclog.LevelFromString(lvl),
		Color: hclog.AutoColor,
	})
	azidx.SetLogger(logger)
}

func buildGithubLink(ptr jsonpointer.Pointer, commit, specdir, fpath string) (string, error) {
	b, err := os.ReadFile(fpath)
	if err != nil {
		return "", err
	}
	offset, err := azidx.JSONPointerOffset(ptr, string(b))
	if err != nil {
		return "", err
	}
	var sc scanner.Scanner
	sc.Init(bytes.NewBuffer(b))
	fmt.Println(offset)
	for i := 0; i < int(offset); i++ {
		sc.Next()
	}
	pos := sc.Pos()

	relFile, err := filepath.Rel(specdir, fpath)
	if err != nil {
		return "", err
	}
	return "https://github.com/Azure/azure-rest-api-specs/blob/" + commit + "/specification/" + relFile + "#L" + strconv.Itoa(pos.Line), nil
}
