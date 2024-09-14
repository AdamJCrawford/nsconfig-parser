package converter

type SGMonitor struct {
	MonitorName string `json:"monitor_name"`
	MonState    string `json:"monstate"`
	State       string `json:"state"`
}

type Response struct {
	MonitorName  string
	LastResponse string
}

type Server struct { // actual server
	ServerName      string `json:"server_name"`
	ServerIP        string `json:"ip"`
	ServerPort      string `json:"port"`
	State           string `json:"state"`
	ServerState     string `json:"svrstate"`
	StateChangeTime string `json:"statechangetimesec"`
	Ticks           string `json:"ticksincelaststatechange"`
	DomainName      string
	Responses       []Response
	SplunkSearch    string
}

type VipCertkey struct {
	CertKeyName string `json:"certkeyname"`
	SniCert     string `json:"snicert"` // Bool string
}

type VIP struct { //lbvs + csvs
	VipName        string
	VipIP          string
	VipPort        string
	VipLbMethod    string //ROUNDROBIN, LEASTCONNECTION, etc
	VipState       string // DOWN or UP
	VipServiceType string // SSL, SSL_BRIDGE
	ADCIP          string // IP of the box
	VipMonitors    []SGMonitor
	VipServers     []Server
	BoundCertkeys  []VipCertkey
}
