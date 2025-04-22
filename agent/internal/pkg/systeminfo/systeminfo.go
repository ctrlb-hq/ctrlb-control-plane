package systeminfo

import (
	"fmt"
	"net"
	"os"
)

type SystemInfoProvider interface {
	GetHostname() (string, error)
	LookupHost(string) ([]string, error)
	GetLocalIP() (string, error)
}

type DefaultSystemInfo struct{}

func (d *DefaultSystemInfo) GetHostname() (string, error) {
	return os.Hostname()
}

func (d *DefaultSystemInfo) LookupHost(hostname string) ([]string, error) {
	return net.LookupHost(hostname)
}

func (d *DefaultSystemInfo) GetLocalIP() (string, error) {
	// Get all interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %v", err)
	}

	for _, i := range interfaces {
		// Ignore interfaces that are down or loopback
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			return "", fmt.Errorf("failed to get addresses: %v", err)
		}

		// Return the first non-loopback IP address found
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Only consider IPv4 addresses (you can modify this to handle IPv6 if needed)
			if ip != nil && ip.To4() != nil {
				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no valid IP address found")
}

func NewSystemInfo() SystemInfoProvider {
	return &DefaultSystemInfo{}
}
