package myhttpclient

// Types and models used by the ALS client.

// RetrievalRequest describes the query body sent to ALS retrieval API.
type RetrievalRequest struct {
	TimeFrom  string `json:"timeFrom"`
	TimeTo    string `json:"timeTo"`
	Source    string `json:"source"`
	EventType string `json:"eventType"`
}

// QueryResponse represents the response returned when creating a retrieval query.
// Example: { "id": "<query-id>" }
type QueryResponse struct {
	ID string `json:"id"`
}

// StatusResponse represents a status check response.
// Example: { "status": "finished" }
type StatusResponse struct {
	Status string `json:"status"`
}

// ALSResponseItem mirrors the JSON structure returned by the ALS service.
// Only fields needed by callers are included; extend as necessary.
type ALSResponseItem struct {
	ID                string  `json:"id"`
	SpecVersion       string  `json:"specversion"`
	Source            string  `json:"source"`
	Type              string  `json:"type"`
	DataSchema        string  `json:"dataschema"`
	Time              string  `json:"time"`
	XsapIngestionTime string  `json:"xsapingestiontime"`
	Data              ALSData `json:"data"`
}

type ALSData struct {
	Data     ALSInnerData `json:"data"`
	Metadata ALSMetadata  `json:"metadata"`
}

type ALSInnerData struct {
	CMKDelete CMKDelete `json:"cmkDelete"`
}

type CMKDelete struct {
	CMKID       string `json:"cmkId"`
	KmsSystemID string `json:"kmsSystemId"`
}

type ALSMetadata struct {
	Ts              string            `json:"ts"`
	UserInitiatorID string            `json:"userInitiatorId"`
	TenantID        string            `json:"tenantId"`
	AppID           string            `json:"appId"`
	AppContext      ALSAppContext     `json:"appContext"`
	Infrastructure  ALSInfrastructure `json:"infrastructure"`
	Platform        ALSPlatform       `json:"platform"`
}

type ALSAppContext struct {
	SapKmsTenantContext string `json:"sap-kms-tenant-context"`
	EventCorrelationID  string `json:"event-correlation-id"`
}

type ALSInfrastructure struct {
	App ALSApp `json:"app"`
	K8s ALSK8s `json:"k8s"`
}

type ALSApp struct {
	Image   string `json:"image"`
	Version string `json:"version"`
}

type ALSK8s struct {
	InfrastructureRegion string `json:"infrastructureRegion"`
	Cluster              string `json:"cluster"`
}

type ALSPlatform struct {
	UnifiedServices ALSUnifiedServices `json:"unifiedServices"`
}

type ALSUnifiedServices struct {
	FolderPath        string `json:"folderPath"`
	AccountID         string `json:"accountId"`
	ResourcegroupPath string `json:"resourcegroupPath"`
}
