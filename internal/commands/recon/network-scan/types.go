package networkscan

import "time"

type ScanConfig struct {
	Target string

	TCP bool
	UDP bool

	ServiceDetection bool
	VersionDetection bool
	OSDetection      bool

	Verbose    bool
	OpenOnly   bool
	ShowReason bool

	PortSpecification string
	AllPorts          bool
	Ports             []int

	Timeout time.Duration
	Timing  int
	Workers int

	OutputFile string
	OutputType string
}

type PortResult struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`

	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`

	Service string `json:"service,omitempty"`
	Product string `json:"product,omitempty"`
	Version string `json:"version,omitempty"`

	Confidence int `json:"confidence,omitempty"`

	Banner string `json:"banner,omitempty"`
}

type OSResult struct {
	Family     string `json:"family"`
	Name       string `json:"name"`
	Version    string `json:"version,omitempty"`
	Confidence int    `json:"confidence"`

	Evidence []string `json:"evidence,omitempty"`
}

type ScanResult struct {
	Target         string
	StartedAt      time.Time
	Duration       time.Duration
	Ports          []PortResult
	TotalPorts     int
	OpenPorts      int
	FilterAnalysis FilterAnalysis
	OS             OSInfo
}
