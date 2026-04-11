// Package network provides network interface, route, DNS, and NTP configuration parsing.
package network

import (
	"strconv"
	"strings"
)

// Interface represents a parsed network interface from `ip addr`.
type Interface struct {
	Name    string   `json:"name"`
	State   string   `json:"state"`   // UP, DOWN
	MTU     string   `json:"mtu"`
	Addrs   []string `json:"addrs"`
	MAC     string   `json:"mac"`
}

// ParseIPAddr parses the output of `ip -brief addr`.
func ParseIPAddr(content string) []Interface {
	var ifaces []Interface
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		iface := Interface{
			Name:  fields[0],
			State: fields[1],
		}

		for _, f := range fields[2:] {
			if strings.Contains(f, "/") {
				iface.Addrs = append(iface.Addrs, f)
			}
		}

		ifaces = append(ifaces, iface)
	}
	return ifaces
}

// ListeningPort represents a parsed listening socket from `ss -tlnp`.
type ListeningPort struct {
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	Port     string `json:"port"`
	Process  string `json:"process"`
}

// ParseSSListening parses the output of `ss -tlnp`.
func ParseSSListening(content string) []ListeningPort {
	var ports []ListeningPort
	for i, line := range strings.Split(content, "\n") {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		local := fields[3]
		lastColon := strings.LastIndex(local, ":")
		addr, port := "", local
		if lastColon >= 0 {
			addr = local[:lastColon]
			port = local[lastColon+1:]
		}

		proc := ""
		if len(fields) > 5 {
			proc = fields[5]
		}

		ports = append(ports, ListeningPort{
			Protocol: fields[0],
			Address:  addr,
			Port:     port,
			Process:  proc,
		})
	}
	return ports
}

// Route represents a parsed route from `ip route show`.
type Route struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Device      string `json:"device"`
	Protocol    string `json:"protocol"`
	Metric      int    `json:"metric"`
}

// ParseIPRoute parses `ip route show` output.
func ParseIPRoute(output string) ([]Route, error) {
	var routes []Route
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 1 {
			continue
		}
		r := Route{Destination: fields[0]}
		for i := 1; i < len(fields)-1; i++ {
			switch fields[i] {
			case "via":
				r.Gateway = fields[i+1]
			case "dev":
				r.Device = fields[i+1]
			case "proto":
				r.Protocol = fields[i+1]
			case "metric":
				r.Metric, _ = strconv.Atoi(fields[i+1])
			}
		}
		routes = append(routes, r)
	}
	return routes, nil
}

// DNSConfig holds parsed /etc/resolv.conf data.
type DNSConfig struct {
	Nameservers   []string `json:"nameservers"`
	SearchDomains []string `json:"search_domains"`
}

// ParseResolvConf parses /etc/resolv.conf content.
func ParseResolvConf(content string) (*DNSConfig, error) {
	cfg := &DNSConfig{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "nameserver":
			cfg.Nameservers = append(cfg.Nameservers, fields[1])
		case "search":
			cfg.SearchDomains = append(cfg.SearchDomains, fields[1:]...)
		}
	}
	return cfg, nil
}

// NTPStatus holds parsed chronyc tracking output.
type NTPStatus struct {
	Server     string  `json:"server"`
	Stratum    int     `json:"stratum"`
	OffsetSec  float64 `json:"offset_sec"`
	LeapStatus string  `json:"leap_status"`
}

// ParseChronyTracking parses `chronyc tracking` output.
func ParseChronyTracking(output string) (*NTPStatus, error) {
	ntp := &NTPStatus{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "Reference ID":
			// Extract server name from parentheses
			if idx := strings.Index(val, "("); idx >= 0 {
				end := strings.Index(val[idx:], ")")
				if end > 0 {
					ntp.Server = val[idx+1 : idx+end]
				}
			}
		case "Stratum":
			ntp.Stratum, _ = strconv.Atoi(val)
		case "System time":
			// "0.000012345 seconds fast of NTP time"
			fields := strings.Fields(val)
			if len(fields) > 0 {
				ntp.OffsetSec, _ = strconv.ParseFloat(fields[0], 64)
			}
		case "Leap status":
			ntp.LeapStatus = val
		}
	}
	return ntp, nil
}
