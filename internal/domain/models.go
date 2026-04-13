package domain

// Snapshot is the persisted runtime projection for a single run.
type Snapshot struct {
	Run           Run
	Settings      RuntimeSettings
	AIProviders   []AIProviderConfig
	ModelProfiles []ModelProfile
	Zones         []SceneZone
	Gates         []WorkflowGate
	Tasks         []Task
	Agents        []Agent
	Handoffs      []Handoff
	Timeline      []TimelineEvent
}

// RuntimeSettings captures the high-impact knobs exposed in the control desk.
type RuntimeSettings struct {
	ConcurrencyLimit int
	ApprovalPolicy   string
	Theme            string
	BudgetMode       string
}

// AIProviderConfig captures one user-configurable AI API connection profile.
type AIProviderConfig struct {
	ID            string
	Name          string
	Format        string
	BaseURL       string
	APIPath       string
	APIVersion    string
	APIKey        string
	DefaultModel  string
	PlannerModel  string
	WorkerModel   string
	ReviewerModel string
	HeadersJSON   string
	Notes         string
	Enabled       bool
	IsPrimary     bool
	SortOrder     int
}

// ModelProfile maps a tier to its current model binding.
type ModelProfile struct {
	SortOrder      int
	Tier           string
	Binding        string
	Responsibility string
}

// Run is the top-level workflow instance.
type Run struct {
	ID               string
	Title            string
	Mission          string
	Deliverable      string
	Priority         string
	TaskType         string
	Subtitle         string
	WorkspaceName    string
	ActiveTemplate   string
	CurrentStage     string
	AtomicTaskPolicy string
	ReviewMode       string
	PlannerSource    string
	SuggestedAgentID string
	Step             int
}

// WorkflowGate tracks a mandatory enterprise gate.
type WorkflowGate struct {
	ID        string
	Label     string
	Status    string
	SortOrder int
}

// SceneZone defines a visible office area.
type SceneZone struct {
	ID        string
	Name      string
	Purpose   string
	Accent    string
	X         int
	Y         int
	W         int
	H         int
	SortOrder int
}

// Task captures one atomic unit of work.
type Task struct {
	ID             string
	Title          string
	Stage          string
	OwnerAgentID   string
	OwnerRole      string
	ArtifactName   string
	ArtifactPath   string
	Status         string
	RiskLevel      string
	Detail         string
	OutputSummary  string
	Progress       int
	ReviewRequired bool
	Dependencies   int
	SortOrder      int
}

// Agent is a visible worker in the office.
type Agent struct {
	ID               string
	Name             string
	Role             string
	Team             string
	DeskLabel        string
	ModelTier        string
	Status           string
	ComputerState    string
	CurrentTask      string
	Detail           string
	Artifact         string
	NextTarget       string
	LastAction       string
	LastOutput       string
	LastHandoff      string
	LastReceiver     string
	LastArtifactPath string
	LastUpdate       string
	RiskLevel        string
	Progress         int
	QueueDepth       int
	X                int
	Y                int
	Highlights       []string
	SortOrder        int
}

// Handoff is a visible artifact transfer between two agents.
type Handoff struct {
	ID           string
	ArtifactName string
	FromAgentID  string
	ToAgentID    string
	Status       string
	SortOrder    int
}

// TimelineEvent is a replayable high-signal event.
type TimelineEvent struct {
	TimeLabel string
	Kind      string
	Title     string
	Detail    string
	SortOrder int
}
