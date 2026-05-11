package main

import "crypto/ed25519"

// SigRecord captures everything needed for independent client-side Ed25519 verification.
// Stored during the e2e flow; used in Phase 10 cross-party audit.
type SigRecord struct {
	ActorID    string             // signer's actor UUID
	ActorName  string             // human label (e.g. "trn_pickup")
	PubKey     ed25519.PublicKey  // signer's public key (from registration)
	EventType  string
	LedgerID   string
	SigPayload map[string]any     // exact fields signed (Layer A input)
	Sig        string             // base64 Ed25519 signature
}

type ActorState struct {
	ID      string
	IaexID  string // "iaex:actor:<uuid>"
	Name    string // human label used in SigLog
	APIKey  string
	PrivKey ed25519.PrivateKey
	PubKey  ed25519.PublicKey
	Kid     string
}

type MasterState struct {
	UUID            string
	LedgerID        string
	LedgerType      string // "FACILITY_ROOT" | "ORDER"
	BusinessRef     string
	Scope           string
	SupplierActorID string
	BuyerActorID    string
}

// ── Supplier party (FarmFresh Ltd) ─────────────────────────────────────────
type SupParty struct {
	BaseURL string
	Region  string
	Org     ActorState
	Prod    ActorState
	QC      ActorState
	Mgr     ActorState
	Fin     ActorState
	WH      ActorState
	IoT1    ActorState
	OrgID   string
	FRLID   string
	Master  MasterState
}

// ── Buyer party (RetailCo Ltd) ─────────────────────────────────────────────
type BuyParty struct {
	BaseURL string
	Region  string
	Org     ActorState
	Proc    ActorState
	QC      ActorState
	Fin     ActorState
	Acc     ActorState
	WH      ActorState
	IoT1    ActorState
	OrgID   string
	FRLID   string
	Master  MasterState
}

// ── Auditor party (TrustAudit Inc) ────────────────────────────────────────
type AudParty struct {
	BaseURL string
	Region  string
	Org     ActorState
	Mgr     ActorState
	Field   ActorState
	OrgID   string
	FRLID   string
}

// ── Transport party (SwiftLogistics) ──────────────────────────────────────
type TrnParty struct {
	BaseURL  string
	Region   string
	Org      ActorState
	Mgr      ActorState
	RegMgr   ActorState
	WHMgr    ActorState
	Pickup   ActorState
	Delivery ActorState
	IoTTruck ActorState
	IoTCont  ActorState
	OrgID    string
	FRLID    string
}

// ── Bank party (NationalBank) ──────────────────────────────────────────────
type BnkParty struct {
	BaseURL   string
	Region    string
	Org       ActorState
	Mgmt      ActorState
	RegBranch ActorState
	BranchMgr ActorState
	Loan      ActorState
	OrgID     string
	FRLID     string
}

// ── NetworkState — shared cross-party context ──────────────────────────────
type NetworkState struct {
	BaseURL string
	Region  string
	RunID   string // unique suffix per run — prevents onboarding conflicts
	Sup     *SupParty
	Buy     *BuyParty
	Aud     *AudParty
	Trn     *TrnParty
	Bnk     *BnkParty
	// Trade ledger (supplier creates, buyer is counterparty)
	OrderLedgerID string
	OrderMaster   MasterState
	// Causal link hashes (set during event phases)
	ProductionHash string
	PickupHash     string
	DeliveredHash  string
	DisbursedHash  string
	// SigLog: all actor-signed events recorded during the flow for Phase 10 verification
	SigLog []SigRecord
}
