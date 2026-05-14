# Genesis X-1 — Network Sandbox

Run a complete 5-party supply-chain trade on the Genesis X-1 live network.

Simulates actor registration → organization onboarding → delegation → order ledger → 15 signed events → cross-party Ed25519 verification.

---

## Option 1 — Hosted (no install, no download)

Trigger the full verification directly from the live server via SSE stream:

```bash
curl -N "https://api.iaexnetwork.com/sandbox/run"
```

Or open in any browser:
```
https://api.iaexnetwork.com/sandbox/run
```

Output streams in real time as each phase completes. No API key required.

---

## Option 2 — Download Binary (no Go needed)

**[→ Download Latest Release](https://github.com/IAEX-Network/iaex-genesis-x-1-v1-sandbox/releases/tag/Release-Sanbox)**

| Platform | File |
|---|---|
| Windows | `e2e_full_windows_amd64.exe` |
| Linux | `e2e_full_linux_amd64` |
| macOS (Apple Silicon) | `e2e_full_macos_arm64` |

### Windows
Double-click `e2e_full_windows_amd64.exe` — the console window stays open until you press Enter.

Or run from Command Prompt / PowerShell:
```
e2e_full_windows_amd64.exe
```

### Linux / macOS
```bash
chmod +x e2e_full_linux_amd64
./e2e_full_linux_amd64
```

---

## Option 3 — Run with Go (no clone needed)

```bash
go run github.com/IAEX-Network/iaex-genesis-x-1-sandbox@latest
```

---

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-base-url` | `https://api.iaexnetwork.com` | API base URL |
| `-region` | `in` | Actor registration region (`in` or `eu`) |

---

## Expected Output

```
PHASE 10A: SIGLOG SELF-VERIFICATION (client-side Ed25519 — zero server trust)
SigLog: 74 PASS  0 FAIL  74 total

CROSS-PARTY VERIFICATION SUMMARY
Ed25519 verified : 116
Failed           : 0
Proof model: client-side Ed25519 — no server trust required
```

Zero exit = all phases passed including cross-party Ed25519 verification.

---

## What It Tests

- 5 organizations, ~30 actors — all Ed25519 keys enrolled at registration
- Every business event actor-signed (`X-Actor-Sig` header)
- Phase 10A: self-verify all 74 signatures locally (zero server trust)
- Phase 10B: each party independently fetches and verifies via own API key
- Causal hash chains: `QC → PRODUCTION`, `DELIVERED → PICKUP`, `LOAN_DISBURSED → DELIVERED`

Protocol documentation: [developer.iaexnetwork.com](https://developer.iaexnetwork.com)
