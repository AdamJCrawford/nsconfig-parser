package converter

import (
	"fmt"
	"slices"
	"strings"
)

func parseNSConfig(line string) (string, error) {
	parts := strings.Fields(line)
	if idx := slices.Index(parts, "-IPAddress"); idx != -1 {
		return parts[idx+1], nil
	}
	return "", fmt.Errorf("IP not found in line: %v", line)
}

func parseServer(line string) TenantServer {
	// Example line: "add server server ip port"
	parts := strings.Fields(line)
	if len(parts) == 5 {
		return TenantServer{
			ServerName: parts[2],
			ServerIP:   parts[3],
			ServerPort: parts[4],
		}
	} else {
		return TenantServer{
			ServerName: parts[2],
			ServerIP:   parts[3],
		}
	}
}

func handleFTPLBVS(parts []string, method, nsIP string) *VIP {
	fmt.Println("handling: ", parts)
	return &VIP{
		VipName:        parts[3],
		VipServiceType: "FTP",
		VipIP:          parts[6],
		VipPort:        parts[7],
		VipLbMethod:    method,
		ADCIP:          nsIP,
	}
}

func parseLBVS(line, nsIP string) *VIP {
	// Example line: "add lb vserver vserver SSL        ip port -persistenceType persistenceType -timeout 0 -lbMethod method -cookieName cookie -cltTimeout cltTimeout -downStateFlush downStateFlush -devno 179077120"
	// Example line: "add lb vserver vserver SSL_BRIDGE ip port -persistenceType persistenceType            -lbMethod method                    -cltTimeout cltTimeout                                -devno 179175424"
	var method string
	parts := strings.Fields(line)

	if idx := slices.Index(parts, "-lbMethod"); idx == -1 {
		method = "LEASTCONNECTION"
	} else {
		method = parts[idx+1]
	}

	if slices.Contains(parts, "[ftp://FTP") {
		return handleFTPLBVS(parts, method, nsIP)
	}

	return &VIP{
		VipName:        parts[3],
		VipServiceType: parts[4],
		VipIP:          parts[5],
		VipPort:        parts[6],
		VipLbMethod:    method,
		ADCIP:          nsIP,
	}
}

func parseSSL(line string) {
	// Example line: "bind ssl vserver vserver -certkeyName cert"
	// Example line: "bind ssl vserver vserver -certkeyName cert -SNICert"
	var certInfo VipCertkey
	parts := strings.Fields(line)
	if len(parts) == 6 {
		certInfo = VipCertkey{CertKeyName: parts[5], SniCert: "False"}
	} else {
		certInfo = VipCertkey{CertKeyName: parts[5], SniCert: "True"}
	}
	if _, ok := config[parts[3]]; ok {
		config[parts[3]].BoundCertkeys = append(config[parts[3]].BoundCertkeys, certInfo)
	}
}

var config map[string]*VIP = make(map[string]*VIP)

func ParseNetScalerConfig(lines []string) map[string]*VIP {

	servers := make(map[string]TenantServer)
	// serviceGroups := make(map[string]any)
	var nsIP string
	for _, line := range lines {
		if strings.HasPrefix(line, "set ns config") {
			tmpNSIP, err := parseNSConfig(line)
			if err != nil {
				panic("IP address of netscaler not found")
			}
			nsIP = tmpNSIP
		} else if strings.HasPrefix(line, "add server") {
			server := parseServer(line)
			servers[server.ServerName] = server
		} else if strings.HasPrefix(line, "add lb vserver") {
			lbvs := parseLBVS(line, nsIP)
			config[lbvs.VipName] = lbvs
		} else if strings.HasPrefix(line, "bind ssl vserver") {
			parseSSL(line)
		}
	}
	return config
}
