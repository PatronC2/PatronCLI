package types

type Profile struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Port      string `json:"port"`
	Username  string `json:"username"`
	LoginTime int    `json:"login_time"`
}

type Credential struct {
	Profile string `json:"profile"`
	IP      string `json:"ip"`
	Port    string `json:"port"`
	Token   string `json:"token"`
}
