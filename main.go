package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nuclide-research/chromascan/internal/scanner"
)

const usage = `chromascan -- ChromaDB unauth enumeration tool
nuclide-research.com

usage: chromascan [flags] <target-url>

  target-url  http://host:8000 (scheme required)

flags:
  --probe              Fingerprint only: version, auth, collection count
  --hunt               Full enumeration (default): fingerprint + collections + counts + samples + PII
  --write-canary       Inject write canary (also requires --authorize-write)
  --authorize-write    Explicit authorization gate for write operations
  --tenant <name>      ChromaDB tenant (default: default_tenant)
  --database <name>    ChromaDB database (default: default_database)
  --timeout <sec>      HTTP timeout (default: 10)
  -o <file>            Write JSON output to file
  -v                   Verbose: log HTTP exchanges

examples:
  chromascan http://target:8000
  chromascan http://target:8000 --probe
  chromascan http://target:8000 -o findings.json
  chromascan http://target:8000 --write-canary --authorize-write
  chromascan http://target:8000 --tenant mytenant --database mydb
`

func main() {
	var (
		probeOnly  = flag.Bool("probe", false, "")
		doHunt     = flag.Bool("hunt", false, "")
		doCanary   = flag.Bool("write-canary", false, "")
		authWrite  = flag.Bool("authorize-write", false, "")
		tenant     = flag.String("tenant", "default_tenant", "")
		database   = flag.String("database", "default_database", "")
		timeout    = flag.Int("timeout", 10, "")
		outputFile = flag.String("o", "", "")
		verbose    = flag.Bool("v", false, "")
	)

	flag.Usage = func() { fmt.Print(usage) }

	// Go's flag package stops at the first non-flag arg, so flags written
	// after the target would be ignored. Reparse iteratively: parse flags,
	// take the first leftover positional, then reparse the rest. This keeps
	// value flags paired with their value ("-o file"), which a naive
	// dash-prefix split would break.
	var posArgs []string
	rest := os.Args[1:]
	for {
		if err := flag.CommandLine.Parse(rest); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		leftover := flag.Args()
		if len(leftover) == 0 {
			break
		}
		posArgs = append(posArgs, leftover[0])
		rest = leftover[1:]
	}

	if len(posArgs) == 0 {
		fmt.Print(usage)
		os.Exit(1)
	}

	if *doCanary && !*authWrite {
		fmt.Fprintln(os.Stderr, "[!] --write-canary requires --authorize-write")
		os.Exit(1)
	}

	target := posArgs[0]
	if !strings.HasPrefix(target, "http") {
		target = "http://" + target
	}
	target = strings.TrimRight(target, "/")

	// --hunt is the default mode; --probe suppresses enumeration.
	// Explicit --hunt flag is accepted but has no additional effect beyond default.
	_ = doHunt

	cfg := &scanner.Config{
		Target:        target,
		Tenant:        *tenant,
		Database:      *database,
		TimeoutSec:    *timeout,
		DoWriteCanary: *doCanary && *authWrite,
		ProbeOnly:     *probeOnly,
		OutputFile:    *outputFile,
		Verbose:       *verbose,
	}

	result := &scanner.ScanResult{
		Target:   target,
		ScanTime: time.Now().UTC(),
	}

	s := scanner.New(cfg)
	if err := s.Run(result); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	scanner.PrintReport(result)

	data, _ := json.MarshalIndent(result, "", "  ")
	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, data, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "write output: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[+] JSON -> %s\n", *outputFile)
	}
}
