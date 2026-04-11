// Package network provides network interface and configuration parsing.
package network

import "strings"

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
