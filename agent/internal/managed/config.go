package managed

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	BackendURL, InstallationID, Token, TokenFile string
	NodeName, NodeUID, Namespace, PodName        string
	PodUID, PodIP, StateDir, Version             string
	ListenAddress, MetricsURL                    string
	PollInterval                                 time.Duration
	AllowInsecureHTTP                            bool
}

func ConfigFromEnv() (Config, error) {
	dir := os.Getenv("STATE_DIR")
	if dir == "" {
		dir = "/var/lib/ctrlb"
		if err := os.Setenv("STATE_DIR", dir); err != nil {
			return Config{}, err
		}
	}
	cfg := Config{
		BackendURL: os.Getenv("BACKEND_URL"), InstallationID: os.Getenv("INSTALLATION_ID"),
		Token: os.Getenv("INSTALLATION_TOKEN"), TokenFile: os.Getenv("INSTALLATION_TOKEN_FILE"),
		NodeName: os.Getenv("K8S_NODE_NAME"), NodeUID: os.Getenv("K8S_NODE_UID"),
		Namespace: os.Getenv("POD_NAMESPACE"), PodName: os.Getenv("POD_NAME"),
		PodUID: os.Getenv("POD_UID"), PodIP: os.Getenv("POD_IP"), StateDir: dir,
		ListenAddress: ":3421", MetricsURL: "http://127.0.0.1:8888/metrics", PollInterval: 10 * time.Second,
		AllowInsecureHTTP: os.Getenv("ALLOW_INSECURE_HTTP") == "true",
	}
	return cfg, cfg.validate()
}

func (c Config) validate() error {
	u, err := url.Parse(c.BackendURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return errors.New("BACKEND_URL must be an absolute control-plane URL without credentials, query or fragment")
	}
	loopback := u.Hostname() == "localhost"
	if ip := net.ParseIP(u.Hostname()); ip != nil {
		loopback = ip.IsLoopback()
	}
	if u.Scheme != "https" && !(u.Scheme == "http" && (loopback || c.AllowInsecureHTTP)) {
		return errors.New("BACKEND_URL must use HTTPS; set ALLOW_INSECURE_HTTP=true explicitly for an HTTP demo")
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(c.InstallationID) {
		return errors.New("INSTALLATION_ID must contain only letters, numbers, hyphens and underscores")
	}
	if strings.TrimSpace(c.NodeName) == "" {
		return errors.New("K8S_NODE_NAME is required in managed mode")
	}
	if c.StateDir == "" {
		return errors.New("STATE_DIR is required")
	}
	_, err = c.token()
	return err
}

func (c Config) token() (string, error) {
	token := c.Token
	if c.TokenFile != "" {
		data, err := os.ReadFile(c.TokenFile)
		if err != nil {
			return "", fmt.Errorf("read installation token file: %w", err)
		}
		token = string(data)
	}
	token = strings.TrimSpace(token)
	if token == "" || strings.ContainsAny(token, "\r\n") {
		return "", errors.New("INSTALLATION_TOKEN or INSTALLATION_TOKEN_FILE must contain a nonempty bearer token")
	}
	return token, nil
}
