package networkscan

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"os"
	"strings"
	"time"
)

type ScanReport struct {
	Target         string         `json:"target" xml:"target"`
	StartedAt      time.Time      `json:"started_at" xml:"started_at"`
	Duration       string         `json:"duration" xml:"duration"`
	TotalPorts     int            `json:"total_ports" xml:"total_ports"`
	OpenPorts      int            `json:"open_ports" xml:"open_ports"`
	FilterAnalysis FilterAnalysis `json:"filter_analysis" xml:"filter_analysis"`
	OS             OSInfo         `json:"os" xml:"os"`
	Ports          []PortResult   `json:"ports" xml:"ports>port"`
}

type XMLReport struct {
	XMLName    xml.Name  `xml:"zebra-scan"`
	Target     string    `xml:"target,attr"`
	StartedAt  string    `xml:"started_at,attr"`
	Duration   string    `xml:"duration,attr"`
	TotalPorts int       `xml:"total_ports,attr"`
	OpenPorts  int       `xml:"open_ports,attr"`
	OS         XMLOS     `xml:"os"`
	Filter     XMLFilter `xml:"filter"`
	Ports      []XMLPort `xml:"ports>port"`
}

type XMLOS struct {
	Name       string   `xml:"name"`
	Family     string   `xml:"family"`
	Version    string   `xml:"version"`
	Confidence int      `xml:"confidence"`
	Evidence   []string `xml:"evidence>item"`
}

type XMLFilter struct {
	Open         int      `xml:"open"`
	Closed       int      `xml:"closed"`
	Filtered     int      `xml:"filtered"`
	OpenFiltered int      `xml:"open_filtered"`
	Unknown      int      `xml:"unknown"`
	Observations []string `xml:"observations>item"`
}

type XMLPort struct {
	Port       int    `xml:"port,attr"`
	Protocol   string `xml:"protocol,attr"`
	State      string `xml:"state,attr"`
	Service    string `xml:"service,omitempty"`
	Product    string `xml:"product,omitempty"`
	Version    string `xml:"version,omitempty"`
	Confidence int    `xml:"confidence,omitempty"`
	Reason     string `xml:"reason,omitempty"`
	Banner     string `xml:"banner,omitempty"`
}

func WriteScanReport(
	config ScanConfig,
	result ScanResult,
) error {
	if strings.TrimSpace(config.OutputFile) == "" {
		return nil
	}

	switch config.OutputType {
	case "text":
		return writeTextReport(
			config,
			result,
		)

	case "json":
		return writeJSONReport(
			result,
			config.OutputFile,
		)

	case "xml":
		return writeXMLReport(
			result,
			config.OutputFile,
		)

	case "html":
		return writeHTMLReport(
			config,
			result,
			config.OutputFile,
		)

	default:
		return fmt.Errorf(
			"unsupported output type: %q",
			config.OutputType,
		)
	}
}

func writeTextReport(
	config ScanConfig,
	result ScanResult,
) error {
	var builder strings.Builder

	builder.WriteString("ZEBRA Network Scan\n")
	builder.WriteString("==================\n\n")

	builder.WriteString(
		fmt.Sprintf(
			"Target:        %s\n",
			result.Target,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"Started:       %s\n",
			result.StartedAt.Format(time.RFC3339),
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"Duration:      %s\n",
			result.Duration.Round(time.Millisecond),
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"Ports scanned: %d\n",
			result.TotalPorts,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"Open:          %d\n",
			result.FilterAnalysis.Open,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"Closed:        %d\n",
			result.FilterAnalysis.Closed,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"Filtered:      %d\n",
			result.FilterAnalysis.Filtered,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"Open|Filtered: %d\n",
			result.FilterAnalysis.OpenFiltered,
		),
	)

	builder.WriteString("\nScan Types\n")
	builder.WriteString("----------\n")

	if config.TCP {
		builder.WriteString("TCP: enabled\n")
	}

	if config.UDP {
		builder.WriteString("UDP: enabled\n")
	}

	if config.ServiceDetection {
		builder.WriteString(
			"Service Detection: enabled\n",
		)
	}

	if config.VersionDetection {
		builder.WriteString(
			"Version Detection: enabled\n",
		)
	}

	if config.OSDetection {
		builder.WriteString(
			"OS Detection: enabled\n",
		)
	}

	builder.WriteString("\nPort Results\n")
	builder.WriteString("------------\n")

	for _, port := range result.Ports {
		if config.OpenOnly && port.State != "open" {
			continue
		}

		builder.WriteString(
			formatReportPort(port),
		)

		builder.WriteString("\n")
	}

	if config.OSDetection {
		builder.WriteString("\nOS Detection\n")
		builder.WriteString("------------\n")
		builder.WriteString(
			fmt.Sprintf(
				"OS: %s\n",
				result.OS.Name,
			),
		)
		builder.WriteString(
			fmt.Sprintf(
				"Family: %s\n",
				result.OS.Family,
			),
		)
		builder.WriteString(
			fmt.Sprintf(
				"Version: %s\n",
				result.OS.Version,
			),
		)
		builder.WriteString(
			fmt.Sprintf(
				"Confidence: %d%%\n",
				result.OS.Confidence,
			),
		)
	}

	if len(result.FilterAnalysis.Observations) > 0 {
		builder.WriteString("\nObservations\n")
		builder.WriteString("------------\n")

		for _, observation := range result.FilterAnalysis.Observations {
			builder.WriteString(
				fmt.Sprintf(
					"- %s\n",
					observation,
				),
			)
		}
	}

	return writeFile(
		config.OutputFile,
		[]byte(builder.String()),
	)
}

func writeJSONReport(
	result ScanResult,
	filename string,
) error {
	report := ScanReport{
		Target:         result.Target,
		StartedAt:      result.StartedAt,
		Duration:       result.Duration.Round(time.Millisecond).String(),
		TotalPorts:     result.TotalPorts,
		OpenPorts:      result.OpenPorts,
		FilterAnalysis: result.FilterAnalysis,
		OS:             result.OS,
		Ports:          result.Ports,
	}

	data, err := json.MarshalIndent(
		report,
		"",
		"  ",
	)

	if err != nil {
		return fmt.Errorf(
			"failed to encode JSON report: %w",
			err,
		)
	}

	data = append(data, '\n')

	return writeFile(
		filename,
		data,
	)
}

func writeXMLReport(
	result ScanResult,
	filename string,
) error {
	report := XMLReport{
		Target:     result.Target,
		StartedAt:  result.StartedAt.Format(time.RFC3339),
		Duration:   result.Duration.Round(time.Millisecond).String(),
		TotalPorts: result.TotalPorts,
		OpenPorts:  result.OpenPorts,

		OS: XMLOS{
			Name:       result.OS.Name,
			Family:     result.OS.Family,
			Version:    result.OS.Version,
			Confidence: result.OS.Confidence,
			Evidence:   result.OS.Evidence,
		},

		Filter: XMLFilter{
			Open:         result.FilterAnalysis.Open,
			Closed:       result.FilterAnalysis.Closed,
			Filtered:     result.FilterAnalysis.Filtered,
			OpenFiltered: result.FilterAnalysis.OpenFiltered,
			Unknown:      result.FilterAnalysis.Unknown,
			Observations: result.FilterAnalysis.Observations,
		},
	}

	for _, port := range result.Ports {
		report.Ports = append(
			report.Ports,
			XMLPort{
				Port:       port.Port,
				Protocol:   port.Protocol,
				State:      port.State,
				Service:    port.Service,
				Product:    port.Product,
				Version:    port.Version,
				Confidence: port.Confidence,
				Reason:     port.Reason,
				Banner:     port.Banner,
			},
		)
	}

	data, err := xml.MarshalIndent(
		report,
		"",
		"  ",
	)

	if err != nil {
		return fmt.Errorf(
			"failed to encode XML report: %w",
			err,
		)
	}

	data = append(
		[]byte(xml.Header),
		data...,
	)

	data = append(data, '\n')

	return writeFile(
		filename,
		data,
	)
}

func writeHTMLReport(
	config ScanConfig,
	result ScanResult,
	filename string,
) error {
	var builder strings.Builder

	builder.WriteString(
		"<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n",
	)

	builder.WriteString(
		"<meta charset=\"UTF-8\">\n",
	)

	builder.WriteString(
		"<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n",
	)

	builder.WriteString(
		"<title>ZEBRA Network Scan</title>\n",
	)

	builder.WriteString(`
<style>
body {
	font-family: Arial, sans-serif;
	margin: 40px;
	background: #f5f5f5;
	color: #222;
}
.container {
	max-width: 1200px;
	margin: auto;
	background: white;
	padding: 30px;
	border-radius: 10px;
}
table {
	width: 100%;
	border-collapse: collapse;
	margin-top: 20px;
}
th, td {
	padding: 10px;
	border: 1px solid #ddd;
	text-align: left;
}
th {
	background: #eee;
}
.open {
	font-weight: bold;
}
.section {
	margin-top: 30px;
}
</style>
</head>
<body>
<div class="container">
`)

	builder.WriteString(
		"<h1>ZEBRA Network Scan</h1>\n",
	)

	builder.WriteString(
		fmt.Sprintf(
			"<p><strong>Target:</strong> %s</p>\n",
			html.EscapeString(result.Target),
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"<p><strong>Started:</strong> %s</p>\n",
			html.EscapeString(
				result.StartedAt.Format(time.RFC3339),
			),
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"<p><strong>Duration:</strong> %s</p>\n",
			html.EscapeString(
				result.Duration.Round(time.Millisecond).String(),
			),
		),
	)

	builder.WriteString(
		"<div class=\"section\"><h2>Summary</h2>",
	)

	builder.WriteString("<ul>")

	builder.WriteString(
		fmt.Sprintf(
			"<li>Total ports: %d</li>",
			result.TotalPorts,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"<li>Open: %d</li>",
			result.FilterAnalysis.Open,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"<li>Closed: %d</li>",
			result.FilterAnalysis.Closed,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"<li>Filtered: %d</li>",
			result.FilterAnalysis.Filtered,
		),
	)

	builder.WriteString(
		fmt.Sprintf(
			"<li>Open|Filtered: %d</li>",
			result.FilterAnalysis.OpenFiltered,
		),
	)

	builder.WriteString("</ul></div>")

	builder.WriteString(
		"<div class=\"section\"><h2>Port Results</h2>",
	)

	builder.WriteString(`
<table>
<thead>
<tr>
<th>Port</th>
<th>Protocol</th>
<th>State</th>
<th>Service</th>
<th>Product</th>
<th>Version</th>
<th>Confidence</th>
<th>Reason</th>
</tr>
</thead>
<tbody>
`)

	for _, port := range result.Ports {
		if config.OpenOnly && port.State != "open" {
			continue
		}

		builder.WriteString("<tr>")

		builder.WriteString(
			fmt.Sprintf(
				"<td>%d</td>",
				port.Port,
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"<td>%s</td>",
				html.EscapeString(port.Protocol),
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"<td class=\"%s\">%s</td>",
				html.EscapeString(port.State),
				html.EscapeString(port.State),
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"<td>%s</td>",
				html.EscapeString(port.Service),
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"<td>%s</td>",
				html.EscapeString(port.Product),
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"<td>%s</td>",
				html.EscapeString(port.Version),
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"<td>%d%%</td>",
				port.Confidence,
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"<td>%s</td>",
				html.EscapeString(port.Reason),
			),
		)

		builder.WriteString("</tr>\n")
	}

	builder.WriteString("</tbody></table></div>")

	if config.OSDetection {
		builder.WriteString(
			"<div class=\"section\"><h2>OS Detection</h2>",
		)

		builder.WriteString(
			fmt.Sprintf(
				"<p><strong>OS:</strong> %s</p>",
				html.EscapeString(result.OS.Name),
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"<p><strong>Family:</strong> %s</p>",
				html.EscapeString(result.OS.Family),
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"<p><strong>Version:</strong> %s</p>",
				html.EscapeString(result.OS.Version),
			),
		)

		builder.WriteString(
			fmt.Sprintf(
				"<p><strong>Confidence:</strong> %d%%</p>",
				result.OS.Confidence,
			),
		)

		builder.WriteString("</div>")
	}

	builder.WriteString(
		"</div>\n</body>\n</html>\n",
	)

	return writeFile(
		filename,
		[]byte(builder.String()),
	)
}

func formatReportPort(
	result PortResult,
) string {
	var builder strings.Builder

	builder.WriteString(
		fmt.Sprintf(
			"%5d/%-3s %-14s",
			result.Port,
			result.Protocol,
			result.State,
		),
	)

	if result.Service != "" {
		builder.WriteString(
			fmt.Sprintf(
				" service=%s",
				result.Service,
			),
		)
	}

	if result.Product != "" {
		builder.WriteString(
			fmt.Sprintf(
				" product=%s",
				result.Product,
			),
		)
	}

	if result.Version != "" {
		builder.WriteString(
			fmt.Sprintf(
				" version=%s",
				result.Version,
			),
		)
	}

	if result.Confidence > 0 {
		builder.WriteString(
			fmt.Sprintf(
				" confidence=%d%%",
				result.Confidence,
			),
		)
	}

	if result.Reason != "" {
		builder.WriteString(
			fmt.Sprintf(
				" reason=%s",
				result.Reason,
			),
		)
	}

	return builder.String()
}

func writeFile(
	filename string,
	data []byte,
) error {
	filename = strings.TrimSpace(filename)

	if filename == "" {
		return fmt.Errorf(
			"output filename cannot be empty",
		)
	}

	if err := os.WriteFile(
		filename,
		data,
		0644,
	); err != nil {
		return fmt.Errorf(
			"failed to write report %q: %w",
			filename,
			err,
		)
	}

	fmt.Printf(
		"\nReport written to: %s\n",
		filename,
	)

	return nil
}
