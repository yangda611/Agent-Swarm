package orchestrator

import (
	"fmt"
	"strings"
	"time"

	"maliangswarm/internal/domain"
)

func buildDemoSnapshot(settings domain.RuntimeSettings, step int, now time.Time) domain.Snapshot {
	step = max(0, min(step, MaxDemoStep))
	settings = normalizeSettings(settings)

	return domain.Snapshot{
		Run: domain.Run{
			ID:               "demo-enterprise-run",
			Title:            productName,
			Subtitle:         "A persisted enterprise workflow where planning, execution, review, approvals, and AI routing all share the same office floor.",
			WorkspaceName:    productName + " / Enterprise delivery command room",
			ActiveTemplate:   "Enterprise Task Delivery v1",
			CurrentStage:     pickString(step, []string{"Intake and workflow selection", "Atomic decomposition and staffing", "Parallel implementation", "Peer review", "Security approval hold", "Delivery packaging"}),
			AtomicTaskPolicy: "Decompose until every task has one owner, one artifact, and an isolated retry path.",
			ReviewMode:       "Four-eyes review by default, human approval for high-risk actions, and persistent audit at every stage.",
			SuggestedAgentID: pickString(step, []string{"chief", "decomposer", "builder-a", "reviewer", "security", "delivery"}),
			Step:             step,
		},
		Settings:      settings,
		AIProviders:   defaultAIProviders(),
		ModelProfiles: defaultModelProfiles(),
		Zones:         defaultZones(),
		Gates:         gatesForStep(step),
		Tasks:         tasksForStep(step),
		Agents:        agentsForStep(step, now),
		Handoffs:      handoffsForStep(step),
		Timeline:      timelineForStep(step, now),
	}
}

func defaultSettings() domain.RuntimeSettings {
	return domain.RuntimeSettings{
		ConcurrencyLimit: 100,
		ApprovalPolicy:   "高风险工具调用和预算超限动作必须经过人工审批。",
		Theme:            "pixel-hq-amber",
		BudgetMode:       "guardrailed",
	}
}

func normalizeSettings(settings domain.RuntimeSettings) domain.RuntimeSettings {
	if settings.ConcurrencyLimit <= 0 {
		settings.ConcurrencyLimit = 100
	}
	settings.ConcurrencyLimit = min(settings.ConcurrencyLimit, 100)
	if compact(settings.ApprovalPolicy) == "" {
		settings.ApprovalPolicy = defaultSettings().ApprovalPolicy
	}
	if compact(settings.Theme) == "" {
		settings.Theme = defaultSettings().Theme
	}
	if compact(settings.BudgetMode) == "" {
		settings.BudgetMode = defaultSettings().BudgetMode
	}
	return settings
}

func defaultModelProfiles() []domain.ModelProfile {
	return modelProfilesForProviders(defaultAIProviders())
}

func defaultZones() []domain.SceneZone {
	return []domain.SceneZone{
		{SortOrder: 1, ID: "command", Name: "指挥台", Purpose: "需求接入、主脑编排和流程选择", Accent: "gold", X: 4, Y: 6, W: 23, H: 20},
		{SortOrder: 2, ID: "planning", Name: "规划室", Purpose: "任务拆解与角色排布", Accent: "cyan", X: 30, Y: 6, W: 32, H: 20},
		{SortOrder: 3, ID: "engineering", Name: "执行区", Purpose: "并行开发、实现与集成", Accent: "blue", X: 4, Y: 33, W: 54, H: 46},
		{SortOrder: 4, ID: "review", Name: "审核走廊", Purpose: "同行评审、QA 与安全复核", Accent: "rose", X: 62, Y: 33, W: 34, H: 24},
		{SortOrder: 5, ID: "delivery", Name: "交付台", Purpose: "交付打包、审批留痕与归档", Accent: "green", X: 62, Y: 61, W: 34, H: 18},
	}
}

func gatesForStep(step int) []domain.WorkflowGate {
	return []domain.WorkflowGate{
		{SortOrder: 1, ID: "gate-intake", Label: "Intake normalization", Status: pickString(step, []string{"in_progress", "completed", "completed", "completed", "completed", "completed"})},
		{SortOrder: 2, ID: "gate-plan", Label: "Planning review", Status: pickString(step, []string{"queued", "in_progress", "completed", "completed", "completed", "completed"})},
		{SortOrder: 3, ID: "gate-staff", Label: "DAG staffing check", Status: pickString(step, []string{"queued", "queued", "completed", "completed", "completed", "completed"})},
		{SortOrder: 4, ID: "gate-peer", Label: "Peer review on execution bundle", Status: pickString(step, []string{"queued", "queued", "queued", "in_progress", "completed", "completed"})},
		{SortOrder: 5, ID: "gate-security", Label: "Security approval on outbound action", Status: pickString(step, []string{"queued", "queued", "queued", "queued", "blocked", "completed"})},
		{SortOrder: 6, ID: "gate-delivery", Label: "Final delivery sign-off", Status: pickString(step, []string{"queued", "queued", "queued", "queued", "queued", "in_progress"})},
	}
}

func tasksForStep(step int) []domain.Task {
	return []domain.Task{
		{SortOrder: 1, ID: "task-intake", Title: "Normalize the user request", Stage: "intake", OwnerAgentID: "intake", OwnerRole: "Intake Analyst", Status: pickString(step, []string{"in_progress", "completed", "completed", "completed", "completed", "completed"}), RiskLevel: "low", Detail: "Capture concurrency ceiling, audit requirements, and approval coverage.", Progress: pickInt(step, []int{45, 100, 100, 100, 100, 100}), ReviewRequired: false, Dependencies: 0},
		{SortOrder: 2, ID: "task-template", Title: "Select the enterprise workflow template", Stage: "planning", OwnerAgentID: "chief", OwnerRole: "Chief Orchestrator", Status: pickString(step, []string{"in_progress", "completed", "completed", "completed", "completed", "completed"}), RiskLevel: "medium", Detail: "Choose the standard enterprise flow before execution begins.", Progress: pickInt(step, []int{38, 100, 100, 100, 100, 100}), ReviewRequired: true, Dependencies: 1},
		{SortOrder: 3, ID: "task-dag", Title: "Decompose work into atomic tasks", Stage: "planning", OwnerAgentID: "decomposer", OwnerRole: "Task Decomposer", Status: pickString(step, []string{"queued", "in_progress", "completed", "completed", "completed", "completed"}), RiskLevel: "medium", Detail: "Split work until every node has one owner and one retry boundary.", Progress: pickInt(step, []int{0, 72, 100, 100, 100, 100}), ReviewRequired: true, Dependencies: 2},
		{SortOrder: 4, ID: "task-ui", Title: "Build the office UI shell", Stage: "execution", OwnerAgentID: "builder-a", OwnerRole: "Implementation Agent", Status: pickString(step, []string{"queued", "queued", "in_progress", "completed", "completed", "completed"}), RiskLevel: "low", Detail: "Lay out the office floor, task board, inspector, and controls.", Progress: pickInt(step, []int{0, 0, 76, 100, 100, 100}), ReviewRequired: true, Dependencies: 1},
		{SortOrder: 5, ID: "task-runtime", Title: "Persist runtime state in SQLite", Stage: "execution", OwnerAgentID: "builder-b", OwnerRole: "Implementation Agent", Status: pickString(step, []string{"queued", "queued", "in_progress", "completed", "completed", "completed"}), RiskLevel: "medium", Detail: "Store run, task, gate, and worker state locally.", Progress: pickInt(step, []int{0, 0, 83, 100, 100, 100}), ReviewRequired: true, Dependencies: 2},
		{SortOrder: 6, ID: "task-review", Title: "Review, secure, and package delivery", Stage: "review", OwnerAgentID: "reviewer", OwnerRole: "Peer Reviewer", Status: pickString(step, []string{"queued", "queued", "queued", "in_progress", "blocked", "completed"}), RiskLevel: "high", Detail: "Peer review opens first, then security hold, then delivery packaging.", Progress: pickInt(step, []int{0, 0, 0, 58, 71, 100}), ReviewRequired: true, Dependencies: 3},
	}
}

func agentsForStep(step int, now time.Time) []domain.Agent {
	return []domain.Agent{
		{SortOrder: 1, ID: "chief", Name: "Atlas", Role: "Chief Orchestrator", Team: "Command", DeskLabel: "C-01", ModelTier: "tier-strategic", Status: pickString(step, []string{"thinking", "thinking", "thinking", "thinking", "thinking", "reviewing"}), ComputerState: pickString(step, []string{"workflow-selection", "enterprise-guardrails", "scheduler-overview", "review-pressure", "approval-watch", "release-readiness"}), CurrentTask: pickString(step, []string{"Select the enterprise workflow template", "Reserve review capacity", "Watch worker pools and gate load", "Watch peer review progress", "Hold the run on security approval", "Prepare final delivery sign-off"}), Detail: "Controls the enterprise path and prevents execution from starving review lanes.", Artifact: "workflow-blueprint.json", NextTarget: pickString(step, []string{"Mina / Intake Analyst", "Oren / Task Decomposer", "Nova / Builder A", "Rae / Peer Reviewer", "Moss / Security Reviewer", "June / Delivery Manager"}), LastUpdate: stamp(now, 32*time.Second), RiskLevel: "medium", Progress: pickInt(step, []int{38, 69, 81, 88, 93, 100}), QueueDepth: pickInt(step, []int{3, 2, 2, 2, 1, 1}), X: 12, Y: 15, Highlights: []string{"Maintains review reserve lanes", "Publishes stage transitions", "Treats approvals as hard gates"}},
		{SortOrder: 2, ID: "intake", Name: "Mina", Role: "Intake Analyst", Team: "Command", DeskLabel: "C-02", ModelTier: "tier-routing", Status: pickString(step, []string{"typing", "handoff", "done", "done", "done", "done"}), ComputerState: pickString(step, []string{"requirements-normalization", "brief-packaging", "brief-archived", "brief-archived", "brief-archived", "brief-archived"}), CurrentTask: pickString(step, []string{"Normalize goals, constraints, and delivery signals", "Package the normalized brief for planning", "Intake work complete", "Intake work complete", "Intake work complete", "Intake work complete"}), Detail: "Turns freeform intent into a structured brief the planning lane can decompose.", Artifact: "scope-brief-v4", NextTarget: "Oren / Task Decomposer", LastUpdate: stamp(now, 18*time.Second), RiskLevel: "low", Progress: pickInt(step, []int{55, 100, 100, 100, 100, 100}), QueueDepth: pickInt(step, []int{2, 1, 0, 0, 0, 0}), X: 20, Y: 15, Highlights: []string{"Captured the 100-agent ceiling", "Flagged audit coverage", "Fed a structured scope brief into planning"}},
		{SortOrder: 3, ID: "decomposer", Name: "Oren", Role: "Task Decomposer", Team: "Planning", DeskLabel: "P-02", ModelTier: "tier-strategic", Status: pickString(step, []string{"idle", "typing", "handoff", "done", "done", "done"}), ComputerState: pickString(step, []string{"awaiting-plan", "atomic-task-check", "task-graph-export", "task-graph-archived", "task-graph-archived", "task-graph-archived"}), CurrentTask: pickString(step, []string{"Waiting for the planning blueprint", "Split work into atomic tasks", "Deliver the task graph to execution", "Task graph complete", "Task graph complete", "Task graph complete"}), Detail: "Refuses compound tasks until each node has one owner, one artifact, and one retry path.", Artifact: "task-graph-v5", NextTarget: pickString(step, []string{"Atlas / Chief", "Nova / Builder A", "Rae / Peer Reviewer", "Rae / Peer Reviewer", "Rae / Peer Reviewer", "June / Delivery Manager"}), LastUpdate: stamp(now, 14*time.Second), RiskLevel: "medium", Progress: pickInt(step, []int{0, 74, 100, 100, 100, 100}), QueueDepth: pickInt(step, []int{0, 2, 1, 0, 0, 0}), X: 53, Y: 15, Highlights: []string{"Expanded work into atomic nodes", "Inserted review gates", "Marked one owner per task"}},
		{SortOrder: 4, ID: "builder-a", Name: "Nova", Role: "Implementation Agent", Team: "Execution", DeskLabel: "E-02", ModelTier: "tier-execution", Status: pickString(step, []string{"idle", "queued", "typing", "handoff", "done", "done"}), ComputerState: pickString(step, []string{"awaiting-task", "task-queue", "frontend-shell", "review-bundle-prep", "ui-shell-archived", "ui-shell-archived"}), CurrentTask: pickString(step, []string{"Waiting for decomposition output", "Queued for the UI slice", "Build the office shell and controls", "Package the UI slice for review", "UI slice approved", "UI slice approved"}), Detail: "Owns the office floor, task board, and the visible operator shell.", Artifact: "office-shell.tsx", NextTarget: "Rae / Peer Reviewer", LastUpdate: stamp(now, 20*time.Second), RiskLevel: "low", Progress: pickInt(step, []int{0, 0, 76, 100, 100, 100}), QueueDepth: pickInt(step, []int{0, 1, 1, 1, 0, 0}), X: 32, Y: 46, Highlights: []string{"Replaced the starter UI", "Added control desk and task board", "Hands artifacts into review"}},
		{SortOrder: 5, ID: "builder-b", Name: "Sol", Role: "Implementation Agent", Team: "Execution", DeskLabel: "E-03", ModelTier: "tier-execution", Status: pickString(step, []string{"idle", "queued", "typing", "handoff", "done", "done"}), ComputerState: pickString(step, []string{"awaiting-task", "task-queue", "sqlite-runtime", "review-bundle-prep", "runtime-archived", "runtime-archived"}), CurrentTask: pickString(step, []string{"Waiting for decomposition output", "Queued for the runtime slice", "Persist run, task, gate, worker, and provider state", "Package the runtime slice for review", "Runtime slice approved", "Runtime slice approved"}), Detail: "Owns the local state store that makes the workflow replayable.", Artifact: "maliang-swarm.db", NextTarget: "Rae / Peer Reviewer", LastUpdate: stamp(now, 10*time.Second), RiskLevel: "medium", Progress: pickInt(step, []int{0, 0, 83, 100, 100, 100}), QueueDepth: pickInt(step, []int{0, 1, 2, 1, 0, 0}), X: 46, Y: 46, Highlights: []string{"Stores run, task, gate, worker, and provider state", "Supports reset and settings persistence", "Packages runtime output for review"}},
		{SortOrder: 6, ID: "reviewer", Name: "Rae", Role: "Peer Reviewer", Team: "Review", DeskLabel: "R-01", ModelTier: "tier-review", Status: pickString(step, []string{"idle", "idle", "idle", "reviewing", "done", "done"}), ComputerState: pickString(step, []string{"review-queue-empty", "review-queue-empty", "incoming-bundle-wait", "diff-review", "review-complete", "review-complete"}), CurrentTask: pickString(step, []string{"Waiting for reviewable artifacts", "Waiting for reviewable artifacts", "Waiting for execution bundles", "Check atomic-task fidelity and UI coverage", "Review notes forwarded to security", "Review work complete"}), Detail: "Looks for brittle assumptions, missing review hooks, and unclear ownership before delivery opens.", Artifact: "review-bundle-en24", NextTarget: "Moss / Security Reviewer", LastUpdate: stamp(now, 27*time.Second), RiskLevel: "medium", Progress: pickInt(step, []int{0, 0, 0, 58, 100, 100}), QueueDepth: pickInt(step, []int{0, 0, 1, 2, 1, 0}), X: 73, Y: 43, Highlights: []string{"Verifies tasks are atomic", "Checks review hooks remain visible", "Packages findings for security"}},
		{SortOrder: 7, ID: "security", Name: "Moss", Role: "Security Reviewer", Team: "Delivery", DeskLabel: "D-01", ModelTier: "tier-review", Status: pickString(step, []string{"idle", "idle", "idle", "idle", "blocked", "reviewing"}), ComputerState: pickString(step, []string{"policy-queue-empty", "policy-queue-empty", "policy-queue-empty", "policy-queue-empty", "approval-hold", "approval-verified"}), CurrentTask: pickString(step, []string{"Waiting for risky actions to review", "Waiting for risky actions to review", "Waiting for risky actions to review", "Waiting for risky actions to review", "Pause risky outbound actions until approval arrives", "Record approval and close the security gate"}), Detail: "Treats risky outbound actions as hard stops until a human explicitly clears them.", Artifact: "security-note-7", NextTarget: "June / Delivery Manager", LastUpdate: stamp(now, 40*time.Second), RiskLevel: "high", Progress: pickInt(step, []int{0, 0, 0, 0, 71, 100}), QueueDepth: pickInt(step, []int{0, 0, 0, 0, 1, 0}), X: 73, Y: 68, Highlights: []string{"Blocks risky actions instead of warning", "Waits on explicit human approval", "Writes the hold state into the floor"}},
		{SortOrder: 8, ID: "delivery", Name: "June", Role: "Delivery Manager", Team: "Delivery", DeskLabel: "D-02", ModelTier: "tier-strategic", Status: pickString(step, []string{"idle", "idle", "idle", "idle", "idle", "typing"}), ComputerState: pickString(step, []string{"release-waiting", "release-waiting", "release-waiting", "release-waiting", "release-waiting", "delivery-package"}), CurrentTask: pickString(step, []string{"Waiting for upstream gates to close", "Waiting for upstream gates to close", "Waiting for upstream gates to close", "Waiting for upstream gates to close", "Waiting for security approval", "Assemble the final delivery summary"}), Detail: "Cannot package delivery until review and security both report green.", Artifact: "delivery-summary.md", NextTarget: "Archive", LastUpdate: stamp(now, 58*time.Second), RiskLevel: "low", Progress: pickInt(step, []int{0, 0, 0, 0, 0, 64}), QueueDepth: pickInt(step, []int{0, 0, 0, 0, 0, 1}), X: 88, Y: 68, Highlights: []string{"Waits for every hard gate to close", "Packages one final handoff", "Prepares the release summary"}},
	}
}

func handoffsForStep(step int) []domain.Handoff {
	switch step {
	case 0:
		return nil
	case 1:
		return []domain.Handoff{{SortOrder: 1, ID: "handoff-intake", ArtifactName: "scope-brief-v4", FromAgentID: "intake", ToAgentID: "decomposer", Status: "in_transit"}}
	case 2:
		return []domain.Handoff{{SortOrder: 1, ID: "handoff-dag", ArtifactName: "task-graph-v5", FromAgentID: "decomposer", ToAgentID: "builder-a", Status: "in_transit"}}
	case 3:
		return []domain.Handoff{{SortOrder: 1, ID: "handoff-build", ArtifactName: "execution-bundle-en24", FromAgentID: "builder-b", ToAgentID: "reviewer", Status: "in_transit"}}
	case 4:
		return []domain.Handoff{{SortOrder: 1, ID: "handoff-review", ArtifactName: "review-bundle-en24", FromAgentID: "reviewer", ToAgentID: "security", Status: "queued"}}
	default:
		return []domain.Handoff{{SortOrder: 1, ID: "handoff-delivery", ArtifactName: "delivery-summary.md", FromAgentID: "delivery", ToAgentID: "chief", Status: "in_transit"}}
	}
}

func timelineForStep(step int, now time.Time) []domain.TimelineEvent {
	switch step {
	case 0:
		return []domain.TimelineEvent{{SortOrder: 1, TimeLabel: stamp(now, 5*time.Minute), Kind: "run", Title: "Run created in the command deck", Detail: "The system opened a persisted enterprise run and started intake normalization."}, {SortOrder: 2, TimeLabel: stamp(now, 2*time.Minute), Kind: "planning", Title: "Chief Orchestrator started workflow selection", Detail: "The control deck is choosing the standard enterprise path before dispatching workers."}}
	case 1:
		return []domain.TimelineEvent{{SortOrder: 1, TimeLabel: stamp(now, 7*time.Minute), Kind: "handoff", Title: "Scope brief sent to decomposition", Detail: "Intake delivered the normalized brief into the planning room."}, {SortOrder: 2, TimeLabel: stamp(now, 2*time.Minute), Kind: "decomposition", Title: "Atomic task expansion started", Detail: "The decomposer is splitting the run until every node has one owner and one retry boundary."}}
	case 2:
		return []domain.TimelineEvent{{SortOrder: 1, TimeLabel: stamp(now, 6*time.Minute), Kind: "staffing", Title: "Execution lanes reserved", Detail: "Worker slots were allocated without starving review capacity."}, {SortOrder: 2, TimeLabel: stamp(now, 1*time.Minute), Kind: "execution", Title: "UI and runtime slices entered the execution bay", Detail: "The system is now writing the office shell and the SQLite runtime state."}}
	case 3:
		return []domain.TimelineEvent{{SortOrder: 1, TimeLabel: stamp(now, 5*time.Minute), Kind: "review", Title: "Execution bundle opened in peer review", Detail: "The review corridor started checking task atomicity, UI hooks, and runtime seams."}, {SortOrder: 2, TimeLabel: stamp(now, 1*time.Minute), Kind: "handoff", Title: "Review packet prepared for security", Detail: "Peer review bundled findings and flagged one outbound path for security attention."}}
	case 4:
		return []domain.TimelineEvent{{SortOrder: 1, TimeLabel: stamp(now, 4*time.Minute), Kind: "security", Title: "Security gate blocked a risky action", Detail: "A human approval is required before the run can continue to delivery."}, {SortOrder: 2, TimeLabel: stamp(now, 1*time.Minute), Kind: "audit", Title: "Approval hold written to the runtime ledger", Detail: "The runtime now persists this hold as a replayable event."}}
	default:
		return []domain.TimelineEvent{{SortOrder: 1, TimeLabel: stamp(now, 3*time.Minute), Kind: "security", Title: "Security approval recorded", Detail: "The blocked action was cleared and the gate was closed."}, {SortOrder: 2, TimeLabel: stamp(now, 1*time.Minute), Kind: "delivery", Title: "Delivery package assembly started", Detail: "The delivery manager began aggregating the release summary and audit trail."}}
	}
}

func pickString(step int, values []string) string {
	if step >= len(values) {
		return values[len(values)-1]
	}
	return values[step]
}

func pickInt(step int, values []int) int {
	if step >= len(values) {
		return values[len(values)-1]
	}
	return values[step]
}

func stamp(now time.Time, delta time.Duration) string {
	return fmt.Sprintf("%s HKT", now.Add(-delta).Format("15:04:05"))
}

func compact(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func titleize(value string) string {
	value = strings.ReplaceAll(value, "-", " ")
	parts := strings.Fields(value)
	for i := range parts {
		parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
	}
	return strings.Join(parts, " ")
}
