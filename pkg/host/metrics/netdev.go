package metrics

import (
	"fmt"
	"strconv"
	"strings"
)

// NetDevMetrics holds per-interface network throughput from /proc/net/dev delta.
type NetDevMetrics struct {
	Interface       string `json:"interface"`
	RxBytesPerSec   uint64 `json:"rx_bytes_per_sec"`
	TxBytesPerSec   uint64 `json:"tx_bytes_per_sec"`
	RxPacketsPerSec uint64 `json:"rx_packets_per_sec"`
	TxPacketsPerSec uint64 `json:"tx_packets_per_sec"`
	RxErrors        uint64 `json:"rx_errors"`
	TxErrors        uint64 `json:"tx_errors"`
	RxDrops         uint64 `json:"rx_drops"`
	TxDrops         uint64 `json:"tx_drops"`
}

type netDevEntry struct {
	rxBytes   uint64
	rxPackets uint64
	rxErrors  uint64
	rxDrops   uint64
	txBytes   uint64
	txPackets uint64
	txErrors  uint64
	txDrops   uint64
}

// ParseNetDevDelta computes per-interface throughput from two /proc/net/dev snapshots.
func ParseNetDevDelta(before, after string, intervalSec float64) (map[string]NetDevMetrics, error) {
	bMap, err := parseNetDev(before)
	if err != nil {
		return nil, fmt.Errorf("parse before: %w", err)
	}
	aMap, err := parseNetDev(after)
	if err != nil {
		return nil, fmt.Errorf("parse after: %w", err)
	}

	result := make(map[string]NetDevMetrics)
	for iface, a := range aMap {
		b, ok := bMap[iface]
		if !ok {
			continue
		}
		result[iface] = NetDevMetrics{
			Interface:       iface,
			RxBytesPerSec:   uint64(float64(a.rxBytes-b.rxBytes) / intervalSec),
			TxBytesPerSec:   uint64(float64(a.txBytes-b.txBytes) / intervalSec),
			RxPacketsPerSec: uint64(float64(a.rxPackets-b.rxPackets) / intervalSec),
			TxPacketsPerSec: uint64(float64(a.txPackets-b.txPackets) / intervalSec),
			RxErrors:        a.rxErrors - b.rxErrors,
			TxErrors:        a.txErrors - b.txErrors,
			RxDrops:         a.rxDrops - b.rxDrops,
			TxDrops:         a.txDrops - b.txDrops,
		}
	}
	return result, nil
}

func parseNetDev(content string) (map[string]netDevEntry, error) {
	result := make(map[string]netDevEntry)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Inter") || strings.Contains(line, "|") {
			continue
		}
		// Format: "iface: rxBytes rxPackets rxErrors rxDrops ... txBytes txPackets txErrors txDrops ..."
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}
		rxBytes, _ := strconv.ParseUint(fields[0], 10, 64)
		rxPackets, _ := strconv.ParseUint(fields[1], 10, 64)
		rxErrors, _ := strconv.ParseUint(fields[2], 10, 64)
		rxDrops, _ := strconv.ParseUint(fields[3], 10, 64)
		txBytes, _ := strconv.ParseUint(fields[8], 10, 64)
		txPackets, _ := strconv.ParseUint(fields[9], 10, 64)
		txErrors, _ := strconv.ParseUint(fields[10], 10, 64)
		txDrops, _ := strconv.ParseUint(fields[11], 10, 64)

		result[iface] = netDevEntry{
			rxBytes: rxBytes, rxPackets: rxPackets, rxErrors: rxErrors, rxDrops: rxDrops,
			txBytes: txBytes, txPackets: txPackets, txErrors: txErrors, txDrops: txDrops,
		}
	}
	return result, nil
}
