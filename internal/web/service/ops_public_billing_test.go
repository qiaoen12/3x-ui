package service

import (
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service/outbound"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

func TestOpsPublicBillingKeepsHistoryAndScalesEveryNewClientDelta(t *testing.T) {
	setupBulkDB(t)
	db := database.GetDB()
	expiry := time.Now().Add(90 * 24 * time.Hour).UnixMilli()
	for _, row := range []model.ClientRecord{
		{Email: "legacy-public", Comment: "ops-public-account:550220377:250GiB:1month", TotalGB: 1000, ExpiryTime: expiry, Enable: true, TrafficReset: "monthly", TrafficResetDay: 26},
		{Email: "new-public", Comment: "ops-public-account:new", TotalGB: 1000, ExpiryTime: expiry, Enable: true},
		{Email: "personal", Comment: "my-device", TotalGB: 1000, ExpiryTime: expiry, Enable: true},
	} {
		if err := db.Create(&row).Error; err != nil {
			t.Fatal(err)
		}
		mkTraffic(t, 999, row.Email, 100, 200, row.TotalGB, expiry, true)
	}
	samples := []*xray.ClientTraffic{
		{Email: "legacy-public", Up: 10, Down: 20},
		{Email: "new-public", Up: 10, Down: 20},
		{Email: "personal", Up: 10, Down: 20},
	}
	service := &InboundService{}
	for tick := int64(1); tick <= 2; tick++ {
		if err := service.addClientTraffic(db, samples); err != nil {
			t.Fatal(err)
		}
		for _, email := range []string{"legacy-public", "new-public", "personal"} {
			multiplier := int64(2)
			row := trafficOf(t, email)
			if row.Up != 100+tick*10*multiplier || row.Down != 200+tick*20*multiplier {
				t.Fatalf("%s tick %d: usage %d/%d", email, tick, row.Up, row.Down)
			}
			if row.Total != 1000 || row.ExpiryTime != expiry {
				t.Fatal("billing changed the quota or expiration")
			}
		}
		for _, sample := range samples {
			if sample.Up != 10 || sample.Down != 20 {
				t.Fatal("raw samples were mutated or multiplied a second time")
			}
		}
	}
	var original model.ClientRecord
	if err := db.Where("email = ?", "legacy-public").First(&original).Error; err != nil {
		t.Fatal(err)
	}
	if original.TrafficReset != "monthly" || original.TrafficResetDay != 26 || original.ExpiryTime != expiry {
		t.Fatal("billing changed monthly reset or expiration")
	}
	if _, err := (&ClientService{}).ResetAllTraffics(); err != nil {
		t.Fatal(err)
	}
	if err := service.addClientTraffic(db, samples); err != nil {
		t.Fatal(err)
	}
	for _, email := range []string{"legacy-public", "new-public", "personal"} {
		row := trafficOf(t, email)
		if row.Up != 20 || row.Down != 40 {
			t.Fatalf("%s new cycle: %d/%d", email, row.Up, row.Down)
		}
	}
}

func TestOpsPublicBillingReachesNativeQuotaAtHalfRawUsage(t *testing.T) {
	setupBulkDB(t)
	client := model.Client{Email: "public-quota", ID: "11111111-1111-1111-1111-111111111111", Enable: true, TotalGB: 100, Comment: "ops-public-account:quota"}
	inbound := mkInbound(t, 46390, model.VLESS, clientsSettings(t, []model.Client{client}))
	if err := (&ClientService{}).SyncInbound(nil, inbound.Id, []model.Client{client}); err != nil {
		t.Fatal(err)
	}
	mkTraffic(t, inbound.Id, client.Email, 0, 0, 100, 0, true)
	service := &InboundService{}
	if _, _, err := service.AddTraffic(nil, []*xray.ClientTraffic{{Email: client.Email, Down: 49}}); err != nil {
		t.Fatal(err)
	}
	if row := trafficOf(t, client.Email); row.Down != 98 || !row.Enable {
		t.Fatalf("client stopped early or usage is wrong: %+v", row)
	}
	if _, _, err := service.AddTraffic(nil, []*xray.ClientTraffic{{Email: client.Email, Down: 1}}); err != nil {
		t.Fatal(err)
	}
	if row := trafficOf(t, client.Email); row.Down != 100 || row.Enable {
		t.Fatalf("native quota did not stop the client at billed 100: %+v", row)
	}
}

func TestOpsPublicBillingDeltaSaturates(t *testing.T) {
	if got := opsPublicBillingDelta(database.TrafficMax); got != database.TrafficMax {
		t.Fatalf("overflow: %d", got)
	}
	if got := opsPublicBillingDelta(17); got != 34 {
		t.Fatalf("unlabelled traffic was not doubled: %d", got)
	}
}

func TestOpsPublicBillingInboundHistoryRawSamplesAndOutbound(t *testing.T) {
	setupBulkDB(t)
	db := database.GetDB()
	ib := mkInbound(t, 46391, model.VLESS, clientsSettings(t, nil))
	if err := db.Model(ib).Updates(map[string]any{"up": 100, "down": 200}).Error; err != nil {
		t.Fatal(err)
	}
	samples := []*xray.Traffic{{Tag: ib.Tag, IsInbound: true, Up: 10, Down: 20}, {Tag: ib.Tag, IsOutbound: true, Up: 500, Down: 700}}
	for tick := int64(1); tick <= 2; tick++ {
		if err := (&InboundService{}).addInboundTraffic(db, samples); err != nil {
			t.Fatal(err)
		}
		var got model.Inbound
		if err := db.First(&got, ib.Id).Error; err != nil {
			t.Fatal(err)
		}
		if got.Up != 100+tick*20 || got.Down != 200+tick*40 {
			t.Fatalf("inbound: %d/%d", got.Up, got.Down)
		}
		if err, _ := (&outbound.OutboundService{}).AddTraffic(samples, nil); err != nil {
			t.Fatal(err)
		}
		var out model.OutboundTraffics
		if err := db.Where("tag = ?", ib.Tag).First(&out).Error; err != nil {
			t.Fatal(err)
		}
		if out.Up != tick*500 || out.Down != tick*700 {
			t.Fatalf("outbound was scaled: %+v", out)
		}
		if samples[0].Up != 10 || samples[0].Down != 20 || samples[1].Up != 500 || samples[1].Down != 700 {
			t.Fatal("raw/outbound samples mutated")
		}
	}
	if err := db.Model(ib).Updates(map[string]any{"up": database.TrafficMax - 5, "down": database.TrafficMax - 5}).Error; err != nil {
		t.Fatal(err)
	}
	if err := (&InboundService{}).addInboundTraffic(db, samples); err != nil {
		t.Fatal(err)
	}
	var got model.Inbound
	if err := db.First(&got, ib.Id).Error; err != nil {
		t.Fatal(err)
	}
	if got.Up != database.TrafficMax || got.Down != database.TrafficMax {
		t.Fatal("inbound did not saturate")
	}
}

func TestOpsPublicBillingInboundQuotaAtHalfRawUsage(t *testing.T) {
	setupBulkDB(t)
	db := database.GetDB()
	ib := mkInbound(t, 46392, model.VLESS, clientsSettings(t, nil))
	if err := db.Model(ib).Update("total", 100).Error; err != nil {
		t.Fatal(err)
	}
	svc := &InboundService{}
	for _, n := range []int64{49, 1} {
		sample := &xray.Traffic{Tag: ib.Tag, IsInbound: true, Down: n}
		if _, _, err := svc.AddTraffic([]*xray.Traffic{sample}, nil); err != nil {
			t.Fatal(err)
		}
		var got model.Inbound
		if err := db.First(&got, ib.Id).Error; err != nil {
			t.Fatal(err)
		}
		if n == 49 && (got.Down != 98 || !got.Enable) {
			t.Fatalf("early disable: %+v", got)
		}
		if n == 1 && (got.Down != 100 || got.Enable) {
			t.Fatalf("quota not enforced: %+v", got)
		}
		if sample.Down != n {
			t.Fatal("raw sample mutated")
		}
	}
}

func TestOpsPublicBillingClientCounterSaturates(t *testing.T) {
	setupBulkDB(t)
	db := database.GetDB()
	mkTraffic(t, 999, "unlabelled", database.TrafficMax-5, database.TrafficMax-5, 0, 0, true)
	if err := (&InboundService{}).addClientTraffic(db, []*xray.ClientTraffic{{Email: "unlabelled", Up: 10, Down: database.TrafficMax}}); err != nil {
		t.Fatal(err)
	}
	got := trafficOf(t, "unlabelled")
	if got.Up != database.TrafficMax || got.Down != database.TrafficMax {
		t.Fatal("client did not saturate")
	}
}
