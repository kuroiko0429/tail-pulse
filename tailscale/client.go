package tailscale

import (
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

type Status struct {
	Peer map[string]PeerStatus `json:"Peer"`
	Self PeerStatus            `json:"Self"`
	ExitNodeStatus *ExitNodeStatus `json:"ExitNodeStatus,omitempty"`
}

type ExitNodeStatus struct {
	ID string `json:"ID"`
	Online bool `json:"Online"`
}

type PeerStatus struct {
	ID           string    `json:"ID"`
	DNSName      string    `json:"DNSName"`
	HostName     string    `json:"HostName"`
	TailscaleIPs []string  `json:"TailscaleIPs"`
	OS           string    `json:"OS"`
	Online       bool      `json:"Online"`
	Active       bool      `json:"Active"`
	Relay        string    `json:"Relay"`
	CurAddr      string    `json:"CurAddr"`
	LastSeen     time.Time `json:"LastSeen"`
	Tags         []string  `json:"Tags"`
	PrimaryRoutes []string `json:"PrimaryRoutes"` // Subnets
	ExitNodeOption bool    `json:"ExitNodeOption"` // Is this peer an exit node option?
	IsSelf       bool
}

func GetStatus() (Status, error) {
	out, err := exec.Command("tailscale", "status", "--json").Output()
	var newStatus Status
	if err != nil {
		return newStatus, err
	}
	if err := json.Unmarshal(out, &newStatus); err != nil {
		return newStatus, err
	}
	
	// Ensure Self is marked
	newStatus.Self.IsSelf = true
	if newStatus.Self.HostName == "" && newStatus.Self.DNSName != "" {
		parts := strings.Split(newStatus.Self.DNSName, ".")
		if len(parts) > 0 {
			newStatus.Self.HostName = parts[0]
		}
	}
	
	for k, v := range newStatus.Peer {
		if v.HostName == "" && v.DNSName != "" {
			parts := strings.Split(v.DNSName, ".")
			if len(parts) > 0 {
				v.HostName = parts[0]
			}
		}
		newStatus.Peer[k] = v
	}

	return newStatus, nil
}

// FileCp executes tailscale file cp
func FileCp(filePath string, target string) error {
	cmd := exec.Command("tailscale", "file", "cp", filePath, target+":")
	return cmd.Run()
}
