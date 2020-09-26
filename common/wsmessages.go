package common

type ControlMsg struct {
	ConType string `json:"ConType"`
}

type UDPControlMsg struct {
	ControlMsg
	IP   string
	Port int
	Zone string // IPv6 scoped addressing zone
}
