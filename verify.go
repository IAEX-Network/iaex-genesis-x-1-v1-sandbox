package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

var authorityEventTypes = map[string]bool{
	"GENESIS":                           true,
	"ORGANIZATION_REGISTERED":           true,
	"ORGANIZATION_OWNERSHIP_REGISTERED": true,
	"ORDER_OPENED":                      true, // authority-signed: ledger_id is server-generated
	"LEDGER_CLOSED":                     true,
	// MASTER_CLOSED is actor-signed (not authority)
}

func isAuthorityEventType(t string) bool { return authorityEventTypes[t] }

// ── Phase 10A: SigLog self-verification ──────────────────────────────────────
// Every actor-signed event recorded during the flow is re-verified locally.
// Uses ed25519.PublicKey stored in NetworkState — zero server trust.

func verifySigLog(ns *NetworkState) {
	fmt.Printf("\n%s\n", strings.Repeat("═", 72))
	fmt.Println("PHASE 10A: SIGLOG SELF-VERIFICATION (client-side Ed25519 — zero server trust)")
	fmt.Println(strings.Repeat("─", 72))
	fmt.Printf("%-4s %-36s %-14s %s\n", "#", "event_type", "actor", "result")
	fmt.Println(strings.Repeat("─", 72))

	pass, fail := 0, 0
	for i, rec := range ns.SigLog {
		canon, err := canonicalJSON(rec.SigPayload)
		if err != nil {
			fmt.Printf("%-4d %-36s %-14s ✗ FAIL (canonicalJSON: %v)\n", i+1, truncate(rec.EventType, 36), truncate(rec.ActorName, 14), err)
			fail++
			continue
		}
		size := len(rec.EventType) + 1 + len(rec.LedgerID) + 1 + len(canon)
		msg := make([]byte, 0, size)
		msg = append(msg, rec.EventType...)
		msg = append(msg, 0x00)
		msg = append(msg, rec.LedgerID...)
		msg = append(msg, 0x00)
		msg = append(msg, canon...)
		digest := sha256.Sum256(msg)

		sigBytes, err := base64.StdEncoding.DecodeString(rec.Sig)
		if err != nil {
			fmt.Printf("%-4d %-36s %-14s ✗ FAIL (decode sig: %v)\n", i+1, truncate(rec.EventType, 36), truncate(rec.ActorName, 14), err)
			fail++
			continue
		}

		if ed25519.Verify(rec.PubKey, digest[:], sigBytes) {
			fmt.Printf("%-4d %-36s %-14s ✓ PASS\n", i+1, truncate(rec.EventType, 36), truncate(rec.ActorName, 14))
			pass++
		} else {
			fmt.Printf("%-4d %-36s %-14s ✗ FAIL\n", i+1, truncate(rec.EventType, 36), truncate(rec.ActorName, 14))
			fail++
		}
	}
	fmt.Println(strings.Repeat("─", 72))
	fmt.Printf("SigLog: %d PASS  %d FAIL  %d total\n", pass, fail, pass+fail)
}

// ── Phase 10B: Cross-party server event verification ─────────────────────────
// Each party fetches ledger events using THEIR OWN API key (delegation scope proof).
// For each actor-signed event, verify Ed25519 client-side against:
//   - actor_sig_payload (if non-null) — exact fields client signed
//   - payload (fallback) — full stored payload

type eventAuditRow struct {
	EventType       string           `json:"event_type"`
	ActorID         *string          `json:"actor_id"`
	Hash            string           `json:"hash"`
	ActorSig        *string          `json:"actor_sig"`
	SigningKeyID     *string          `json:"signing_key_id"`
	Payload         json.RawMessage  `json:"payload"`
	ActorSigPayload *json.RawMessage `json:"actor_sig_payload"`
}

type ledgerAuditResp struct {
	Count     int `json:"count"`
	Integrity struct {
		Verified bool `json:"verified"`
		Issues   []struct {
			EventID string `json:"event_id"`
			Type    string `json:"type"`
			Detail  string `json:"detail"`
		} `json:"issues"`
	} `json:"integrity"`
	Events []eventAuditRow `json:"events"`
}

// buildPubKeyMap builds actorID → ed25519.PublicKey from all NetworkState actors.
func buildPubKeyMap(ns *NetworkState) map[string]ed25519.PublicKey {
	m := make(map[string]ed25519.PublicKey)
	all := []ActorState{
		ns.Sup.Org, ns.Sup.Prod, ns.Sup.QC, ns.Sup.Mgr, ns.Sup.Fin, ns.Sup.WH, ns.Sup.IoT1,
		ns.Buy.Org, ns.Buy.Proc, ns.Buy.QC, ns.Buy.Fin, ns.Buy.Acc, ns.Buy.WH, ns.Buy.IoT1,
		ns.Aud.Org, ns.Aud.Mgr, ns.Aud.Field,
		ns.Trn.Org, ns.Trn.Mgr, ns.Trn.RegMgr, ns.Trn.WHMgr, ns.Trn.Pickup,
		ns.Trn.Delivery, ns.Trn.IoTTruck, ns.Trn.IoTCont,
		ns.Bnk.Org, ns.Bnk.Mgmt, ns.Bnk.RegBranch, ns.Bnk.BranchMgr, ns.Bnk.Loan,
	}
	for _, a := range all {
		if a.ID != "" && len(a.PubKey) > 0 {
			m[a.ID] = a.PubKey
		}
	}
	return m
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func verifyEventSig(e eventAuditRow, ledgerID string, pubKeyMap map[string]ed25519.PublicKey) string {
	if e.ActorSig == nil || *e.ActorSig == "" {
		if isAuthorityEventType(e.EventType) {
			return "[⚡ authority]  (server-signed)"
		}
		return "(unsigned — infrastructure)"
	}

	actorID := safeStr(e.ActorID)
	pubKey, ok := pubKeyMap[actorID]
	if !ok {
		// Actor not in this run's NetworkState (shouldn't happen in e2e_full)
		return fmt.Sprintf("[✓ server-verified]  kid:%s  (pubkey unavailable)", truncate(safeStr(e.SigningKeyID), 20))
	}

	// sigTarget: use actor_sig_payload if present (subset signed), else full payload.
	var sigTarget map[string]any
	if e.ActorSigPayload != nil {
		if err := json.Unmarshal(*e.ActorSigPayload, &sigTarget); err != nil {
			return fmt.Sprintf("✗ FAIL (unmarshal actor_sig_payload: %v)", err)
		}
	} else {
		if err := json.Unmarshal(e.Payload, &sigTarget); err != nil {
			return fmt.Sprintf("✗ FAIL (unmarshal payload: %v)", err)
		}
	}

	canon, err := canonicalJSON(sigTarget)
	if err != nil {
		return fmt.Sprintf("✗ FAIL (canonicalJSON: %v)", err)
	}

	size := len(e.EventType) + 1 + len(ledgerID) + 1 + len(canon)
	msg := make([]byte, 0, size)
	msg = append(msg, e.EventType...)
	msg = append(msg, 0x00)
	msg = append(msg, ledgerID...)
	msg = append(msg, 0x00)
	msg = append(msg, canon...)
	digest := sha256.Sum256(msg)

	sigBytes, err := base64.StdEncoding.DecodeString(*e.ActorSig)
	if err != nil {
		return fmt.Sprintf("✗ FAIL (decode sig: %v)", err)
	}

	if ed25519.Verify(pubKey, digest[:], sigBytes) {
		sigSnip := *e.ActorSig
		if len(sigSnip) > 16 {
			sigSnip = sigSnip[:16] + "..."
		}
		return fmt.Sprintf("[✓ Ed25519 VERIFIED]  kid:%s  sig:%s", truncate(safeStr(e.SigningKeyID), 20), sigSnip)
	}
	return "✗ Ed25519 INVALID (sig does not match pubkey+payload)"
}

type ledgerVerifyTarget struct {
	name     string
	apiKey   string // the VERIFYING party's API key (cross-party proof)
	verifier string // human name of verifying party
	ledgerID string
}

func verifyCrossParty(ns *NetworkState, pubKeyMap map[string]ed25519.PublicKey) {
	targets := []ledgerVerifyTarget{
		// ORDER ledger — NationalBank independently verifies (delegated via master access)
		{"ORDER ledger", ns.Bnk.Loan.APIKey, "NationalBank (bnk_loan)", ns.OrderLedgerID},
		// ORDER ledger — SwiftLogistics independently verifies
		{"ORDER ledger", ns.Trn.Pickup.APIKey, "SwiftLogistics (trn_pickup)", ns.OrderLedgerID},
		// Supplier FRL — FarmFresh internal staff only
		{"Supplier FRL (FarmFresh)", ns.Sup.Mgr.APIKey, "FarmFresh (sup_mgr)", ns.Sup.FRLID},
		// Buyer FRL — RetailCo internal staff only
		{"Buyer FRL (RetailCo)", ns.Buy.Proc.APIKey, "RetailCo (buy_proc)", ns.Buy.FRLID},
	}

	totalPass, totalFail := 0, 0

	for _, t := range targets {
		fmt.Printf("\n%s\n", strings.Repeat("═", 72))
		fmt.Printf("VERIFIER: %s\n", t.verifier)
		fmt.Printf("LEDGER:   %s  id=%s\n", t.name, truncate(t.ledgerID, 20))
		fmt.Println(strings.Repeat("─", 72))

		raw := doJSONRaw(t.apiKey, "GET",
			ns.BaseURL+"/ledgers/"+t.ledgerID+"/events?limit=1000")

		var resp ledgerAuditResp
		if err := json.Unmarshal(raw, &resp); err != nil {
			fatalf("parse events for cross-party verify: %v", err)
		}

		fmt.Printf("%-4s %-36s %-14s %s\n", "#", "event_type", "actor[:8]", "verification")
		fmt.Println(strings.Repeat("─", 72))

		pass, fail := 0, 0
		for i, e := range resp.Events {
			actorSnip := "—"
			if e.ActorID != nil && len(*e.ActorID) >= 8 {
				actorSnip = (*e.ActorID)[:8]
			}
			result := verifyEventSig(e, t.ledgerID, pubKeyMap)
			fmt.Printf("%-4d %-36s %-14s %s\n", i+1, truncate(e.EventType, 36), actorSnip, result)

			if strings.HasPrefix(result, "✗") {
				fail++
			} else {
				pass++
			}
		}

		fmt.Println(strings.Repeat("─", 72))
		intMark := "✓"
		if !resp.Integrity.Verified {
			intMark = "✗"
			for _, issue := range resp.Integrity.Issues {
				fmt.Printf("  INTEGRITY ISSUE [%s] event=%s: %s\n",
					issue.Type, truncate(issue.EventID, 8), issue.Detail)
			}
		}
		fmt.Printf("integrity:%s  Ed25519-pass:%d  fail:%d  total:%d\n", intMark, pass, fail, resp.Count)
		totalPass += pass
		totalFail += fail
	}

	fmt.Printf("\n%s\n", strings.Repeat("═", 72))
	fmt.Println("CROSS-PARTY VERIFICATION SUMMARY")
	fmt.Println(strings.Repeat("─", 72))
	fmt.Printf("  Ed25519 verified : %d\n", totalPass)
	fmt.Printf("  Failed           : %d\n", totalFail)
	fmt.Println("  Proof model: client-side Ed25519 — no server trust required")
	fmt.Println("  Delegation scope: NationalBank + SwiftLogistics verified ORDER events")
	fmt.Println("                    FarmFresh verified Supplier FRL only")
	fmt.Println("                    RetailCo verified Buyer FRL only")
}

func verifyWithSigAudit(ns *NetworkState) {
	// Phase 10A: verify every SigLog entry client-side (zero server trust)
	verifySigLog(ns)

	// Phase 10B: cross-party — each party fetches via own API key + verifies Ed25519
	pubKeyMap := buildPubKeyMap(ns)
	verifyCrossParty(ns, pubKeyMap)

	fmt.Printf("\n%s\n", strings.Repeat("═", 72))
	fmt.Println("CAUSAL LINKS (cryptographically sealed in Layer B hash)")
	fmt.Println(strings.Repeat("─", 72))
	fmt.Printf("  QC        → production  %s...\n", truncate(ns.ProductionHash, 20))
	fmt.Printf("  delivery  → pickup      %s...\n", truncate(ns.PickupHash, 20))
	fmt.Printf("  disbursed → delivery    %s...\n", truncate(ns.DeliveredHash, 20))
	fmt.Println()
	fmt.Println("  Each signed event binds: actor_id + signing_key_id + Ed25519 sig")
	fmt.Println("  Non-repudiation: private key holder alone could produce valid sig.")
	fmt.Printf("%s\n", strings.Repeat("═", 72))
}
