package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yaop-labs/apiary/internal/openapi"
	"github.com/yaop-labs/apiary/internal/parser"
	"gopkg.in/yaml.v3"
)

var toolVersion = "dev"

func main() {
	log.SetFlags(0)

	if len(os.Args) > 1 && os.Args[1] == "serve" {
		runServe(os.Args[2:])
		return
	}

	out := flag.String("out", "openapi.yaml", "output file path (use - for stdout)")
	title := flag.String("title", "API", "API title")
	version := flag.String("version", "0.0.1", "API version")
	description := flag.String("description", "", "API description (info.description)")
	security := flag.String("security", "", "comma-separated global security schemes, e.g. bearer or adminAuth:bearer")
	server := flag.String("server", "", "comma-separated server URLs for info.servers, e.g. https://api.example.com")
	format := flag.String("format", "", "output format: yaml or json (default: inferred from -out extension)")
	check := flag.Bool("check", false, "verify -out is up to date; exit non-zero if it differs (writes nothing)")
	changeDir := flag.String("C", "", "load packages as if run from this directory (output paths stay relative to the cwd)")
	showVersion := flag.Bool("V", false, "print apiary version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("apiary %s\n", toolVersion)
		return
	}

	set := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { set[f.Name] = true })
	cfg := loadConfig()

	patterns := flag.Args()
	if len(patterns) == 0 {
		patterns = cfg.Patterns
	}
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	opt := specOptions{
		title:       resolveStr(set["title"], *title, cfg.Title),
		version:     resolveStr(set["version"], *version, cfg.Version),
		description: resolveStr(set["description"], *description, cfg.Description),
		security:    resolveList(set["security"], splitCSV(*security), cfg.Security),
		servers:     resolveList(set["server"], splitCSV(*server), cfg.Servers),
		dir:         *changeDir,
	}

	spec, opCount, err := buildSpec(patterns, opt)
	if err != nil {
		log.Fatalf("apiary: %v", err)
	}
	if opCount == 0 {
		fmt.Fprintln(os.Stderr, "apiary: no operations found")
		os.Exit(1)
	}

	outPath := resolveStr(set["out"], *out, cfg.Out)
	data, err := encodeSpec(spec, chooseFormat(*format, cfg.Format, outPath))
	if err != nil {
		log.Fatalf("apiary: %v", err)
	}

	if *check {
		runCheck(outPath, data)
		return
	}

	if outPath == "-" {
		fmt.Print(string(data))
		return
	}

	if dir := filepath.Dir(outPath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("apiary: create output dir: %v", err)
		}
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		log.Fatalf("apiary: write file: %v", err)
	}
	schemaCount := 0
	if spec.Components != nil {
		schemaCount = len(spec.Components.Schemas)
	}
	fmt.Printf("apiary: wrote %s — %d operation(s) across %d path(s), %d schema(s)\n",
		outPath, opCount, len(spec.Paths), schemaCount)
}

type specOptions struct {
	title, version, description string
	security, servers           []string
	dir                         string
}

func buildSpec(patterns []string, opt specOptions) (*openapi.OpenAPI, int, error) {
	p := parser.New()
	if err := p.LoadDir(opt.dir, patterns...); err != nil {
		return nil, 0, err
	}
	ops := p.Operations()
	if len(ops) == 0 {
		return nil, 0, nil
	}
	builder := openapi.NewBuilder(opt.title, opt.version).WithDescription(opt.description)
	for _, s := range opt.security {
		builder.WithSecurity(s)
	}
	for _, s := range opt.servers {
		builder.WithServer(s)
	}
	spec, err := builder.Build(ops, p.Types(), p.Enums())
	if err != nil {
		return nil, 0, err
	}
	return spec, len(ops), nil
}

func encodeSpec(spec *openapi.OpenAPI, format string) ([]byte, error) {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal yaml: %w", err)
	}
	if format != "json" {
		return data, nil
	}
	var generic any
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return nil, fmt.Errorf("re-encode json: %w", err)
	}
	out, err := json.MarshalIndent(generic, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	return append(out, '\n'), nil
}

func chooseFormat(flagFmt, cfgFmt, outPath string) string {
	switch {
	case flagFmt != "":
		return flagFmt
	case cfgFmt != "":
		return cfgFmt
	case strings.HasSuffix(strings.ToLower(outPath), ".json"):
		return "json"
	default:
		return "yaml"
	}
}

func runCheck(outPath string, data []byte) {
	if outPath == "-" {
		log.Fatalf("apiary: -check cannot be used with stdout output")
	}
	existing, err := os.ReadFile(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "apiary: %s is missing or unreadable — run apiary to generate it\n", outPath)
		os.Exit(1)
	}
	if !bytes.Equal(existing, data) {
		fmt.Fprintf(os.Stderr, "apiary: %s is out of date — re-run apiary to regenerate it\n", outPath)
		os.Exit(1)
	}
	fmt.Printf("apiary: %s is up to date\n", outPath)
}

func resolveStr(flagSet bool, flagVal, cfgVal string) string {
	if flagSet {
		return flagVal
	}
	if cfgVal != "" {
		return cfgVal
	}
	return flagVal
}

func resolveList(flagSet bool, flagVal, cfgVal []string) []string {
	if flagSet {
		return flagVal
	}
	if len(cfgVal) > 0 {
		return cfgVal
	}
	return flagVal
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
