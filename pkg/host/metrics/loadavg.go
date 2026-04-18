package metrics

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// LoadAvg holds parsed /proc/loadavg values.
type LoadAvg struct {
	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}

// ParseLoadAvg parses /proc/loadavg content.
func ParseLoadAvg(content string) (*LoadAvg, error) {
	fields := strings.Fields(strings.TrimSpace(content))
	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid loadavg: %q", content)
	}
	l1, _ := strconv.ParseFloat(fields[0], 64)
	l5, _ := strconv.ParseFloat(fields[1], 64)
	l15, _ := strconv.ParseFloat(fields[2], 64)
	return &LoadAvg{Load1: l1, Load5: l5, Load15: l15}, nil
}

// CPUInfo holds parsed lscpu output.
type CPUInfo struct {
	CPUCount       int    `json:"cpu_count"`
	CoresPerSocket int    `json:"cores_per_socket"`
	Sockets        int    `json:"sockets"`
	ThreadsPerCore int    `json:"threads_per_core"`
	ModelName      string `json:"model_name"`
	Architecture   string `json:"architecture"`
}

// ParseLscpu parses `lscpu` output.
func ParseLscpu(content string) (*CPUInfo, error) {
	info := &CPUInfo{}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "Architecture":
			info.Architecture = val
		case "CPU(s)":
			info.CPUCount, _ = strconv.Atoi(val)
		case "Core(s) per socket":
			info.CoresPerSocket, _ = strconv.Atoi(val)
		case "Socket(s)":
			info.Sockets, _ = strconv.Atoi(val)
		case "Thread(s) per core":
			info.ThreadsPerCore, _ = strconv.Atoi(val)
		case "Model name":
			info.ModelName = val
		}
	}
	return info, nil
}
