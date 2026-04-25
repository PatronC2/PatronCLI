package types

type Profile struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Port      string `json:"port"`
	Username  string `json:"username"`
	LoginTime int    `json:"loginTime"`
	// Optional SOCKS5 proxy support
	SOCKS5Enabled  bool   `json:"socks5Enabled"`
	SOCKS5Host     string `json:"socks5Host,omitempty"`
	SOCKS5Port     string `json:"socks5Port,omitempty"`
	SOCKS5Username string `json:"socks5Username,omitempty"`
	SOCKS5Password string `json:"socks5Password,omitempty"`
}

type Credential struct {
	Profile string `json:"profile"`
	IP      string `json:"ip"`
	Port    string `json:"port"`
	Token   string `json:"token"`
	// Optional SOCKS5 proxy support
	SOCKS5Enabled  bool   `json:"socks5Enabled"`
	SOCKS5Host     string `json:"socks5Host,omitempty"`
	SOCKS5Port     string `json:"socks5Port,omitempty"`
	SOCKS5Username string `json:"socks5Username,omitempty"`
	SOCKS5Password string `json:"socks5Password,omitempty"`
}
