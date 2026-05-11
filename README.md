# Genesis X-1 — Network Sandbox (e2e_full)

Run a complete 5-party supply-chain trade on the Genesis X-1 live network.

Simulates: actor registration → organization onboarding → delegation → order ledger → 15 signed events → cross-party Ed25519 verification.

## Requirements

Go 1.21+

## Run (no clone needed)

```bash
go run github.com/IAEX-Network/iaex-genesis-x-1-sandbox@latest
```

## Run from clone

```bash
git clone https://github.com/IAEX-Network/iaex-genesis-x-1-sandbox
cd iaex-genesis-x-1-sandbox
go run .
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-base-url` | `https://api.iaexnetwork.com` | API base URL |
| `-region` | `in` | Actor registration region (`in` or `eu`) |

## Expected Output

```
SigLog: 74 PASS  0 FAIL  74 total
Ed25519 verified : 116
Failed           : 0
Proof model: client-side Ed25519 — no server trust required
```

Zero exit = all phases passed including cross-party verification.

## What It Tests

- 5 organizations, ~30 actors, all Ed25519 keys enrolled at registration
- Every business event actor-signed (`X-Actor-Sig` header)
- Phase 10A: self-verify all 74 signatures locally (zero server trust)
- Phase 10B: each party independently fetches and verifies via own API key

Protocol: [developer.iaexnetwork.com](https://developer.iaexnetwork.com)
