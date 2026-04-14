package orchestrator

// DashboardState is the single snapshot returned to the desktop UI.
type DashboardState struct {
	HasActiveRun     bool            `json:"hasActiveRun"`
	Title            string          `json:"title"`
	Subtitle         string          `json:"subtitle"`
	WorkspaceName    string          `json:"workspaceName"`
	ActiveTemplate   string          `json:"activeTemplate"`
	SuggestedAgentID string          `json:"suggestedAgentId"`
	RunStep          int             `json:"runStep"`
	MaxStep          int             `json:"maxStep"`
	Metrics          []MetricCard    `json:"metrics"`
	Workflow         WorkflowSummary `json:"workflow"`
	Settings         SettingsSummary `json:"settings"`
	Zones            []SceneZone     `json:"zones"`
	Tasks            []TaskState     `json:"tasks"`
	Agents           []AgentState    `json:"agents"`
	Handoffs         []HandoffState  `json:"handoffs"`
	Timeline         []TimelineItem  `json:"timeline"`
}

// MetricCard highlights a top-level operational KPI.
type MetricCard struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	Accent string `json:"accent"`
	Detail string `json:"detail"`
}

// WorkflowSummary explains which enterprise path is active.
type WorkflowSummary struct {
	CurrentStage     string   `json:"currentStage"`
	AtomicTaskPolicy string   `json:"atomicTaskPolicy"`
	ReviewMode       string   `json:"reviewMode"`
	PlannerSource    string   `json:"plannerSource"`
	PendingGates     []string `json:"pendingGates"`
	CompletedGates   []string `json:"completedGates"`
}

// SettingsSummary exposes the high-impact controls.
type SettingsSummary struct {
	ConcurrencyLimit int            `json:"concurrencyLimit"`
	ApprovalPolicy   string         `json:"approvalPolicy"`
	Theme            string         `json:"theme"`
	BudgetMode       string         `json:"budgetMode"`
	PrimaryProvider  string         `json:"primaryProvider"`
	AIProviders      []AIProvider   `json:"aiProviders"`
	ModelProfiles    []ModelProfile `json:"modelProfiles"`
}

// ModelProfile maps a model tier to a role family.
type ModelProfile struct {
	Tier           string `json:"tier"`
	Binding        string `json:"binding"`
	Responsibility string `json:"responsibility"`
}

// AIProvider describes one configured AI API route in the control center.
type AIProvider struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Format           string `json:"format"`
	BaseURL          string `json:"baseUrl"`
	APIPath          string `json:"apiPath"`
	APIVersion       string `json:"apiVersion"`
	DefaultModel     string `json:"defaultModel"`
	PlannerModel     string `json:"plannerModel"`
	WorkerModel      string `json:"workerModel"`
	ReviewerModel    string `json:"reviewerModel"`
	HeadersJSON      string `json:"headersJson"`
	Notes            string `json:"notes"`
	Enabled          bool   `json:"enabled"`
	IsPrimary        bool   `json:"isPrimary"`
	APIKeyConfigured bool   `json:"apiKeyConfigured"`
	APIKeyPreview    string `json:"apiKeyPreview"`
}

// SceneZone defines a spatial office area on the pixel map.
type SceneZone struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Purpose string `json:"purpose"`
	Accent  string `json:"accent"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	W       int    `json:"w"`
	H       int    `json:"h"`
}

// TaskState summarises one atomic task on the board.
type TaskState struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Stage          string `json:"stage"`
	OwnerRole      string `json:"ownerRole"`
	ArtifactName   string `json:"artifactName"`
	ArtifactPath   string `json:"artifactPath"`
	Status         string `json:"status"`
	RiskLevel      string `json:"riskLevel"`
	Detail         string `json:"detail"`
	OutputSummary  string `json:"outputSummary"`
	Progress       int    `json:"progress"`
	ReviewRequired bool   `json:"reviewRequired"`
	Dependencies   int    `json:"dependencies"`
}

// AgentState describes one visible office worker.
type AgentState struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Role             string   `json:"role"`
	Team             string   `json:"team"`
	DeskLabel        string   `json:"deskLabel"`
	ModelTier        string   `json:"modelTier"`
	Status           string   `json:"status"`
	ComputerState    string   `json:"computerState"`
	CurrentTask      string   `json:"currentTask"`
	Detail           string   `json:"detail"`
	Artifact         string   `json:"artifact"`
	NextTarget       string   `json:"nextTarget"`
	LastAction       string   `json:"lastAction"`
	LastOutput       string   `json:"lastOutput"`
	LastHandoff      string   `json:"lastHandoff"`
	LastReceiver     string   `json:"lastReceiver"`
	LastArtifactPath string   `json:"lastArtifactPath"`
	LastUpdate       string   `json:"lastUpdate"`
	RiskLevel        string   `json:"riskLevel"`
	Progress         int      `json:"progress"`
	QueueDepth       int      `json:"queueDepth"`
	X                int      `json:"x"`
	Y                int      `json:"y"`
	Highlights       []string `json:"highlights"`
}

// HandoffState visualises an artifact moving between two agents.
type HandoffState struct {
	ID           string `json:"id"`
	ArtifactName string `json:"artifactName"`
	FromAgentID  string `json:"fromAgentId"`
	ToAgentID    string `json:"toAgentId"`
	Status       string `json:"status"`
}

// TimelineItem records the most recent high-signal events.
type TimelineItem struct {
	TimeLabel string `json:"timeLabel"`
	Kind      string `json:"kind"`
	Title     string `json:"title"`
	Detail    string `json:"detail"`
}

// SettingsInput is the editable subset of runtime settings exposed to the UI.
type SettingsInput struct {
	ConcurrencyLimit int    `json:"concurrencyLimit"`
	ApprovalPolicy   string `json:"approvalPolicy"`
	Theme            string `json:"theme"`
	BudgetMode       string `json:"budgetMode"`
}

// AIProviderInput is the editable provider payload submitted from the UI.
type AIProviderInput struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Format        string `json:"format"`
	BaseURL       string `json:"baseUrl"`
	APIPath       string `json:"apiPath"`
	APIVersion    string `json:"apiVersion"`
	APIKey        string `json:"apiKey"`
	DefaultModel  string `json:"defaultModel"`
	PlannerModel  string `json:"plannerModel"`
	WorkerModel   string `json:"workerModel"`
	ReviewerModel string `json:"reviewerModel"`
	HeadersJSON   string `json:"headersJson"`
	Notes         string `json:"notes"`
	Enabled       bool   `json:"enabled"`
	IsPrimary     bool   `json:"isPrimary"`
}

// RunCreationInput is the user-submitted task brief that seeds a new orchestration run.
type RunCreationInput struct {
	Title                 string `json:"title"`
	Mission               string `json:"mission"`
	Deliverable           string `json:"deliverable"`
	TaskType              string `json:"taskType"`
	Priority              string `json:"priority"`
	MaxAgents             int    `json:"maxAgents"`
	RequiresQA            bool   `json:"requiresQA"`
	RequiresSecurity      bool   `json:"requiresSecurity"`
	RequiresHumanApproval bool   `json:"requiresHumanApproval"`
}
