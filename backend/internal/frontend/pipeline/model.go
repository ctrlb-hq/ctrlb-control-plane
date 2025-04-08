package frontendpipeline

type Pipeline struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Agents        int    `json:"agents"`
	IncomingBytes int    `json:"incoming_bytes"`
	OutgoingBytes int    `json:"outgoing_bytes"`
	UpdatedAt     int    `json:"updatedAt"`
}

type PipelineInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"created_by"`
	CreatedAt int    `json:"created_at"`
	UpdatedAt int    `json:"updated_at"`
}
