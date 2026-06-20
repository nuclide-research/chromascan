# chromascan

ChromaDB unauth enumeration tool. Fingerprints, enumerates, and PII-scans exposed ChromaDB instances with no credentials.

```
nuclide-research.com
```

## Install

```bash
go install github.com/nuclide-research/chromascan@latest
```

Or build from source (no external dependencies):

```bash
git clone https://github.com/nuclide-research/chromascan
cd chromascan
go build -o chromascan .
```

## Usage

```
chromascan [flags] <target-url>
```

### Flags

| Flag | Description |
|------|-------------|
| `--probe` | Fingerprint only: version, auth, collection count |
| `--hunt` | Full enumeration (default): fingerprint + collections + counts + samples + PII |
| `--write-canary` | Inject write canary (requires `--authorize-write`) |
| `--authorize-write` | Explicit authorization gate for write operations |
| `--tenant <name>` | ChromaDB tenant (default: `default_tenant`) |
| `--database <name>` | ChromaDB database (default: `default_database`) |
| `--timeout <sec>` | HTTP timeout in seconds (default: 10) |
| `-o <file>` | Write JSON output to file |
| `-v` | Verbose: log HTTP exchanges |

### Examples

```bash
# Full enumeration (default)
chromascan http://target:8000

# Probe only -- auth check + version, no document reads
chromascan http://target:8000 --probe

# Full enumeration with JSON output
chromascan http://target:8000 -o findings.json

# Non-default tenant/database
chromascan http://target:8000 --tenant mytenant --database mydb

# Confirm write access (authorized assessments only)
chromascan http://target:8000 --write-canary --authorize-write

# Verbose HTTP logging
chromascan http://target:8000 -v
```

## What It Does

### Phase 0 -- Probe

- `GET /api/v2/heartbeat` (falls back to v1 if 404)
- `GET /api/v2/version` -- raw version string
- Auth check: collections list; 401/403 = AUTH_REQUIRED, stop

### Phase 1 -- Collections

- `GET /api/v2/tenants/{tenant}/databases/{db}/collections` (v1 fallback)
- `GET .../collections/{id}/count` -- record count per collection

### Phase 2 -- Sample + PII Scan

- `POST .../collections/{id}/get` body `{"limit":3,"include":["documents","metadatas"]}`
- Embeddings never requested
- PII scan across document text and metadata values:
  - email addresses (regex)
  - API keys (`sk-`, `AKIA`, `ghp_`, `Bearer `)
  - medical terms (patient, diagnosis, prescription, clinical, PHI)
  - personal names in metadata fields named author/user/name/email

### Write Canary

- `POST .../collections/{id}/add` -- injects `nuclide-canary-{ts}`
- Immediately deleted on success (`POST .../delete`)
- Requires both `--write-canary` and `--authorize-write`; neither alone fires

## Scoring

Per-collection score (additive):

| Condition | Points |
|-----------|--------|
| AUTH_OFF (collections reachable) | +4.0 |
| PII detected | +3.0 |
| Write confirmed | +1.5 |
| Large corpus (>10k records) | +0.5 |

Severity thresholds:

| Score | Severity |
|-------|----------|
| 0.0 - 3.9 | LOW |
| 4.0 - 5.9 | MEDIUM |
| 6.0 - 7.9 | HIGH |
| 8.0+ | CRITICAL |

## Output

Human-readable colored output by default. JSON with `-o`.

```
  version    : 1.4.2
  api        : v2
  auth       : OPEN (no auth)

  COLLECTION                        COUNT  SCORE  SEVERITY
  customer_support_tickets          14823    7.5  HIGH [PII: email,personal_name]
  internal_docs                       312    4.0  MEDIUM

  severity   : HIGH
  score      : 7.5
  collections: 2 total
  objects    : ~15135 estimated
  pii hits   : 1 collections
  findings   : 6 extracted
```

## API Coverage

Tested against ChromaDB v1 and v2. Heartbeat, version, collection list, count, and get endpoints all fall back from v2 to v1 on 404.

## Notes

Built for authorized security assessment of AI/ML infrastructure. Operates within the NuClide research methodology -- verification-first, restraint-governed, primary-source evidence. Use on systems you are authorized to test.
