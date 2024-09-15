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

func parseServer(line string) Server {
	// Example line: "add server server ip"
	parts := strings.Fields(line)

	return Server{
		ServerName: parts[2],
		ServerIP:   parts[3],
	}

}

func parseSG(line string) ServiceGroup {
	// Example line: "add serviceGroup serviceGroup HTTP -maxClient # -maxReq # -cip DISABLED X-Forwarded-For -usip NO  -useproxyport NO  -cltTimeout # -svrTimeout # -CKA NO  -TCPB YES -CMP NO  -downStateFlush DISABLED -appflowLog DISABLED -devno ########"
	// Example line: "add serviceGroup serviceGroup SSL  -maxClient # -maxReq # -cip ENABLED  X-Forwarded-For -usip YES -useproxyport YES -cltTimeout # -svrTimeout # -CKA YES -TCPB NO  -CMP YES -downStateFlush ENABLED
	//						-tcpProfileName nstcp_large_buffer_default_profile -httpProfileName nshttp_profile_with_websocket -appflowLog ENABLED  -devno ########"
	parts := strings.Fields(line)
	return ServiceGroup{
		Name: parts[2],
	}
}

func handleFTPLBVS(parts []string, method, nsIP string) *VIP {
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
	// Example line: "add lb vserver vserver SSL        ip port -persistenceType persistenceType -timeout 0 -lbMethod method -cookieName cookie -cltTimeout cltTimeout -downStateFlush downStateFlush -devno ########"
	// Example line: "add lb vserver vserver SSL_BRIDGE ip port -persistenceType persistenceType            -lbMethod method                    -cltTimeout cltTimeout                                -devno ########"
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

func parseBindLBVserver(line string) {
	// Example line: "bind lb vserver LBVS SG"
	parts := strings.Fields(line)
	if len(parts) == 5 {
		if _, ok := serviceGroups[parts[4]]; ok {
			port := config[parts[3]].VipPort
			for i := range serviceGroups[parts[4]].Servers {
				serviceGroups[parts[4]].Servers[i].ServerPort = port
			}
			config[parts[3]].VipServers = serviceGroups[parts[4]].Servers
			config[parts[3]].VipMonitors = serviceGroups[parts[4]].Moniors
		}
	}
}

func parseBindServiceGroup(line string) {
	// Example line: bind serviceGroup SG server port      -devno ########"
	// Example line: bind serviceGroup SG -monitorName Mon -devno ########"
	parts := strings.Fields(line)
	if server, ok := servers[parts[3]]; ok {
		if serviceGroups[parts[2]].Servers == nil {
			serviceGroups[parts[2]].Servers = []Server{server}
		} else {
			serviceGroups[parts[2]].Servers = append(serviceGroups[parts[2]].Servers, server)
		}
	}
	if idx := slices.Index(parts, "-monitorName"); idx != -1 {
		serviceGroups[parts[2]].Moniors = append(serviceGroups[parts[2]].Moniors, SGMonitor{
			MonitorName: parts[idx+1],
		})
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
var serviceGroups map[string]*ServiceGroup = make(map[string]*ServiceGroup)
var servers map[string]Server = make(map[string]Server)

func ParseNetScalerConfig(lines []string) map[string]*VIP {
	var nsIP string
	var todo []string
	for _, line := range lines {
		if strings.HasPrefix(line, "set ns config") {
			tmpNSIP, err := parseNSConfig(line)
			if err != nil {
				panic("IP address of netscaler not found")
			}
			nsIP = tmpNSIP
		} else if strings.HasPrefix(line, "add serviceGroup") {
			serviceGroup := parseSG(line)
			serviceGroups[serviceGroup.Name] = &serviceGroup
		} else if strings.HasPrefix(line, "add server") {
			server := parseServer(line)
			servers[server.ServerName] = server
		} else if strings.HasPrefix(line, "add lb vserver") {
			lbvs := parseLBVS(line, nsIP)
			config[lbvs.VipName] = lbvs
		} else if strings.HasPrefix(line, "bind lb vserver") {
			todo = append(todo, line)
		} else if strings.HasPrefix(line, "bind serviceGroup") {
			parseBindServiceGroup(line)
		} else if strings.HasPrefix(line, "bind ssl vserver") {
			parseSSL(line)
		}
	}
	for _, line := range todo {
		parseBindLBVserver(line)
	}
	return config
}
