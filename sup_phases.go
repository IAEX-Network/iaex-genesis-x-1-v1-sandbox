package main

import "fmt"

// ── Supplier party (FarmFresh Ltd) ────────────────────────────────────────────

func supRegister(ns *NetworkState) {
	fmt.Println("\n  [FarmFresh Ltd — Supplier]")
	ns.Sup.Org = registerActor(ns.BaseURL, ns.Region, "FarmFresh Ltd")
	ns.Sup.Prod = registerActor(ns.BaseURL, ns.Region, "sup_prod")
	ns.Sup.QC = registerActor(ns.BaseURL, ns.Region, "sup_qc")
	ns.Sup.Mgr = registerActor(ns.BaseURL, ns.Region, "sup_mgr")
	ns.Sup.Fin = registerActor(ns.BaseURL, ns.Region, "sup_fin")
	ns.Sup.WH = registerActor(ns.BaseURL, ns.Region, "sup_wh")
	ns.Sup.IoT1 = registerActor(ns.BaseURL, ns.Region, "sup_iot1")
	fmt.Printf("    org=%s  (+6 staff)  [keys enrolled at registration]\n", truncate(ns.Sup.Org.ID, 8))
}

func supOnboard(ns *NetworkState) {
	fmt.Println("\n  [FarmFresh Ltd] onboarding...")
	ns.Sup.FRLID, ns.Sup.OrgID = onboardOrg(ns.BaseURL, &ns.Sup.Org, "FarmFresh Ltd", "IN", "FARM"+ns.RunID+"IN", "sup."+ns.RunID+"@e2e.test")
	fmt.Printf("    frl_id=%s  org_id=%s\n", truncate(ns.Sup.FRLID, 8), truncate(ns.Sup.OrgID, 8))
}

func supDelegate(ns *NetworkState) {
	fmt.Println("\n  [FarmFresh Ltd] delegating 6 staff (Zone B/C) — signed...")
	type d struct {
		actor *ActorState
		role  string
	}
	for _, dl := range []d{
		{&ns.Sup.Prod, "OPS"}, {&ns.Sup.QC, "OPS"}, {&ns.Sup.Mgr, "OWNER"},
		{&ns.Sup.Fin, "FINANCE"}, {&ns.Sup.WH, "OPS"}, {&ns.Sup.IoT1, "AGENT"},
	} {
		grantOrgDelegation(ns, ns.Sup.Org, ns.Sup.FRLID, ns.Sup.OrgID, dl.actor.ID, dl.role)
		fmt.Printf("    → %s role=%s [signed]\n", truncate(dl.actor.ID, 8), dl.role)
	}
}

func supCreateFRLMaster(ns *NetworkState) {
	fmt.Println("\n  [FarmFresh Ltd] creating FRL TraceLedgerMaster — signed...")
	businessRef := "sup-stock-" + ns.RunID
	masterUUID := createMaster(ns, ns.Sup.Org, ns.Sup.FRLID, businessRef, "STOCK")
	ns.Sup.Master = MasterState{
		UUID:            masterUUID,
		LedgerID:        ns.Sup.FRLID,
		LedgerType:      "FACILITY_ROOT",
		BusinessRef:     businessRef,
		Scope:           "STOCK",
		SupplierActorID: ns.Sup.Org.ID,
	}
	fmt.Printf("    sup_frl_master=%s [signed]\n", truncate(ns.Sup.Master.UUID, 8))
}

func supGrantFRLMasterAccess(ns *NetworkState) {
	fmt.Println("\n  [FarmFresh Ltd] granting 6 staff write on sup FRL master — signed...")
	for _, a := range []*ActorState{
		&ns.Sup.Prod, &ns.Sup.QC, &ns.Sup.Mgr, &ns.Sup.Fin, &ns.Sup.WH, &ns.Sup.IoT1,
	} {
		grantMasterAccess(ns, ns.Sup.Org, ns.Sup.FRLID, ns.Sup.Master.UUID, a.ID, "OPS")
		fmt.Printf("    → write granted to %s [signed]\n", truncate(a.ID, 8))
	}
}

// supFRLEvents sends all signed supplier FRL business events.
func supFRLEvents(ns *NetworkState) {
	m := ns.Sup.Master

	printBanner("FarmFresh", "sup_prod", "PRODUCTION_STARTED",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Sup.Prod, m, "PRODUCTION_STARTED", map[string]any{
		"product":  "wheat-batch-A1",
		"quantity": 500,
	})
	ns.ProductionHash = getLastEventHash(ns.Sup.Org.APIKey, ns.BaseURL, ns.Sup.FRLID)
	fmt.Printf("  → production_hash=%s\n", truncate(ns.ProductionHash, 20))

	printBanner("FarmFresh", "sup_iot1", "TEMPERATURE_LOG ×3",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	for i, temp := range []float64{4.2, 3.9, 4.1} {
		sendMasterEvent(ns, ns.Sup.IoT1, m, "TEMPERATURE_LOG", map[string]any{
			"sensor_id":    "S-001",
			"temp_celsius": temp,
			"location":     "warehouse-A",
			"reading_seq":  i + 1,
		})
	}

	printBanner("FarmFresh", "sup_qc", "QC_INSPECTION_PASSED (caused_by: production)",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Sup.QC, m, "QC_INSPECTION_PASSED", map[string]any{
		"batch_id":       "wheat-batch-A1",
		"grade":          "A",
		"caused_by_hash": ns.ProductionHash,
	})

	printBanner("FarmFresh", "sup_wh", "SHIPMENT_READY",
		"POST", ns.BaseURL+"/traceledger/master/"+m.UUID+"/events")
	sendMasterEvent(ns, ns.Sup.WH, m, "SHIPMENT_READY", map[string]any{
		"batch_id":     "wheat-batch-A1",
		"pallet_count": 10,
		"ready_at":     "2026-05-01T09:00:00Z",
	})
}
