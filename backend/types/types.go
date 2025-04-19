package types

// Deployment tracks basic deployment metadata
type Deployment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Language  string `json:"language"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
	Port      string `json:"port,omitempty"` // Store the port if running
}

// DeploymentDetail includes the deployment metadata plus code and package content
type DeploymentDetail struct {
	Deployment
	Code    string `json:"code,omitempty"`
	Package string `json:"package,omitempty"`
}
