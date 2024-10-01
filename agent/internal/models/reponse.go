package models

type HealthResponse struct {
	Status       string
	AgentVersion string
	UpTime       int
}
