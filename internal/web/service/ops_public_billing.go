package service

import "github.com/mhsanaei/3x-ui/v3/internal/database"

// Bill every local client and inbound at 2x for new deltas only. Keep historical
// counters and raw Xray samples intact; outbound statistics retain native units.
func opsPublicBillingDelta(value int64) int64 {
	if value <= 0 {
		return value
	}
	if value > database.TrafficMax/2 {
		return database.TrafficMax
	}
	return value * 2
}
