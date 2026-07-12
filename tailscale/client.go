package tailscale

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Status struct {
	Peer           map[string]PeerStatus `json:"Peer"`
	Self           PeerStatus            `json:"Self"`
	ExitNodeStatus *ExitNodeStatus       `json:"ExitNodeStatus,omitempty"`
}

type ExitNodeStatus struct {
	ID     string `json:"ID"`
	Online bool   `json:"Online"`
}

type PeerStatus struct {
	ID                  string    `json:"ID"`
	DNSName             string    `json:"DNSName"`
	HostName            string    `json:"HostName"`
	TailscaleIPs        []string  `json:"TailscaleIPs"`
	OS                  string    `json:"OS"`
	Online              bool      `json:"Online"`
	Active              bool      `json:"Active"`
	Relay               string    `json:"Relay"`
	CurAddr             string    `json:"CurAddr"`
	LastSeen            time.Time `json:"LastSeen"`
	Tags                []string  `json:"Tags"`
	PrimaryRoutes       []string  `json:"PrimaryRoutes"`  // Subnets
	ExitNodeOption      bool      `json:"ExitNodeOption"` // Is this peer an exit node option?
	ExitNode            bool      `json:"ExitNode"`
	TailscaleSSHEnabled bool      `json:"TailscaleSSHEnabled"`
	KeyExpiry           time.Time `json:"KeyExpiry"`
	TxBytes             int64     `json:"TxBytes"`
	RxBytes             int64     `json:"RxBytes"`
	IsSelf              bool
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

// FileCp sends a file to a peer via Taildrop.
func FileCp(filePath string, target string) error {
	cmd := exec.Command("tailscale", "file", "cp", filePath, target+":")
	return cmd.Run()
}

// FileGet receives pending Taildrop files into the current directory.
func FileGet() error {
	cmd := exec.Command("tailscale", "file", "get", ".")
	return cmd.Run()
}

// ServeStatus returns the output of `tailscale serve status`.
func ServeStatus() (string, error) {
	out, err := exec.Command("tailscale", "serve", "status").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// GetSSHPort resolves the SSH port for a host via the local ssh client config
// (falls back to 22 if it can't be determined).
func GetSSHPort(host string) string {
	out, err := exec.Command("ssh", "-G", host).Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(strings.ToLower(line), "port ") {
				return strings.TrimSpace(strings.SplitN(line, " ", 2)[1])
			}
		}
	}
	return "22"
}

// DaemonCmd runs `tailscale <args...>`, used for up/down/shields/accept-routes/exit-node.
func DaemonCmd(args ...string) error {
	cmd := exec.Command("tailscale", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tailscale %s: %w", strings.Join(args, " "), err)
	}
	return nil
}
