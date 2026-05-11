package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

var httpClient = &http.Client{}

// ── HTTP helpers ─────────────────────────────────────────────────────────────

// doJSON: verbose — prints status + body. Use for business events shown to user.
func doJSON(apiKey, kid, sig, method, rawURL string, body any) map[string]any {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, rawURL, reqBody)
	if err != nil {
		fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	if kid != "" {
		req.Header.Set("X-Signing-Key-ID", kid)
	}
	if sig != "" {
		req.Header.Set("X-Actor-Sig", sig)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		fatalf("http %s %s: %v", method, rawURL, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	fmt.Printf("HTTP %d\n", resp.StatusCode)
	fmt.Println(prettyBytes(raw))
	if resp.StatusCode >= 400 {
		fatalf("request failed — see response above")
	}
	return result
}

// doJSONQ: quiet — suppresses output, used for bulk setup calls.
func doJSONQ(apiKey, kid, sig, method, rawURL string, body any) map[string]any {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, rawURL, reqBody)
	if err != nil {
		fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	if kid != "" {
		req.Header.Set("X-Signing-Key-ID", kid)
	}
	if sig != "" {
		req.Header.Set("X-Actor-Sig", sig)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		fatalf("http %s %s: %v", method, rawURL, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Printf("HTTP %d\n%s\n", resp.StatusCode, raw)
		fatalf("request failed")
	}
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	return result
}

func doJSONRaw(apiKey, method, rawURL string) []byte {
	req, err := http.NewRequest(method, rawURL, nil)
	if err != nil {
		fatalf("new request: %v", err)
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		fatalf("http %s %s: %v", method, rawURL, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Printf("HTTP %d\n%s\n", resp.StatusCode, raw)
		fatalf("request failed")
	}
	return raw
}

// ── Network discovery ─────────────────────────────────────────────────────────

func lookupActor(apiKey, baseURL, iaexID string) string {
	raw := doJSONRaw(apiKey, "GET", baseURL+"/actors/lookup?iaex_id="+url.QueryEscape(iaexID))
	var resp map[string]any
	if err := json.Unmarshal(raw, &resp); err != nil {
		fatalf("lookupActor parse: %v", err)
	}
	id, _ := resp["actor_id"].(string)
	if id == "" {
		fatalf("lookupActor: actor_id not found for %s", iaexID)
	}
	return id
}

// ── Actor registration (key enrolled at registration) ─────────────────────────

// registerActor registers a developer actor and enrolls an Ed25519 key in one shot.
// The server returns sandbox.signing_key.kid — no separate /actors/{id}/keys call needed.
func registerActor(baseURL, region, name string) ActorState {
	slug := strings.ToLower(strings.NewReplacer(" ", ".", "/", "-").Replace(name))
	email := slug + "@e2e.iaex.test"

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fatalf("generate key: %v", err)
	}
	pubB64 := base64.StdEncoding.EncodeToString(pub)

	resp := doJSONQ("", "", "", "POST", baseURL+"/developer/register", map[string]any{
		"name":       name,
		"email":      email,
		"region":     region,
		"public_key": pubB64,
	})
	sb := nested(resp, "sandbox")
	id := str(sb, "actor_id")
	apiKey := str(sb, "api_key")
	kid := strPath(sb, "signing_key", "kid")
	if id == "" || apiKey == "" || kid == "" {
		fatalf("registerActor: missing actor_id, api_key, or signing_key.kid for %s", name)
	}
	return ActorState{
		ID:      id,
		IaexID:  "iaex:actor:" + id,
		Name:    name,
		APIKey:  apiKey,
		PrivKey: priv,
		PubKey:  pub,
		Kid:     kid,
	}
}

// ── Signed setup helpers ──────────────────────────────────────────────────────

// onboardOrg registers an organization. Returns (frlLedgerID, organizationID) directly
// from the POST /onboarding/organization response (no additional GET needed).
func onboardOrg(baseURL string, org *ActorState, legalName, country, vatTax, email string) (frlID, orgID string) {
	resp := doJSONQ(org.APIKey, "", "", "POST", baseURL+"/onboarding/organization", map[string]any{
		"legal_name":       legalName,
		"country":          country,
		"region":           "Delhi",
		"city":             "New Delhi",
		"vat_tax_number":   vatTax,
		"primary_email":    email,
		"terms_version":    "1.0",
		"declaration_hash": "e2e-full-declaration",
	})
	frlID = str(resp, "frl_ledger_id")
	orgID = str(resp, "organization_id")
	if frlID == "" || orgID == "" {
		fatalf("onboardOrg: missing frl_ledger_id or organization_id for %s", legalName)
	}
	return
}

// grantOrgDelegation signs SUPPLIER_DELEGATION_GRANTED and calls POST /delegation/organization.
// ledgerID = frlID (parent ledger of the org delegation event).
func grantOrgDelegation(ns *NetworkState, actor ActorState, frlID, orgID, delegateID, role string) {
	sigPayload := map[string]any{
		"delegate_actor_id": delegateID,
		"organization_id":   orgID,
		"permissions":       map[string]any{},
		"role":              role,
	}
	sig := signEvent(actor.PrivKey, "SUPPLIER_DELEGATION_GRANTED", frlID, sigPayload)
	ns.SigLog = append(ns.SigLog, SigRecord{
		ActorID: actor.ID, ActorName: actor.Name, PubKey: actor.PubKey,
		EventType: "SUPPLIER_DELEGATION_GRANTED", LedgerID: frlID,
		SigPayload: sigPayload, Sig: sig,
	})
	doJSONQ(actor.APIKey, actor.Kid, sig, "POST", ns.BaseURL+"/delegation/organization", map[string]any{
		"organization_id":   orgID,
		"delegate_actor_id": delegateID,
		"role":              role,
		"permissions":       map[string]any{},
	})
}

// createMaster signs TRACELEDGER_MASTER_CREATED and calls POST /traceledger/master.
// Returns the master UUID.
func createMaster(ns *NetworkState, actor ActorState, ledgerID, businessRef, scope string) string {
	sigPayload := map[string]any{
		"ledger_id":    ledgerID,
		"business_ref": businessRef,
		"scope":        scope,
	}
	sig := signEvent(actor.PrivKey, "TRACELEDGER_MASTER_CREATED", ledgerID, sigPayload)
	ns.SigLog = append(ns.SigLog, SigRecord{
		ActorID: actor.ID, ActorName: actor.Name, PubKey: actor.PubKey,
		EventType: "TRACELEDGER_MASTER_CREATED", LedgerID: ledgerID,
		SigPayload: sigPayload, Sig: sig,
	})
	resp := doJSONQ(actor.APIKey, actor.Kid, sig, "POST", ns.BaseURL+"/traceledger/master", map[string]any{
		"ledger_id":    ledgerID,
		"business_ref": businessRef,
		"scope":        scope,
	})
	masterUUID := str(resp, "uuid")
	if masterUUID == "" {
		fatalf("createMaster: missing uuid in response")
	}
	return masterUUID
}

// grantMasterAccess signs MASTER_DELEGATION_GRANTED and calls POST /delegation/master.
// parentLedgerID = the ledger that owns the master (frlID for FRL masters, orderLedgerID for ORDER masters).
func grantMasterAccess(ns *NetworkState, actor ActorState, parentLedgerID, masterUUID, delegateID, role string) {
	sigPayload := map[string]any{
		"traceledger_master_id": masterUUID,
		"delegate_actor_id":     delegateID,
		"role":                  role,
		"permissions":           map[string]any{"access": "write"},
	}
	sig := signEvent(actor.PrivKey, "MASTER_DELEGATION_GRANTED", parentLedgerID, sigPayload)
	ns.SigLog = append(ns.SigLog, SigRecord{
		ActorID: actor.ID, ActorName: actor.Name, PubKey: actor.PubKey,
		EventType: "MASTER_DELEGATION_GRANTED", LedgerID: parentLedgerID,
		SigPayload: sigPayload, Sig: sig,
	})
	doJSONQ(actor.APIKey, actor.Kid, sig, "POST", ns.BaseURL+"/delegation/master", map[string]any{
		"master_id":         masterUUID,
		"delegate_actor_id": delegateID,
		"role":              role,
		"permissions":       map[string]any{"access": "write"},
	})
}

// ── Layer A signing ───────────────────────────────────────────────────────────

func signEvent(priv ed25519.PrivateKey, eventType, ledgerID string, payload map[string]any) string {
	canon, err := canonicalJSON(payload)
	if err != nil {
		fatalf("canonicalJSON: %v", err)
	}
	size := len(eventType) + 1 + len(ledgerID) + 1 + len(canon)
	msg := make([]byte, 0, size)
	msg = append(msg, eventType...)
	msg = append(msg, 0x00)
	msg = append(msg, ledgerID...)
	msg = append(msg, 0x00)
	msg = append(msg, canon...)
	digest := sha256.Sum256(msg)
	sig := ed25519.Sign(priv, digest[:])
	return base64.StdEncoding.EncodeToString(sig)
}

func injectMasterPayload(m MasterState, userPayload map[string]any) map[string]any {
	p := clonePayload(userPayload)
	p["traceledger_master_uuid"] = m.UUID
	p["business_ref"] = m.BusinessRef
	p["scope"] = m.Scope
	switch m.LedgerType {
	case "FACILITY_ROOT":
		p["supplier_actor_id"] = m.SupplierActorID
	case "ORDER":
		p["buyer_actor_id"] = m.BuyerActorID
		p["supplier_actor_id"] = m.SupplierActorID
	}
	return p
}

// sendMasterEvent: inject server fields → sign → POST → print response + signing key.
func sendMasterEvent(ns *NetworkState, actor ActorState, m MasterState, eventType string, userPayload map[string]any) map[string]any {
	fullPayload := injectMasterPayload(m, userPayload)
	sig := signEvent(actor.PrivKey, eventType, m.LedgerID, fullPayload)
	ns.SigLog = append(ns.SigLog, SigRecord{
		ActorID: actor.ID, ActorName: actor.Name, PubKey: actor.PubKey,
		EventType: eventType, LedgerID: m.LedgerID,
		SigPayload: fullPayload, Sig: sig,
	})
	result := doJSON(
		actor.APIKey, actor.Kid, sig,
		"POST",
		ns.BaseURL+"/traceledger/master/"+m.UUID+"/events",
		map[string]any{"event_type": eventType, "payload": fullPayload},
	)
	fmt.Printf("  [signed] kid:%s\n", actor.Kid)
	return result
}

// ── Event hash helpers ────────────────────────────────────────────────────────

func getLastEventHash(apiKey, baseURL, ledgerID string) string {
	raw := doJSONRaw(apiKey, "GET", baseURL+"/ledgers/"+ledgerID+"/events?limit=1000")
	var resp struct {
		Events []struct {
			Hash string `json:"hash"`
		} `json:"events"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		fatalf("parse events: %v", err)
	}
	if len(resp.Events) == 0 {
		fatalf("no events in ledger %s", ledgerID)
	}
	return resp.Events[len(resp.Events)-1].Hash
}

// ── CanonicalJSON ─────────────────────────────────────────────────────────────

func canonicalJSON(input map[string]any) ([]byte, error) {
	if input == nil {
		return []byte("{}"), nil
	}
	return canonicalValue(input)
}

func canonicalValue(v any) ([]byte, error) {
	switch typed := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for k := range typed {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var buf bytes.Buffer
		buf.WriteByte('{')
		for i, k := range keys {
			enc, err := json.Marshal(k)
			if err != nil {
				return nil, err
			}
			val, err := canonicalValue(typed[k])
			if err != nil {
				return nil, err
			}
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(enc)
			buf.WriteByte(':')
			buf.Write(val)
		}
		buf.WriteByte('}')
		return buf.Bytes(), nil
	case []any:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, elem := range typed {
			val, err := canonicalValue(elem)
			if err != nil {
				return nil, err
			}
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.Write(val)
		}
		buf.WriteByte(']')
		return buf.Bytes(), nil
	default:
		return json.Marshal(v)
	}
}

// ── Utilities ─────────────────────────────────────────────────────────────────

func clonePayload(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func prettyBytes(b []byte) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, b, "", "  "); err != nil {
		return string(b)
	}
	return buf.String()
}

func printBanner(party, actorCode, eventType, method, rawURL string) {
	line := strings.Repeat("═", 60)
	fmt.Printf("\n%s\n", line)
	fmt.Printf("[%s] %s → %s\n", party, actorCode, eventType)
	fmt.Printf("%s %s\n", method, rawURL)
	fmt.Println(strings.Repeat("─", 60))
}

func fatalf(format string, args ...any) {
	panic(fmt.Sprintf("FATAL: "+format, args...))
}

func str(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, _ := m[key].(string)
	return v
}

// strPath traverses nested maps by key path.
func strPath(m map[string]any, keys ...string) string {
	v := any(m)
	for _, k := range keys {
		mv, ok := v.(map[string]any)
		if !ok {
			return ""
		}
		v = mv[k]
	}
	s, _ := v.(string)
	return s
}

func nested(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	v, _ := m[key].(map[string]any)
	return v
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
