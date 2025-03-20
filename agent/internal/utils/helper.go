package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/ctrlb-hq/ctrlb-collector/agent/internal/pkg/logger"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}

func UnmarshalJSONRequest(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	defer r.Body.Close()
	return nil
}

func UnmarshalJSONResponse(r *http.Response, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	defer r.Body.Close()
	return nil
}

func SendJSONError(w http.ResponseWriter, statusCode int, errMsg string) {
	errResp := ErrorResponse{
		Error: errMsg,
	}

	// Set the response content type
	w.Header().Set("Content-Type", "application/json")

	// Set the response status code
	w.WriteHeader(statusCode)

	// Marshal the error response struct to JSON
	jsonData, err := json.Marshal(errResp)
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("Failed to marshal error response to JSON: %s", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content", errMsg)

	// Write the JSON response to the response writer
	w.Write(jsonData)
}

func GetLocalIP() (string, error) {
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
