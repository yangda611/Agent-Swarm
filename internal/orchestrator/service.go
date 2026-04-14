package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"maliangswarm/internal/domain"
	"maliangswarm/internal/eventbus"
	"maliangswarm/internal/storage"
)

const MaxDemoStep = 5

var ErrNoActiveRun = errors.New("no active run found")

type Service struct {
	bus          *eventbus.Bus
	store        *storage.Store
	artifactRoot string
	mu           sync.Mutex
}

func NewService(workdir string, bus *eventbus.Bus) (*Service, error) {
	store, err := storage.NewSQLiteStore(filepath.Join(workdir, "data", "maliang-swarm-v1.db"))
	if err != nil {
		return nil, err
	}

	s := &Service{
		bus:          bus,
		store:        store,
		artifactRoot: filepath.Join(workdir, "data", "artifacts"),
	}
	return s, nil
}

func (s *Service) Close() error {
	if s == nil || s.store == nil {
		return nil
	}
	return s.store.Close()
}

func (s *Service) GetDashboardState(ctx context.Context) (DashboardState, error) {
	snapshot, err := s.loadSnapshotOrDefault(ctx)
	if err != nil {
		return DashboardState{}, err
	}
	state := toDashboardState(snapshot)
	s.publish("dashboard.snapshot.served", map[string]any{"step": state.RunStep, "tasks": len(state.Tasks), "agents": len(state.Agents)})
	return state, nil
}

func (s *Service) CreateRun(ctx context.Context, input RunCreationInput) (DashboardState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, err := s.loadSnapshotOrDefault(ctx)
	if err != nil {
		return DashboardState{}, err
	}

	settings := current.Settings
	providers := current.AIProviders

	blueprint := buildPlanningBlueprint(ctx, input, providers)
	snapshot := compileRunFromBlueprint(blueprint, settings, providers, time.Now())
	snapshot = advanceRunSnapshot(ctx, snapshot, 1, time.Now())
	snapshot, err = s.persistRunArtifacts(snapshot)
	if err != nil {
		return DashboardState{}, err
	}
	if err := s.store.SaveSnapshot(ctx, snapshot); err != nil {
		return DashboardState{}, err
	}

	s.publish("run.created", map[string]any{
		"runID":   snapshot.Run.ID,
		"title":   snapshot.Run.Title,
		"agents":  len(snapshot.Agents),
		"planner": blueprint.PlanningSource,
	})
	return toDashboardState(snapshot), nil
}

func (s *Service) AdvanceDemoRun(ctx context.Context) (DashboardState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, err := s.loadSnapshotOrDefault(ctx)
	if err != nil {
		return DashboardState{}, err
	}
	if !hasActiveRun(snapshot) {
		return DashboardState{}, ErrNoActiveRun
	}

	next := min(snapshot.Run.Step+1, MaxDemoStep)
	updated := advanceRunSnapshot(ctx, snapshot, next, time.Now())
	updated, err = s.persistRunArtifacts(updated)
	if err != nil {
		return DashboardState{}, err
	}
	if err := s.store.SaveSnapshot(ctx, updated); err != nil {
		return DashboardState{}, err
	}

	s.publish("run.advanced", map[string]any{"step": next})
	return toDashboardState(updated), nil
}

func (s *Service) ResetDemoRun(ctx context.Context) (DashboardState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, err := s.loadSnapshotOrDefault(ctx)
	if err != nil {
		return DashboardState{}, err
	}
	if !hasActiveRun(snapshot) {
		return DashboardState{}, ErrNoActiveRun
	}

	updated := clearRunOutputs(snapshot, time.Now())
	if err := s.clearRunArtifacts(updated.Run.ID); err != nil {
		return DashboardState{}, err
	}
	updated, err = s.persistRunArtifacts(updated)
	if err != nil {
		return DashboardState{}, err
	}
	if err := s.store.SaveSnapshot(ctx, updated); err != nil {
		return DashboardState{}, err
	}

	s.publish("run.reset", map[string]any{"step": 0})
	return toDashboardState(updated), nil
}

func (s *Service) UpdateRuntimeSettings(ctx context.Context, input SettingsInput) (DashboardState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, err := s.loadSnapshotOrDefault(ctx)
	if err != nil {
		return DashboardState{}, err
	}

	settings := snapshot.Settings
	if input.ConcurrencyLimit > 0 {
		settings.ConcurrencyLimit = min(input.ConcurrencyLimit, 100)
	}
	if v := compact(input.ApprovalPolicy); v != "" {
		settings.ApprovalPolicy = v
	}
	if v := compact(input.Theme); v != "" {
		settings.Theme = v
	}
	if v := compact(input.BudgetMode); v != "" {
		settings.BudgetMode = v
	}

	snapshot.Settings = settings
	updated := snapshot
	if hasActiveRun(snapshot) {
		updated = applyProgression(snapshot, snapshot.Run.Step, time.Now())
		updated, err = s.persistRunArtifacts(updated)
		if err != nil {
			return DashboardState{}, err
		}
	}
	if err := s.store.SaveSnapshot(ctx, updated); err != nil {
		return DashboardState{}, err
	}

	s.publish("settings.updated", map[string]any{"theme": settings.Theme, "budget": settings.BudgetMode})
	return toDashboardState(updated), nil
}

func (s *Service) UpsertAIProvider(ctx context.Context, input AIProviderInput) (DashboardState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, err := s.loadSnapshotOrDefault(ctx)
	if err != nil {
		return DashboardState{}, err
	}

	snapshot.AIProviders = upsertProviderConfig(snapshot.AIProviders, input, time.Now())
	snapshot.ModelProfiles = modelProfilesForProviders(snapshot.AIProviders)
	updated := snapshot
	if hasActiveRun(snapshot) {
		updated = applyProgression(snapshot, snapshot.Run.Step, time.Now())
		updated, err = s.persistRunArtifacts(updated)
		if err != nil {
			return DashboardState{}, err
		}
	}
	if err := s.store.SaveSnapshot(ctx, updated); err != nil {
		return DashboardState{}, err
	}

	s.publish("provider.upserted", map[string]any{"providerID": input.ID, "providerName": input.Name})
	return toDashboardState(updated), nil
}

func (s *Service) DeleteAIProvider(ctx context.Context, id string) (DashboardState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, err := s.loadSnapshotOrDefault(ctx)
	if err != nil {
		return DashboardState{}, err
	}

	snapshot.AIProviders = deleteProviderConfig(snapshot.AIProviders, id)
	snapshot.ModelProfiles = modelProfilesForProviders(snapshot.AIProviders)
	updated := snapshot
	if hasActiveRun(snapshot) {
		updated = applyProgression(snapshot, snapshot.Run.Step, time.Now())
		updated, err = s.persistRunArtifacts(updated)
		if err != nil {
			return DashboardState{}, err
		}
	}
	if err := s.store.SaveSnapshot(ctx, updated); err != nil {
		return DashboardState{}, err
	}

	s.publish("provider.deleted", map[string]any{"providerID": id})
	return toDashboardState(updated), nil
}

func (s *Service) ProbeAIProvider(ctx context.Context, input AIProviderInput) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, err := s.loadSnapshotOrDefault(ctx)
	if err != nil {
		return "", err
	}

	provider := previewProviderConfig(snapshot.AIProviders, input, time.Now())
	report := probeProvider(ctx, provider)
	s.publish("provider.probed", map[string]any{"providerID": provider.ID, "providerName": provider.Name, "format": provider.Format})
	return report, nil
}

func (s *Service) loadSnapshotOrDefault(ctx context.Context) (domain.Snapshot, error) {
	snapshot, err := s.store.LoadSnapshot(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrNoSnapshot) {
			return emptySnapshot(), nil
		}
		return domain.Snapshot{}, err
	}

	snapshot.Settings = normalizeSettings(snapshot.Settings)
	snapshot.AIProviders = normalizeProviders(snapshot.AIProviders)
	if len(snapshot.ModelProfiles) == 0 {
		snapshot.ModelProfiles = modelProfilesForProviders(snapshot.AIProviders)
	}
	if len(snapshot.Zones) == 0 {
		snapshot.Zones = defaultZones()
	}
	return snapshot, nil
}

func emptySnapshot() domain.Snapshot {
	settings := defaultSettings()
	providers := normalizeProviders(nil)
	return domain.Snapshot{
		Settings:      settings,
		AIProviders:   providers,
		ModelProfiles: modelProfilesForProviders(providers),
		Zones:         defaultZones(),
	}
}

func hasActiveRun(snapshot domain.Snapshot) bool {
	return compact(snapshot.Run.ID) != "" && compact(snapshot.Run.Title) != ""
}

func (s *Service) publish(name string, payload map[string]any) {
	if s.bus == nil {
		return
	}
	s.bus.Publish(eventbus.Event{Name: name, OccurredAt: time.Now(), Payload: payload})
}

func (s *Service) persistRunArtifacts(snapshot domain.Snapshot) (domain.Snapshot, error) {
	runID := compact(snapshot.Run.ID)
	if runID == "" {
		return snapshot, nil
	}

	root := filepath.Join(s.artifactRoot, sanitizePathSegment(runID))
	if err := os.MkdirAll(root, 0o755); err != nil {
		return snapshot, err
	}

	for index := range snapshot.Tasks {
		task := &snapshot.Tasks[index]
		if compact(task.OutputSummary) == "" {
			task.ArtifactPath = ""
			continue
		}

		filename := buildArtifactFilename(*task)
		target := filepath.Join(root, filename)
		content := renderTaskArtifact(snapshot.Run, *task)
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return snapshot, err
		}
		task.ArtifactPath = target
	}

	snapshot.Agents = hydrateAgentArtifacts(snapshot.Agents, snapshot.Tasks, snapshot.Handoffs)
	return snapshot, nil
}

func (s *Service) clearRunArtifacts(runID string) error {
	normalized := compact(runID)
	if normalized == "" {
		return nil
	}
	target := filepath.Join(s.artifactRoot, sanitizePathSegment(normalized))
	return os.RemoveAll(target)
}

func buildArtifactFilename(task domain.Task) string {
	artifact := sanitizeFileName(task.ArtifactName)
	if artifact == "" {
		artifact = sanitizeFileName(task.ID)
	}
	if artifact == "" {
		artifact = "task-output.md"
	}
	if filepath.Ext(artifact) == "" {
		artifact += ".md"
	}

	order := task.SortOrder
	if order <= 0 {
		order = 999
	}

	return fmt.Sprintf("%03d-%s", order, artifact)
}

func renderTaskArtifact(run domain.Run, task domain.Task) string {
	var builder strings.Builder
	builder.WriteString("# ")
	builder.WriteString(task.Title)
	builder.WriteString("\n\n")
	builder.WriteString("## 运行任务\n")
	builder.WriteString(run.Title)
	builder.WriteString("\n\n")
	builder.WriteString("## 负责角色\n")
	builder.WriteString(task.OwnerRole)
	builder.WriteString("\n\n")
	builder.WriteString("## 产物名称\n")
	builder.WriteString(task.ArtifactName)
	builder.WriteString("\n\n")
	builder.WriteString("## 当前阶段\n")
	builder.WriteString(task.Stage)
	builder.WriteString("\n\n")
	builder.WriteString("## 结果内容\n")
	builder.WriteString(task.OutputSummary)
	builder.WriteString("\n")
	return builder.String()
}

func sanitizePathSegment(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "run"
	}

	replacer := strings.NewReplacer(
		"<", "-",
		">", "-",
		":", "-",
		"\"", "-",
		"/", "-",
		"\\", "-",
		"|", "-",
		"?", "-",
		"*", "-",
	)
	normalized = replacer.Replace(normalized)
	normalized = strings.Trim(normalized, ". ")
	if normalized == "" {
		return "run"
	}
	return normalized
}

func sanitizeFileName(value string) string {
	name := sanitizePathSegment(filepath.Base(strings.TrimSpace(value)))
	switch name {
	case "", ".", "..":
		return ""
	default:
		return name
	}
}

func toDashboardState(snapshot domain.Snapshot) DashboardState {
	completedTasks := 0
	activeAgents := 0
	completedGates := []string{}
	pendingGates := []string{}
	tasks := make([]TaskState, 0, len(snapshot.Tasks))
	agents := make([]AgentState, 0, len(snapshot.Agents))
	handoffs := make([]HandoffState, 0, len(snapshot.Handoffs))
	timeline := make([]TimelineItem, 0, len(snapshot.Timeline))
	zones := make([]SceneZone, 0, len(snapshot.Zones))
	profiles := make([]ModelProfile, 0, len(snapshot.ModelProfiles))
	providers := make([]AIProvider, 0, len(snapshot.AIProviders))

	for _, gate := range snapshot.Gates {
		if gate.Status == "completed" {
			completedGates = append(completedGates, gate.Label)
		} else {
			pendingGates = append(pendingGates, fmt.Sprintf("%s (%s)", gate.Label, gate.Status))
		}
	}

	for _, task := range snapshot.Tasks {
		if task.Status == "completed" {
			completedTasks++
		}
		tasks = append(tasks, TaskState{
			ID:             task.ID,
			Title:          task.Title,
			Stage:          task.Stage,
			OwnerRole:      task.OwnerRole,
			ArtifactName:   task.ArtifactName,
			ArtifactPath:   task.ArtifactPath,
			Status:         task.Status,
			RiskLevel:      task.RiskLevel,
			Detail:         task.Detail,
			OutputSummary:  task.OutputSummary,
			Progress:       task.Progress,
			ReviewRequired: task.ReviewRequired,
			Dependencies:   task.Dependencies,
		})
	}

	for _, agent := range snapshot.Agents {
		if !slices.Contains([]string{"idle", "done"}, agent.Status) {
			activeAgents++
		}
		agents = append(agents, AgentState{
			ID:               agent.ID,
			Name:             agent.Name,
			Role:             agent.Role,
			Team:             agent.Team,
			DeskLabel:        agent.DeskLabel,
			ModelTier:        agent.ModelTier,
			Status:           agent.Status,
			ComputerState:    agent.ComputerState,
			CurrentTask:      agent.CurrentTask,
			Detail:           agent.Detail,
			Artifact:         agent.Artifact,
			NextTarget:       agent.NextTarget,
			LastAction:       agent.LastAction,
			LastOutput:       agent.LastOutput,
			LastHandoff:      agent.LastHandoff,
			LastReceiver:     agent.LastReceiver,
			LastArtifactPath: agent.LastArtifactPath,
			LastUpdate:       agent.LastUpdate,
			RiskLevel:        agent.RiskLevel,
			Progress:         agent.Progress,
			QueueDepth:       agent.QueueDepth,
			X:                agent.X,
			Y:                agent.Y,
			Highlights:       agent.Highlights,
		})
	}

	for _, handoff := range snapshot.Handoffs {
		handoffs = append(handoffs, HandoffState{
			ID:           handoff.ID,
			ArtifactName: handoff.ArtifactName,
			FromAgentID:  handoff.FromAgentID,
			ToAgentID:    handoff.ToAgentID,
			Status:       handoff.Status,
		})
	}

	for _, item := range snapshot.Timeline {
		timeline = append(timeline, TimelineItem{
			TimeLabel: item.TimeLabel,
			Kind:      item.Kind,
			Title:     item.Title,
			Detail:    item.Detail,
		})
	}

	for _, zone := range snapshot.Zones {
		zones = append(zones, SceneZone{
			ID:      zone.ID,
			Name:    zone.Name,
			Purpose: zone.Purpose,
			Accent:  zone.Accent,
			X:       zone.X,
			Y:       zone.Y,
			W:       zone.W,
			H:       zone.H,
		})
	}

	for _, profile := range snapshot.ModelProfiles {
		profiles = append(profiles, ModelProfile{
			Tier:           profile.Tier,
			Binding:        profile.Binding,
			Responsibility: profile.Responsibility,
		})
	}
	for _, provider := range snapshot.AIProviders {
		providers = append(providers, providerState(provider))
	}

	hasRun := hasActiveRun(snapshot)
	title := snapshot.Run.Title
	subtitle := snapshot.Run.Subtitle
	workspaceName := snapshot.Run.WorkspaceName
	activeTemplate := snapshot.Run.ActiveTemplate
	suggestedAgentID := snapshot.Run.SuggestedAgentID
	workflow := WorkflowSummary{
		CurrentStage:     snapshot.Run.CurrentStage,
		AtomicTaskPolicy: snapshot.Run.AtomicTaskPolicy,
		ReviewMode:       snapshot.Run.ReviewMode,
		PlannerSource:    snapshot.Run.PlannerSource,
		PendingGates:     pendingGates,
		CompletedGates:   completedGates,
	}

	if !hasRun {
		title = "Empty project"
		subtitle = "Create the first task to start planning, execution, and review."
		workspaceName = productName + " / Empty workspace"
		activeTemplate = ""
		suggestedAgentID = ""
		workflow = WorkflowSummary{
			CurrentStage:     "No active run",
			AtomicTaskPolicy: "Create a task to let the planner decompose work into atomic units.",
			ReviewMode:       "This workspace has no active orchestration run yet.",
			PlannerSource:    "not configured",
			PendingGates:     nil,
			CompletedGates:   nil,
		}
	}

	return DashboardState{
		HasActiveRun:     hasRun,
		Title:            title,
		Subtitle:         subtitle,
		WorkspaceName:    workspaceName,
		ActiveTemplate:   activeTemplate,
		SuggestedAgentID: suggestedAgentID,
		RunStep:          snapshot.Run.Step,
		MaxStep:          MaxDemoStep,
		Metrics: []MetricCard{
			{Label: "活跃 Agent", Value: fmt.Sprintf("%d / %d", activeAgents, snapshot.Settings.ConcurrencyLimit), Accent: "cyan", Detail: "当前工作中的智能体数量。"},
			{Label: "原子任务", Value: fmt.Sprintf("%d / %d 已完成", completedTasks, len(snapshot.Tasks)), Accent: "gold", Detail: "每个任务都保持单一负责人和单一产物。"},
			{Label: "审核门禁", Value: fmt.Sprintf("%d / %d 已关闭", len(completedGates), len(snapshot.Gates)), Accent: "rose", Detail: "同评、QA、安全与交付门禁闭环情况。"},
			{Label: "AI 接口注册表", Value: fmt.Sprintf("%d 在线 / %d 总计", enabledProviderCount(snapshot.AIProviders), len(snapshot.AIProviders)), Accent: "green", Detail: primaryProviderLabel(snapshot.AIProviders)},
		},
		Workflow: workflow,
		Settings: SettingsSummary{
			ConcurrencyLimit: snapshot.Settings.ConcurrencyLimit,
			ApprovalPolicy:   snapshot.Settings.ApprovalPolicy,
			Theme:            snapshot.Settings.Theme,
			BudgetMode:       snapshot.Settings.BudgetMode,
			PrimaryProvider:  primaryProviderLabel(snapshot.AIProviders),
			AIProviders:      providers,
			ModelProfiles:    profiles,
		},
		Zones:    zones,
		Tasks:    tasks,
		Agents:   agents,
		Handoffs: handoffs,
		Timeline: timeline,
	}
}
