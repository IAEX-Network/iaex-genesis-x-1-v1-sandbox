package main

import "fmt"

// ── Auditor, Transport, Bank registration + onboarding ───────────────────────

func audRegister(ns *NetworkState) {
	fmt.Println("\n  [TrustAudit Inc — Auditor]")
	ns.Aud.Org = registerActor(ns.BaseURL, ns.Region, "TrustAudit Inc")
	ns.Aud.Mgr = registerActor(ns.BaseURL, ns.Region, "aud_mgr")
	ns.Aud.Field = registerActor(ns.BaseURL, ns.Region, "aud_field")
	fmt.Printf("    org=%s  (+2 staff)  [keys enrolled at registration]\n", truncate(ns.Aud.Org.ID, 8))
}

func audOnboard(ns *NetworkState) {
	fmt.Println("\n  [TrustAudit Inc] onboarding...")
	ns.Aud.FRLID, ns.Aud.OrgID = onboardOrg(ns.BaseURL, &ns.Aud.Org, "TrustAudit Inc", "IN", "TAUD"+ns.RunID+"IN", "aud."+ns.RunID+"@e2e.test")
	fmt.Printf("    frl_id=%s  org_id=%s\n", truncate(ns.Aud.FRLID, 8), truncate(ns.Aud.OrgID, 8))
}

func audDelegate(ns *NetworkState) {
	fmt.Println("\n  [TrustAudit Inc] delegating 2 staff — signed...")
	for _, dl := range []struct {
		a    *ActorState
		role string
	}{
		{&ns.Aud.Mgr, "OWNER"}, {&ns.Aud.Field, "OPS"},
	} {
		grantOrgDelegation(ns, ns.Aud.Org, ns.Aud.FRLID, ns.Aud.OrgID, dl.a.ID, dl.role)
		fmt.Printf("    → %s role=%s [signed]\n", truncate(dl.a.ID, 8), dl.role)
	}
}

func trnRegister(ns *NetworkState) {
	fmt.Println("\n  [SwiftLogistics — Transport]")
	ns.Trn.Org = registerActor(ns.BaseURL, ns.Region, "SwiftLogistics")
	ns.Trn.Mgr = registerActor(ns.BaseURL, ns.Region, "trn_mgr")
	ns.Trn.RegMgr = registerActor(ns.BaseURL, ns.Region, "trn_regmgr")
	ns.Trn.WHMgr = registerActor(ns.BaseURL, ns.Region, "trn_whmgr")
	ns.Trn.Pickup = registerActor(ns.BaseURL, ns.Region, "trn_pickup")
	ns.Trn.Delivery = registerActor(ns.BaseURL, ns.Region, "trn_delivery")
	ns.Trn.IoTTruck = registerActor(ns.BaseURL, ns.Region, "trn_iot_truck")
	ns.Trn.IoTCont = registerActor(ns.BaseURL, ns.Region, "trn_iot_cont")
	fmt.Printf("    org=%s  (+7 staff)  [keys enrolled at registration]\n", truncate(ns.Trn.Org.ID, 8))
}

func trnOnboard(ns *NetworkState) {
	fmt.Println("\n  [SwiftLogistics] onboarding...")
	ns.Trn.FRLID, ns.Trn.OrgID = onboardOrg(ns.BaseURL, &ns.Trn.Org, "SwiftLogistics Pvt Ltd", "IN", "SWFT"+ns.RunID+"IN", "trn."+ns.RunID+"@e2e.test")
	fmt.Printf("    frl_id=%s  org_id=%s\n", truncate(ns.Trn.FRLID, 8), truncate(ns.Trn.OrgID, 8))
}

func trnDelegate(ns *NetworkState) {
	fmt.Println("\n  [SwiftLogistics] delegating 7 staff — signed...")
	for _, dl := range []struct {
		a    *ActorState
		role string
	}{
		{&ns.Trn.Mgr, "OWNER"}, {&ns.Trn.RegMgr, "OPS"}, {&ns.Trn.WHMgr, "OPS"},
		{&ns.Trn.Pickup, "AGENT"}, {&ns.Trn.Delivery, "AGENT"},
		{&ns.Trn.IoTTruck, "AGENT"}, {&ns.Trn.IoTCont, "AGENT"},
	} {
		grantOrgDelegation(ns, ns.Trn.Org, ns.Trn.FRLID, ns.Trn.OrgID, dl.a.ID, dl.role)
		fmt.Printf("    → %s role=%s [signed]\n", truncate(dl.a.ID, 8), dl.role)
	}
}

func bnkRegister(ns *NetworkState) {
	fmt.Println("\n  [NationalBank — Bank]")
	ns.Bnk.Org = registerActor(ns.BaseURL, ns.Region, "NationalBank")
	ns.Bnk.Mgmt = registerActor(ns.BaseURL, ns.Region, "bnk_mgmt")
	ns.Bnk.RegBranch = registerActor(ns.BaseURL, ns.Region, "bnk_regbranch")
	ns.Bnk.BranchMgr = registerActor(ns.BaseURL, ns.Region, "bnk_branch_mgr")
	ns.Bnk.Loan = registerActor(ns.BaseURL, ns.Region, "bnk_loan")
	fmt.Printf("    org=%s  (+4 staff)  [keys enrolled at registration]\n", truncate(ns.Bnk.Org.ID, 8))
}

func bnkOnboard(ns *NetworkState) {
	fmt.Println("\n  [NationalBank] onboarding...")
	ns.Bnk.FRLID, ns.Bnk.OrgID = onboardOrg(ns.BaseURL, &ns.Bnk.Org, "NationalBank Ltd", "IN", "NBNK"+ns.RunID+"IN", "bnk."+ns.RunID+"@e2e.test")
	fmt.Printf("    frl_id=%s  org_id=%s\n", truncate(ns.Bnk.FRLID, 8), truncate(ns.Bnk.OrgID, 8))
}

func bnkDelegate(ns *NetworkState) {
	fmt.Println("\n  [NationalBank] delegating 4 staff — signed...")
	for _, dl := range []struct {
		a    *ActorState
		role string
	}{
		{&ns.Bnk.Mgmt, "OWNER"}, {&ns.Bnk.RegBranch, "OPS"},
		{&ns.Bnk.BranchMgr, "OPS"}, {&ns.Bnk.Loan, "FINANCE"},
	} {
		grantOrgDelegation(ns, ns.Bnk.Org, ns.Bnk.FRLID, ns.Bnk.OrgID, dl.a.ID, dl.role)
		fmt.Printf("    → %s role=%s [signed]\n", truncate(dl.a.ID, 8), dl.role)
	}
}

// ── Trade: ORDER ledger + masters + master access grants ─────────────────────

// tradeCreateOrderLedger: supplier creates ORDER ledger.
// ORDER_OPENED is authority-signed server-side (ledger_id is server-generated;
// client cannot compute the sig digest before the response). No X-Actor-Sig needed.
func tradeCreateOrderLedger(ns *NetworkState) {
	fmt.Println("\n  [FarmFresh] creating ORDER ledger...")
	fmt.Printf("    network lookup: resolving buyer iaex_id=%s\n", ns.Buy.Org.IaexID)
	buyActorID := lookupActor(ns.Sup.Org.APIKey, ns.BaseURL, ns.Buy.Org.IaexID)
	fmt.Printf("    buyer resolved: %s\n", truncate(buyActorID, 8))

	resp := doJSONQ(ns.Sup.Org.APIKey, "", "", "POST", ns.BaseURL+"/ledgers", map[string]any{
		"buyer_iaex_id": ns.Buy.Org.IaexID,
	})
	ns.OrderLedgerID = str(resp, "id")
	fmt.Printf("    order_ledger_id=%s\n", truncate(ns.OrderLedgerID, 8))
}

func tradeCreateOrderMaster(ns *NetworkState) {
	fmt.Println("\n  [FarmFresh] creating ORDER TraceLedgerMaster — signed...")
	businessRef := "shipment-" + ns.RunID
	masterUUID := createMaster(ns, ns.Sup.Org, ns.OrderLedgerID, businessRef, "ORDER")
	ns.OrderMaster = MasterState{
		UUID:            masterUUID,
		LedgerID:        ns.OrderLedgerID,
		LedgerType:      "ORDER",
		BusinessRef:     businessRef,
		Scope:           "ORDER",
		SupplierActorID: ns.Sup.Org.ID,
		BuyerActorID:    ns.Buy.Org.ID,
	}
	fmt.Printf("    order_master=%s [signed]\n", truncate(ns.OrderMaster.UUID, 8))
}

// tradeGrantOrderMasterAccess: supplier grants all external delegates write on ORDER master.
// Each external delegate's ID resolved via network lookup (simulates IAEX ID exchange).
func tradeGrantOrderMasterAccess(ns *NetworkState) {
	fmt.Println("\n  [FarmFresh] granting ORDER master write to 13 cross-party delegates — signed...")
	delegates := []*ActorState{
		&ns.Trn.WHMgr, &ns.Trn.Pickup, &ns.Trn.Delivery, &ns.Trn.IoTTruck, &ns.Trn.IoTCont,
		&ns.Aud.Mgr, &ns.Aud.Field,
		&ns.Bnk.Mgmt, &ns.Bnk.RegBranch, &ns.Bnk.BranchMgr, &ns.Bnk.Loan,
		&ns.Buy.Fin, &ns.Buy.Proc,
	}
	for _, a := range delegates {
		resolvedID := lookupActor(ns.Sup.Org.APIKey, ns.BaseURL, a.IaexID)
		grantMasterAccess(ns, ns.Sup.Org, ns.OrderLedgerID, ns.OrderMaster.UUID, resolvedID, "OPS")
		fmt.Printf("    → write granted to %s [signed]\n", truncate(resolvedID, 8))
	}
}

// ── ORDER ledger business events (signed) ────────────────────────────────────

func tradeOrderEvents(ns *NetworkState) {
	m := ns.OrderMaster

	printBanner("SwiftLogistics", "trn_whmgr", "PICKUP_SCHEDULED",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Trn.WHMgr, m, "PICKUP_SCHEDULED", map[string]any{
		"pickup_date": "2026-05-02",
		"driver_id":   ns.Trn.Pickup.ID,
		"vehicle":     "TRK-001",
	})

	printBanner("SwiftLogistics", "trn_pickup", "GOODS_PICKED_UP",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Trn.Pickup, m, "GOODS_PICKED_UP", map[string]any{
		"pallet_count": 10,
		"seal_number":  "SEAL-XYZ",
	})
	ns.PickupHash = getLastEventHash(ns.Sup.Org.APIKey, ns.BaseURL, ns.OrderLedgerID)
	fmt.Printf("  → pickup_hash=%s\n", truncate(ns.PickupHash, 20))

	printBanner("SwiftLogistics", "trn_iot_truck", "IN_TRANSIT ×2",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	for _, g := range []map[string]any{
		{"lat": 28.6, "lon": 77.2, "speed_kmh": 80, "checkpoint": "Delhi"},
		{"lat": 27.1, "lon": 78.0, "speed_kmh": 95, "checkpoint": "Agra"},
	} {
		sendMasterEvent(ns, ns.Trn.IoTTruck, m, "IN_TRANSIT", g)
	}

	printBanner("SwiftLogistics", "trn_iot_cont", "CONTAINER_TEMP_LOG ×4",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	for i, temp := range []float64{3.8, 4.0, 3.7, 3.9} {
		sendMasterEvent(ns, ns.Trn.IoTCont, m, "CONTAINER_TEMP_LOG", map[string]any{
			"container_id": "CONT-01",
			"temp_celsius": temp,
			"humidity":     62,
			"reading_seq":  i + 1,
		})
	}

	printBanner("SwiftLogistics", "trn_delivery", "DELIVERED (caused_by: pickup)",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Trn.Delivery, m, "DELIVERED", map[string]any{
		"recipient_actor_id": ns.Buy.WH.ID,
		"pallet_count":       10,
		"caused_by_hash":     ns.PickupHash,
	})
	ns.DeliveredHash = getLastEventHash(ns.Sup.Org.APIKey, ns.BaseURL, ns.OrderLedgerID)
	fmt.Printf("  → delivered_hash=%s\n", truncate(ns.DeliveredHash, 20))

	printBanner("TrustAudit", "aud_field", "FIELD_INSPECTION_DONE",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Aud.Field, m, "FIELD_INSPECTION_DONE", map[string]any{
		"checklist_score": 98,
		"remarks":         "compliant",
	})

	printBanner("TrustAudit", "aud_mgr", "AUDIT_REPORT_ISSUED",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Aud.Mgr, m, "AUDIT_REPORT_ISSUED", map[string]any{
		"report_id": "AUDIT-2026-001",
		"verdict":   "PASS",
	})

	printBanner("NationalBank", "bnk_loan", "LOAN_APPLICATION_RECEIVED",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Bnk.Loan, m, "LOAN_APPLICATION_RECEIVED", map[string]any{
		"applicant_actor_id": ns.Sup.Org.ID,
		"amount":             2500000,
		"currency":           "INR",
	})

	printBanner("NationalBank", "bnk_branch_mgr", "LOAN_APPROVED",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Bnk.BranchMgr, m, "LOAN_APPROVED", map[string]any{
		"approved_amount": 2500000,
		"interest_rate":   8.5,
	})

	printBanner("NationalBank", "bnk_loan", "LOAN_DISBURSED (caused_by: delivered)",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Bnk.Loan, m, "LOAN_DISBURSED", map[string]any{
		"amount":         2500000,
		"utr_ref":        "UTR2026XYZ",
		"caused_by_hash": ns.DeliveredHash,
	})
	ns.DisbursedHash = getLastEventHash(ns.Sup.Org.APIKey, ns.BaseURL, ns.OrderLedgerID)
	fmt.Printf("  → disbursed_hash=%s\n", truncate(ns.DisbursedHash, 20))
}

// ── Close ─────────────────────────────────────────────────────────────────────

func tradeClose(ns *NetworkState) {
	caller := ns.Sup.Org

	// ── Close ORDER master ────────────────────────────────────────────────────
	fmt.Println("\n  Closing ORDER master...")
	masterPayload := map[string]any{
		"traceledger_master_uuid": ns.OrderMaster.UUID,
		"ledger_id":               ns.OrderLedgerID,
		"ledger_type":             "ORDER",
		"requested_by_actor_id":   caller.ID,
		"status":                  "CLOSED",
		"supplier_actor_id":       ns.Sup.Org.ID,
	}
	if ns.Sup.OrgID != "" {
		masterPayload["supplier_organization_id"] = ns.Sup.OrgID
	}
	if ns.Buy.OrgID != "" {
		masterPayload["buyer_organization_id"] = ns.Buy.OrgID
	}
	masterSig := signEvent(caller.PrivKey, "MASTER_CLOSED", ns.OrderLedgerID, masterPayload)
	ns.SigLog = append(ns.SigLog, SigRecord{
		ActorID: caller.ID, ActorName: caller.Name, PubKey: caller.PubKey,
		EventType: "MASTER_CLOSED", LedgerID: ns.OrderLedgerID,
		SigPayload: masterPayload, Sig: masterSig,
	})
	doJSONQ(caller.APIKey, caller.Kid, masterSig, "PATCH",
		ns.BaseURL+"/traceledger/master/"+ns.OrderMaster.UUID+"/close", nil)

	// ── Close ORDER ledger ────────────────────────────────────────────────────
	fmt.Println("  Closing ORDER ledger...")
	ledgerCloseIntent := map[string]any{
		"ledger_id":             ns.OrderLedgerID,
		"ledger_type":           "ORDER",
		"requested_by_actor_id": caller.ID,
		"status":                "CLOSED",
		"buyer_actor_id":        ns.Buy.Org.ID,
		"supplier_actor_id":     ns.Sup.Org.ID,
	}
	if ns.Buy.OrgID != "" {
		ledgerCloseIntent["buyer_organization_id"] = ns.Buy.OrgID
	}
	if ns.Sup.OrgID != "" {
		ledgerCloseIntent["supplier_organization_id"] = ns.Sup.OrgID
	}
	ledgerSig := signEvent(caller.PrivKey, "LEDGER_CLOSED", ns.OrderLedgerID, ledgerCloseIntent)
	ns.SigLog = append(ns.SigLog, SigRecord{
		ActorID: caller.ID, ActorName: caller.Name, PubKey: caller.PubKey,
		EventType: "LEDGER_CLOSED", LedgerID: ns.OrderLedgerID,
		SigPayload: ledgerCloseIntent, Sig: ledgerSig,
	})
	doJSONQ(caller.APIKey, caller.Kid, ledgerSig, "PATCH",
		ns.BaseURL+"/ledgers/"+ns.OrderLedgerID+"/close", nil)
	fmt.Println("  ORDER ledger closed.")
}
