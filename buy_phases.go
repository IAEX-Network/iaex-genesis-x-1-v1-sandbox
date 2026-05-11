package main

import "fmt"

// ── Buyer party (RetailCo Ltd) ─────────────────────────────────────────────────

func buyRegister(ns *NetworkState) {
	fmt.Println("\n  [RetailCo Ltd — Buyer]")
	ns.Buy.Org = registerActor(ns.BaseURL, ns.Region, "RetailCo Ltd")
	ns.Buy.Proc = registerActor(ns.BaseURL, ns.Region, "buy_proc")
	ns.Buy.QC = registerActor(ns.BaseURL, ns.Region, "buy_qc")
	ns.Buy.Fin = registerActor(ns.BaseURL, ns.Region, "buy_fin")
	ns.Buy.Acc = registerActor(ns.BaseURL, ns.Region, "buy_acc")
	ns.Buy.WH = registerActor(ns.BaseURL, ns.Region, "buy_wh")
	ns.Buy.IoT1 = registerActor(ns.BaseURL, ns.Region, "buy_iot1")
	fmt.Printf("    org=%s  (+6 staff)  [keys enrolled at registration]\n", truncate(ns.Buy.Org.ID, 8))
}

func buyOnboard(ns *NetworkState) {
	fmt.Println("\n  [RetailCo Ltd] onboarding...")
	ns.Buy.FRLID, ns.Buy.OrgID = onboardOrg(ns.BaseURL, &ns.Buy.Org, "RetailCo Ltd", "IN", "RTCO"+ns.RunID+"IN", "buy."+ns.RunID+"@e2e.test")
	fmt.Printf("    frl_id=%s  org_id=%s\n", truncate(ns.Buy.FRLID, 8), truncate(ns.Buy.OrgID, 8))
}

func buyDelegate(ns *NetworkState) {
	fmt.Println("\n  [RetailCo Ltd] delegating 6 staff (Zone B/C) — signed...")
	type d struct {
		actor *ActorState
		role  string
	}
	for _, dl := range []d{
		{&ns.Buy.Proc, "OPS"}, {&ns.Buy.QC, "OPS"}, {&ns.Buy.Fin, "FINANCE"},
		{&ns.Buy.Acc, "OPS"}, {&ns.Buy.WH, "OPS"}, {&ns.Buy.IoT1, "AGENT"},
	} {
		grantOrgDelegation(ns, ns.Buy.Org, ns.Buy.FRLID, ns.Buy.OrgID, dl.actor.ID, dl.role)
		fmt.Printf("    → %s role=%s [signed]\n", truncate(dl.actor.ID, 8), dl.role)
	}
}

func buyCreateFRLMaster(ns *NetworkState) {
	fmt.Println("\n  [RetailCo Ltd] creating FRL TraceLedgerMaster — signed...")
	businessRef := "buy-stock-" + ns.RunID
	masterUUID := createMaster(ns, ns.Buy.Org, ns.Buy.FRLID, businessRef, "STOCK")
	ns.Buy.Master = MasterState{
		UUID:            masterUUID,
		LedgerID:        ns.Buy.FRLID,
		LedgerType:      "FACILITY_ROOT",
		BusinessRef:     businessRef,
		Scope:           "STOCK",
		SupplierActorID: ns.Buy.Org.ID,
	}
	fmt.Printf("    buy_frl_master=%s [signed]\n", truncate(ns.Buy.Master.UUID, 8))
}

func buyGrantFRLMasterAccess(ns *NetworkState) {
	fmt.Println("\n  [RetailCo Ltd] granting 6 staff write on buy FRL master — signed...")
	for _, a := range []*ActorState{
		&ns.Buy.Proc, &ns.Buy.QC, &ns.Buy.Fin, &ns.Buy.Acc, &ns.Buy.WH, &ns.Buy.IoT1,
	} {
		grantMasterAccess(ns, ns.Buy.Org, ns.Buy.FRLID, ns.Buy.Master.UUID, a.ID, "OPS")
		fmt.Printf("    → write granted to %s [signed]\n", truncate(a.ID, 8))
	}
}

// buyFRLEvents sends all signed buyer FRL business events.
func buyFRLEvents(ns *NetworkState) {
	m := ns.Buy.Master

	printBanner("RetailCo", "buy_acc", "GOODS_ACCEPTED (caused_by: delivered)",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Buy.Acc, m, "GOODS_ACCEPTED", map[string]any{
		"pallet_count":   10,
		"caused_by_hash": ns.DeliveredHash,
	})

	printBanner("RetailCo", "buy_iot1", "DOCK_SCAN_COMPLETE",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Buy.IoT1, m, "DOCK_SCAN_COMPLETE", map[string]any{
		"scan_id":         "DS-001",
		"barcode_matched": true,
	})

	printBanner("RetailCo", "buy_qc", "BUYER_QC_PASSED",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Buy.QC, m, "BUYER_QC_PASSED", map[string]any{
		"batch_id":      "wheat-batch-A1",
		"sample_tested": 50,
		"passed":        50,
	})

	printBanner("RetailCo", "buy_wh", "WAREHOUSE_RECEIVED",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Buy.WH, m, "WAREHOUSE_RECEIVED", map[string]any{
		"bin_location":    "BIN-A7",
		"quantity_stored": 500,
	})
}
