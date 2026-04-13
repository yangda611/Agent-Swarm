package orchestrator

import (
	"fmt"
	"strings"
	"time"

	"maliangswarm/internal/domain"
)

type workstreamSpec struct {
	id       string
	name     string
	artifact string
	detail   string
}

type planningBlueprint struct {
	Input            RunCreationInput
	Workstreams      []workstreamSpec
	ActiveTemplate   string
	AtomicTaskPolicy string
	ReviewMode       string
	PlanningSource   string
}

func defaultRunInput() RunCreationInput {
	return RunCreationInput{
		Title:                 "maliang swarm 基础工程",
		Mission:               "使用 Go 语言将 maliang swarm 构建为桌面应用，支持企业级多 Agent 编排、像素办公室、角色规划、任务拆解、审核门禁以及可配置的 AI 接口注册表。",
		Deliverable:           "一个可运行的 Windows 桌面应用，包含编排运行时、像素办公室、接口注册表、主设置中心和审核流程。",
		TaskType:              "engineering",
		Priority:              "high",
		MaxAgents:             12,
		RequiresQA:            true,
		RequiresSecurity:      true,
		RequiresHumanApproval: true,
	}
}

func compileRun(input RunCreationInput, settings domain.RuntimeSettings, providers []domain.AIProviderConfig, now time.Time) domain.Snapshot {
	return compileRunFromBlueprint(buildFallbackBlueprint(input), settings, providers, now)
}

func buildFallbackBlueprint(input RunCreationInput) planningBlueprint {
	input = normalizeRunInputForCompiler(input)
	return planningBlueprint{
		Input:            input,
		Workstreams:      deriveWorkstreams(input),
		ActiveTemplate:   localizedTemplateName(input.TaskType),
		AtomicTaskPolicy: "必须拆解到每个任务只有一个负责人、一个产物、一个独立重试边界。",
		ReviewMode:       "默认双人复核，高风险动作必须审批，每个阶段都记录持久化审计轨迹。",
		PlanningSource:   "rules fallback",
	}
}

func compileRunFromBlueprint(blueprint planningBlueprint, settings domain.RuntimeSettings, providers []domain.AIProviderConfig, now time.Time) domain.Snapshot {
	input := normalizeRunInputForCompiler(blueprint.Input)
	settings = normalizeSettings(settings)
	providers = normalizeProviders(providers)
	if len(providers) == 0 {
		providers = defaultAIProviders()
	}

	workstreams := normalizeBlueprintWorkstreams(blueprint.Workstreams, input)
	agents, executionIDs := buildCompiledAgents(input, workstreams)
	tasks := buildCompiledTasks(input, workstreams, executionIDs)
	activeTemplate := pickBlueprintString(blueprint.ActiveTemplate, localizedTemplateName(input.TaskType))
	atomicTaskPolicy := pickBlueprintString(blueprint.AtomicTaskPolicy, "必须拆解到每个任务只有一个负责人、一个产物、一个独立重试边界。")
	reviewMode := pickBlueprintString(blueprint.ReviewMode, "默认双人复核，高风险动作必须审批，每个阶段都记录持久化审计轨迹。")
	planningSource := pickBlueprintString(blueprint.PlanningSource, "rules fallback")

	snapshot := domain.Snapshot{
		Run: domain.Run{
			ID:               buildRunID(input.Title, now),
			Title:            input.Title,
			Mission:          input.Mission,
			Deliverable:      input.Deliverable,
			Priority:         input.Priority,
			TaskType:         input.TaskType,
			Subtitle:         fmt.Sprintf("当前优先级：%s。目标交付物：%s。系统已规划 %d 个 Agent、%d 条工作流，并自动挂载企业审核门禁与可配置 AI 路由。规划来源：%s。", zhPriority(input.Priority), input.Deliverable, len(agents), len(workstreams), zhPlannerSource(planningSource)),
			WorkspaceName:    fmt.Sprintf("%s / %s指挥室", productName, zhTaskType(input.TaskType)),
			ActiveTemplate:   activeTemplate,
			CurrentStage:     "需求接入与流程选择",
			AtomicTaskPolicy: atomicTaskPolicy,
			ReviewMode:       reviewMode,
			PlannerSource:    planningSource,
			SuggestedAgentID: "chief",
			Step:             0,
		},
		Settings:      settings,
		AIProviders:   providers,
		ModelProfiles: modelProfilesForProviders(providers),
		Zones:         defaultZones(),
		Gates:         buildCompiledGates(input),
		Tasks:         tasks,
		Agents:        agents,
	}

	return applyProgression(snapshot, 0, now)
}

func applyProgression(snapshot domain.Snapshot, step int, now time.Time) domain.Snapshot {
	step = max(0, min(step, MaxDemoStep))
	snapshot.Run.Step = step
	snapshot.Run.CurrentStage = pickString(step, []string{
		"需求接入与流程选择",
		"原子拆解与角色排布",
		"并行执行阶段",
		"同行评审与治理整理",
		"安全与审批门禁",
		"交付打包与归档",
	})
	snapshot.Run.SuggestedAgentID = suggestedAgentIDForStep(snapshot.Agents, step)
	snapshot.Gates = progressedGates(snapshot.Gates, step)
	snapshot.Tasks = progressedTasks(snapshot.Tasks, step)
	snapshot.Agents = progressedAgents(snapshot.Agents, snapshot.Tasks, step, now)
	snapshot.Handoffs = progressedHandoffs(snapshot.Agents, step)
	snapshot.Timeline = progressedTimeline(snapshot, step, now)
	return snapshot
}

func normalizeRunInputForCompiler(input RunCreationInput) RunCreationInput {
	input.Title = compact(input.Title)
	input.Mission = compact(input.Mission)
	input.Deliverable = compact(input.Deliverable)
	input.TaskType = strings.ToLower(compact(input.TaskType))
	input.Priority = strings.ToLower(compact(input.Priority))

	if input.Mission == "" {
		input.Mission = defaultRunInput().Mission
	}
	if input.Title == "" {
		input.Title = deriveTitle(input.Mission)
	}
	if input.TaskType == "" {
		input.TaskType = inferTaskType(input.Mission)
	}
	if input.Priority == "" {
		input.Priority = "medium"
	}
	if input.Deliverable == "" {
		input.Deliverable = defaultDeliverableForType(input.TaskType)
	}
	if input.MaxAgents <= 0 {
		input.MaxAgents = 12
	}
	input.MaxAgents = min(input.MaxAgents, 100)
	if input.TaskType == "engineering" {
		input.RequiresQA = true
	}
	if containsAnyLower(input.Mission, "security", "approval", "publish", "automation", "sensitive") {
		input.RequiresSecurity = true
	}
	if input.RequiresSecurity && containsAnyLower(input.Mission, "approval", "publish", "automation", "sensitive") {
		input.RequiresHumanApproval = true
	}
	return input
}

func defaultDeliverableForType(taskType string) string {
	switch taskType {
	case "research":
		return "一份带证据、结论与建议的审核后研究简报。"
	case "product":
		return "一份产品方案、流程设计和面向操作者的界面概念稿。"
	case "content":
		return "一份经过事实核验并完成最终排版的内容交付包。"
	default:
		return "一个可运行的实现包，包含企业流程、审核门禁和交付说明。"
	}
}

func deriveWorkstreams(input RunCreationInput) []workstreamSpec {
	var workstreams []workstreamSpec
	switch input.TaskType {
	case "research":
		workstreams = []workstreamSpec{
			{id: "evidence", name: "证据收集", artifact: "evidence-pack.md", detail: "收集与任务相关的结构化证据。"},
			{id: "synthesis", name: "结论综合", artifact: "synthesis-brief.md", detail: "将证据提炼成可审核的建议集合。"},
			{id: "fact-check", name: "事实核验", artifact: "fact-check.md", detail: "核对论点、来源与可信度信号。"},
		}
	case "product":
		workstreams = []workstreamSpec{
			{id: "requirements", name: "需求地图", artifact: "requirements-map.md", detail: "沉淀用户意图、范围与流程约束。"},
			{id: "workflow", name: "流程设计", artifact: "workflow-design.md", detail: "把目标翻译成带门禁的阶段与责任边界。"},
			{id: "operator-ui", name: "操作台界面", artifact: "operator-shell.tsx", detail: "规划面向操作者的指挥台和任务面板。"},
		}
	case "content":
		workstreams = []workstreamSpec{
			{id: "outline", name: "内容大纲", artifact: "content-outline.md", detail: "规划整体结构与交付意图。"},
			{id: "draft", name: "内容草稿", artifact: "content-draft.md", detail: "产出交付物主体内容。"},
			{id: "formatting", name: "排版整理", artifact: "formatted-package.md", detail: "准备最终发布包。"},
		}
	default:
		workstreams = []workstreamSpec{
			{id: "orchestrator-core", name: "编排核心", artifact: "orchestrator-core.go", detail: "负责编译运行、规划角色并把任务送入企业流程阶段。"},
			{id: "runtime-store", name: "运行时存储", artifact: "runtime-store.db", detail: "持久化运行、任务、门禁、Agent 与时间线状态。"},
			{id: "operator-ui", name: "操作台界面", artifact: "operator-shell.tsx", detail: "构建主控台、任务板和右侧抽屉。"},
			{id: "pixel-office", name: "像素办公室场景", artifact: "office-scene.tsx", detail: "把 Agent、交接动作和办公分区投影到可视化办公室。"},
		}
	}

	mission := strings.ToLower(input.Mission)
	if containsAnyLower(mission, "pixel", "office", "sprite", "drawer") {
		workstreams = append(workstreams, workstreamSpec{id: "pixel-experience", name: "像素体验", artifact: "pixel-experience.tsx", detail: "继续打磨办公室场景、移动动画与抽屉交互。"})
	}
	if containsAnyLower(mission, "go", "sqlite", "backend", "runtime") {
		workstreams = append(workstreams, workstreamSpec{id: "backend-runtime", name: "后端运行时", artifact: "backend-runtime.go", detail: "实现 Go 编排核心与可持久化状态流转。"})
	}
	if containsAnyLower(mission, "review", "security", "approval", "audit", "compliance") {
		workstreams = append(workstreams, workstreamSpec{id: "governance", name: "治理层", artifact: "governance-policy.md", detail: "把审核、审批和审计要求嵌入工作流。"})
	}
	if containsAnyLower(mission, "api", "provider", "model", "llm", "gateway") {
		workstreams = append(workstreams, workstreamSpec{id: "provider-registry", name: "AI 接口注册表", artifact: "provider-registry.json", detail: "在控制中心管理多种 API 格式、密钥和默认路由策略。"})
	}

	seen := map[string]bool{}
	unique := make([]workstreamSpec, 0, len(workstreams))
	for _, stream := range workstreams {
		if seen[stream.id] {
			continue
		}
		seen[stream.id] = true
		unique = append(unique, stream)
	}
	return unique
}

func normalizeBlueprintWorkstreams(workstreams []workstreamSpec, input RunCreationInput) []workstreamSpec {
	if len(workstreams) == 0 {
		return deriveWorkstreams(input)
	}

	normalized := make([]workstreamSpec, 0, len(workstreams))
	seen := map[string]bool{}
	for index, stream := range workstreams {
		id := compact(stream.id)
		name := compact(stream.name)
		artifact := compact(stream.artifact)
		detail := compact(stream.detail)

		if name == "" {
			name = fmt.Sprintf("工作流 %d", index+1)
		}
		if id == "" {
			id = slugify(name)
		}
		if artifact == "" {
			artifact = fmt.Sprintf("%s.md", id)
		}
		if detail == "" {
			detail = fmt.Sprintf("执行并交接%s这条工作流。", name)
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		normalized = append(normalized, workstreamSpec{id: id, name: name, artifact: artifact, detail: detail})
	}

	if len(normalized) == 0 {
		return deriveWorkstreams(input)
	}

	return normalized
}

func pickBlueprintString(value string, fallback string) string {
	if compact(value) == "" {
		return fallback
	}
	return compact(value)
}

func buildCompiledAgents(input RunCreationInput, workstreams []workstreamSpec) ([]domain.Agent, []string) {
	agents := []domain.Agent{
		baseAgent("chief", "Atlas", "Chief Orchestrator", "Command", "C-01", "tier-strategic", 12, 15, "workflow-blueprint.json", []string{"Owns the workflow choice", "Reserves review capacity", "Keeps enterprise gates visible"}),
		baseAgent("intake", "Mina", "Intake Analyst", "Command", "C-02", "tier-routing", 20, 15, "scope-brief.md", []string{"Normalizes the mission into structured constraints", "Captures concurrency and review requirements", "Feeds the planning lane"}),
		baseAgent("planner", "Jules", "Solution Planner", "Planning", "P-01", "tier-strategic", 38, 15, "delivery-plan.md", []string{"Maps the mission into staged workstreams", "Keeps review lanes explicit", "Prepares the staffing outline"}),
		baseAgent("decomposer", "Oren", "Task Decomposer", "Planning", "P-02", "tier-strategic", 53, 15, "task-graph.json", []string{"Splits work into atomic tasks", "Assigns one owner per output", "Marks review boundaries"}),
		baseAgent("reviewer", "Rae", "Peer Reviewer", "Review", "R-01", "tier-review", 73, 43, "review-bundle.md", []string{"Checks task atomicity and integration seams", "Verifies reviewer visibility", "Packages findings for governance lanes"}),
		baseAgent("delivery", "June", "Delivery Manager", "Delivery", "D-02", "tier-strategic", 88, 68, "delivery-summary.md", []string{"Aggregates final outputs", "Waits for every hard gate to close", "Packages the release handoff"}),
		baseAgent("audit", "Nori", "Audit Recorder", "Delivery", "D-03", "tier-routing", 80, 68, "audit-ledger.jsonl", []string{"Captures approvals and lineage", "Emits replayable stage history", "Seals the delivery ledger"}),
	}

	if input.RequiresQA {
		agents = append(agents, baseAgent("qa", "Ivy", "QA Verifier", "Review", "R-02", "tier-review", 87, 43, "acceptance-grid.md", []string{"Runs acceptance against persisted workflow", "Checks operator controls", "Feeds validated bundles into delivery"}))
	}
	if input.RequiresSecurity {
		agents = append(agents, baseAgent("security", "Moss", "Security Reviewer", "Delivery", "D-01", "tier-review", 73, 68, "security-note.md", []string{"Reviews risky outbound actions", "Treats policy holds as hard stops", "Waits for explicit clearance"}))
	}
	if input.RequiresHumanApproval {
		agents = append(agents, baseAgent("approver", "Ari", "Human Approval Proxy", "Delivery", "D-04", "tier-review", 66, 68, "approval-ticket.md", []string{"Turns blocked actions into approval tickets", "Waits on explicit human decision", "Releases the run once approval lands"}))
	}

	execNames := []string{"Nova", "Sol", "Tao", "Kira", "Vega", "Ivo", "Mira", "Pax"}
	execSeats := [][2]int{{18, 46}, {32, 46}, {46, 46}, {18, 60}, {32, 60}, {46, 60}, {10, 53}, {54, 53}}
	execCount := executionAgentCount(input, len(workstreams), len(agents))
	executionIDs := make([]string, 0, execCount)

	for i := 0; i < execCount; i++ {
		id := fmt.Sprintf("exec-%d", i+1)
		seat := execSeats[i%len(execSeats)]
		agents = append(agents, baseAgent(id, execNames[i%len(execNames)], "Implementation Agent", "Execution", fmt.Sprintf("E-%02d", i+1), "tier-execution", seat[0], seat[1], "work-bundle.md", []string{"Owns one or more atomic execution tasks", "Packages artifacts for review", "Works inside the execution bay"}))
		executionIDs = append(executionIDs, id)
	}

	return agents, executionIDs
}

func buildCompiledTasks(input RunCreationInput, workstreams []workstreamSpec, executionIDs []string) []domain.Task {
	tasks := []domain.Task{
		{SortOrder: 1, ID: "task-intake-normalize", Title: "整理并标准化任务需求", Stage: "intake", OwnerAgentID: "intake", OwnerRole: "Intake Analyst", ArtifactName: "scope-brief.md", Status: "queued", RiskLevel: "low", Detail: "提炼范围、约束、交付预期和企业流程要求。", ReviewRequired: false, Dependencies: 0},
		{SortOrder: 2, ID: "task-plan-template", Title: "选择合适的企业流程模板", Stage: "planning", OwnerAgentID: "chief", OwnerRole: "Chief Orchestrator", ArtifactName: "workflow-template.json", Status: "queued", RiskLevel: "medium", Detail: "根据任务类型、优先级和治理要求选择流程模板。", ReviewRequired: true, Dependencies: 1},
		{SortOrder: 3, ID: "task-plan-decompose", Title: "拆解为原子任务", Stage: "planning", OwnerAgentID: "decomposer", OwnerRole: "Task Decomposer", ArtifactName: "task-graph.json", Status: "queued", RiskLevel: "medium", Detail: "持续拆解，直到每个任务都只有一个负责人、一个产物和一个审核边界。", ReviewRequired: true, Dependencies: 2},
		{SortOrder: 4, ID: "task-plan-staff", Title: "规划角色编制与工位覆盖", Stage: "planning", OwnerAgentID: "planner", OwnerRole: "Solution Planner", ArtifactName: "staffing-plan.json", Status: "queued", RiskLevel: "medium", Detail: fmt.Sprintf("在最多 %d 个 Agent 内完成角色排布，同时保留审核产能。", input.MaxAgents), ReviewRequired: true, Dependencies: 2},
	}

	sortOrder := 5
	for index, stream := range workstreams {
		execID := executionIDs[index%len(executionIDs)]
		tasks = append(tasks, domain.Task{
			SortOrder:      sortOrder,
			ID:             fmt.Sprintf("task-exec-%s", stream.id),
			Title:          fmt.Sprintf("执行工作流：%s", stream.name),
			Stage:          "execution",
			OwnerAgentID:   execID,
			OwnerRole:      "Implementation Agent",
			ArtifactName:   stream.artifact,
			Status:         "queued",
			RiskLevel:      workstreamRisk(stream.id),
			Detail:         stream.detail,
			ReviewRequired: true,
			Dependencies:   2,
		})
		sortOrder++
	}

	tasks = append(tasks, domain.Task{SortOrder: sortOrder, ID: "task-review-peer", Title: "同行评审执行产物包", Stage: "review", OwnerAgentID: "reviewer", OwnerRole: "Peer Reviewer", ArtifactName: "review-bundle.md", Status: "queued", RiskLevel: "medium", Detail: "在下游门禁打开前，检查执行结果、责任边界和交接清晰度。", ReviewRequired: false, Dependencies: len(workstreams)})
	sortOrder++

	if input.RequiresQA {
		tasks = append(tasks, domain.Task{SortOrder: sortOrder, ID: "task-review-qa", Title: "执行 QA 验收检查", Stage: "review", OwnerAgentID: "qa", OwnerRole: "QA Verifier", ArtifactName: "acceptance-grid.md", Status: "queued", RiskLevel: "medium", Detail: "根据验收标准核对操作台、运行时状态和阶段产出。", ReviewRequired: false, Dependencies: 1})
		sortOrder++
	}
	if input.RequiresSecurity {
		tasks = append(tasks, domain.Task{SortOrder: sortOrder, ID: "task-review-security", Title: "审核高风险动作与治理边界", Stage: "review", OwnerAgentID: "security", OwnerRole: "Security Reviewer", ArtifactName: "security-note.md", Status: "queued", RiskLevel: "high", Detail: "检查外发动作、敏感操作和策略敏感的状态迁移。", ReviewRequired: true, Dependencies: 1})
		sortOrder++
	}
	if input.RequiresHumanApproval {
		tasks = append(tasks, domain.Task{SortOrder: sortOrder, ID: "task-review-approval", Title: "等待人工明确审批", Stage: "review", OwnerAgentID: "approver", OwnerRole: "Human Approval Proxy", ArtifactName: "approval-ticket.md", Status: "queued", RiskLevel: "high", Detail: "把阻塞动作转换为审批请求，并在放行前等待明确批准。", ReviewRequired: false, Dependencies: 1})
		sortOrder++
	}

	tasks = append(tasks,
		domain.Task{SortOrder: sortOrder, ID: "task-delivery-audit", Title: "汇总审计台账", Stage: "delivery", OwnerAgentID: "audit", OwnerRole: "Audit Recorder", ArtifactName: "audit-ledger.jsonl", Status: "queued", RiskLevel: "low", Detail: "收集血缘、审批、门禁流转和产物历史，形成可回放台账。", ReviewRequired: false, Dependencies: 1},
		domain.Task{SortOrder: sortOrder + 1, ID: "task-delivery-package", Title: "打包最终交付物", Stage: "delivery", OwnerAgentID: "delivery", OwnerRole: "Delivery Manager", ArtifactName: "delivery-summary.md", Status: "queued", RiskLevel: "low", Detail: fmt.Sprintf("把最终结果整理为目标交付物：%s", input.Deliverable), ReviewRequired: true, Dependencies: 2},
	)

	return tasks
}

func buildCompiledGates(input RunCreationInput) []domain.WorkflowGate {
	gates := []domain.WorkflowGate{
		{SortOrder: 1, ID: "gate-intake", Label: "需求标准化", Status: "queued"},
		{SortOrder: 2, ID: "gate-plan", Label: "规划评审", Status: "queued"},
		{SortOrder: 3, ID: "gate-staff", Label: "编制与任务覆盖检查", Status: "queued"},
		{SortOrder: 4, ID: "gate-peer", Label: "执行产物同行评审", Status: "queued"},
	}
	order := 5
	if input.RequiresQA {
		gates = append(gates, domain.WorkflowGate{SortOrder: order, ID: "gate-qa", Label: "QA 验收", Status: "queued"})
		order++
	}
	if input.RequiresSecurity {
		gates = append(gates, domain.WorkflowGate{SortOrder: order, ID: "gate-security", Label: "安全审核", Status: "queued"})
		order++
	}
	if input.RequiresHumanApproval {
		gates = append(gates, domain.WorkflowGate{SortOrder: order, ID: "gate-approval", Label: "人工审批", Status: "queued"})
		order++
	}
	gates = append(gates, domain.WorkflowGate{SortOrder: order, ID: "gate-delivery", Label: "最终交付签收", Status: "queued"})
	return gates
}

func progressedGates(gates []domain.WorkflowGate, step int) []domain.WorkflowGate {
	hasApprovalGate := false
	for _, gate := range gates {
		if gate.ID == "gate-approval" {
			hasApprovalGate = true
			break
		}
	}

	updated := make([]domain.WorkflowGate, 0, len(gates))
	for _, gate := range gates {
		next := gate
		switch gate.ID {
		case "gate-intake":
			next.Status = lifecycleStatus(step, 0)
		case "gate-plan", "gate-staff":
			next.Status = lifecycleStatus(step, 1)
		case "gate-peer", "gate-qa":
			next.Status = lifecycleStatus(step, 3)
		case "gate-security":
			if step < 4 {
				next.Status = "queued"
			} else if step == 4 {
				if hasApprovalGate {
					next.Status = "blocked"
				} else {
					next.Status = "in_progress"
				}
			} else {
				next.Status = "completed"
			}
		case "gate-approval":
			if step < 4 {
				next.Status = "queued"
			} else if step == 4 {
				next.Status = "in_progress"
			} else {
				next.Status = "completed"
			}
		case "gate-delivery":
			if step < 5 {
				next.Status = "queued"
			} else {
				next.Status = "completed"
			}
		}
		updated = append(updated, next)
	}
	return updated
}

func progressedTasks(tasks []domain.Task, step int) []domain.Task {
	updated := make([]domain.Task, 0, len(tasks))
	for _, task := range tasks {
		next := task
		next.Status = taskLifecycleStatus(task.ID, step)
		next.Progress = taskProgressForStatus(task.ID, next.Status)
		updated = append(updated, next)
	}
	return updated
}

func progressedAgents(agents []domain.Agent, tasks []domain.Task, step int, now time.Time) []domain.Agent {
	updated := make([]domain.Agent, 0, len(agents))
	for _, agent := range agents {
		next := agent
		owned := tasksOwnedBy(tasks, agent.ID)
		active, hasActive := activeOwnedTask(owned)
		next.QueueDepth = outstandingCount(owned)
		next.LastUpdate = stamp(now, time.Duration(agent.SortOrder*7)*time.Second)
		next.NextTarget = nextTargetForAgent(agent.ID, agents)

		if hasActive {
			next.CurrentTask = active.Title
			next.Detail = active.Detail
			next.Progress = active.Progress
			next.RiskLevel = active.RiskLevel
			next.Status = roleStatus(agent, active.Status)
			next.ComputerState = computerStateForStatus(agent, next.Status)
		} else if len(owned) > 0 && allOwnedCompleted(owned) {
			next.CurrentTask = "当前负责的任务已经完成，资料已打包等待下游接收。"
			next.Progress = 100
			next.Status = doneStatus(agent)
			next.ComputerState = computerStateForStatus(agent, next.Status)
		} else {
			next.CurrentTask = waitingMessage(agent.Team)
			next.Progress = min(step*18, 100)
			next.Status = idleStatus(agent, step)
			next.ComputerState = computerStateForStatus(agent, next.Status)
		}

		updated = append(updated, next)
	}
	return updated
}

func progressedHandoffs(agents []domain.Agent, step int) []domain.Handoff {
	firstExec := firstAgentIDByTeam(agents, "Execution")
	lastExec := lastAgentIDByTeam(agents, "Execution")
	switch step {
	case 0:
		return nil
	case 1:
		return []domain.Handoff{{SortOrder: 1, ID: "handoff-intake", ArtifactName: "scope-brief.md", FromAgentID: "intake", ToAgentID: "decomposer", Status: "in_transit"}}
	case 2:
		return []domain.Handoff{{SortOrder: 1, ID: "handoff-plan", ArtifactName: "task-graph.json", FromAgentID: "decomposer", ToAgentID: firstExec, Status: "in_transit"}}
	case 3:
		return []domain.Handoff{{SortOrder: 1, ID: "handoff-review", ArtifactName: "execution-bundle.md", FromAgentID: lastExec, ToAgentID: "reviewer", Status: "in_transit"}}
	case 4:
		target := "delivery"
		if hasAgentID(agents, "security") {
			target = "security"
		} else if hasAgentID(agents, "qa") {
			target = "qa"
		}
		return []domain.Handoff{{SortOrder: 1, ID: "handoff-governance", ArtifactName: "review-bundle.md", FromAgentID: "reviewer", ToAgentID: target, Status: "queued"}}
	default:
		return []domain.Handoff{{SortOrder: 1, ID: "handoff-delivery", ArtifactName: "delivery-summary.md", FromAgentID: "delivery", ToAgentID: "audit", Status: "in_transit"}}
	}
}

func progressedTimeline(snapshot domain.Snapshot, step int, now time.Time) []domain.TimelineEvent {
	executionCount := teamCount(snapshot.Agents, "Execution")
	switch step {
	case 0:
		return []domain.TimelineEvent{
			{SortOrder: 1, TimeLabel: stamp(now, 5*time.Minute), Kind: "run", Title: fmt.Sprintf("已为 %s 创建运行实例", snapshot.Run.Title), Detail: "控制台已根据提交的任务简报创建新的企业级运行。"},
			{SortOrder: 2, TimeLabel: stamp(now, 2*time.Minute), Kind: "intake", Title: "需求接入开始标准化", Detail: "系统已将范围、优先级、审核要求和交付目标整理为规划输入。"},
		}
	case 1:
		return []domain.TimelineEvent{
			{SortOrder: 1, TimeLabel: stamp(now, 6*time.Minute), Kind: "planning", Title: "规划区已选定流程模板", Detail: fmt.Sprintf("系统选择了 %s，并正式开启任务拆解阶段。", snapshot.Run.ActiveTemplate)},
			{SortOrder: 2, TimeLabel: stamp(now, 1*time.Minute), Kind: "decomposition", Title: "原子任务和编制方案已生成", Detail: fmt.Sprintf("当前运行已拆解为 %d 个持久化任务，并规划 %d 个 Agent。", len(snapshot.Tasks), len(snapshot.Agents))},
		}
	case 2:
		return []domain.TimelineEvent{
			{SortOrder: 1, TimeLabel: stamp(now, 7*time.Minute), Kind: "staffing", Title: "执行工位已完成排布", Detail: fmt.Sprintf("系统已分配 %d 个执行 Agent，同时保留足够审核产能。", executionCount)},
			{SortOrder: 2, TimeLabel: stamp(now, 1*time.Minute), Kind: "execution", Title: "执行工作流正式打开", Detail: "专业执行 Agent 已开始处理规划阶段产出的原子任务包。"},
		}
	case 3:
		return []domain.TimelineEvent{
			{SortOrder: 1, TimeLabel: stamp(now, 5*time.Minute), Kind: "review", Title: "执行产物已进入同行评审", Detail: "审核区开始检查责任边界、界面覆盖度和交接清晰度。"},
			{SortOrder: 2, TimeLabel: stamp(now, 1*time.Minute), Kind: "handoff", Title: "执行资料进入治理链路", Detail: "完成的工作包已打包并送往评审、QA 与策略检查节点。"},
		}
	case 4:
		return []domain.TimelineEvent{
			{SortOrder: 1, TimeLabel: stamp(now, 4*time.Minute), Kind: "security", Title: "治理挂起已生效", Detail: "高风险动作或审批敏感流转已被显式标记为硬门禁。"},
			{SortOrder: 2, TimeLabel: stamp(now, 1*time.Minute), Kind: "audit", Title: "审计台账已记录挂起状态", Detail: "运行时已持久化本次治理挂起及其血缘信息，支持后续回放。"},
		}
	default:
		return []domain.TimelineEvent{
			{SortOrder: 1, TimeLabel: stamp(now, 3*time.Minute), Kind: "delivery", Title: "交付包已完成组装", Detail: "交付区已整合目标产物、审计台账和签收记录。"},
			{SortOrder: 2, TimeLabel: stamp(now, 1*time.Minute), Kind: "archive", Title: "运行已带完整审计轨迹归档", Detail: "当前办公室画面反映的是一个可回放、可追踪的完整企业级运行。"},
		}
	}
}

func executionAgentCount(input RunCreationInput, workstreamCount int, fixedAgentCount int) int {
	effectiveMax := max(input.MaxAgents, fixedAgentCount+1)
	available := max(1, effectiveMax-fixedAgentCount)
	base := 2
	if len(strings.Fields(input.Mission)) > 30 {
		base++
	}
	if input.Priority == "high" || input.Priority == "critical" {
		base++
	}
	return min(workstreamCount, max(1, min(base+(workstreamCount/2), available)))
}

func baseAgent(id, name, role, team, desk, modelTier string, x, y int, artifact string, highlights []string) domain.Agent {
	return domain.Agent{ID: id, Name: name, Role: role, Team: team, DeskLabel: desk, ModelTier: modelTier, Status: "idle", Artifact: artifact, Progress: 0, X: x, Y: y, Highlights: highlights}
}

func workstreamRisk(id string) string {
	if containsAnyLower(id, "runtime", "backend", "governance", "provider") {
		return "medium"
	}
	return "low"
}

func lifecycleStatus(step int, activeStep int) string {
	if step < activeStep {
		return "queued"
	}
	if step == activeStep {
		return "in_progress"
	}
	return "completed"
}

func taskLifecycleStatus(id string, step int) string {
	switch {
	case strings.HasPrefix(id, "task-intake"):
		return lifecycleStatus(step, 0)
	case strings.HasPrefix(id, "task-plan"):
		return lifecycleStatus(step, 1)
	case strings.HasPrefix(id, "task-exec"):
		return lifecycleStatus(step, 2)
	case id == "task-review-peer" || id == "task-review-qa":
		return lifecycleStatus(step, 3)
	case id == "task-review-security":
		if step < 4 {
			return "queued"
		}
		if step == 4 {
			return "blocked"
		}
		return "completed"
	case id == "task-review-approval":
		if step < 4 {
			return "queued"
		}
		if step == 4 {
			return "in_progress"
		}
		return "completed"
	default:
		if step < 5 {
			return "queued"
		}
		return "completed"
	}
}

func taskProgressForStatus(id string, status string) int {
	switch status {
	case "completed":
		return 100
	case "blocked":
		return 74
	case "in_progress":
		switch {
		case strings.HasPrefix(id, "task-intake"):
			return 48
		case strings.HasPrefix(id, "task-plan"):
			return 67
		case strings.HasPrefix(id, "task-exec"):
			return 79
		case strings.HasPrefix(id, "task-review"):
			return 58
		default:
			return 84
		}
	default:
		return 0
	}
}

func tasksOwnedBy(tasks []domain.Task, agentID string) []domain.Task {
	owned := []domain.Task{}
	for _, task := range tasks {
		if task.OwnerAgentID == agentID {
			owned = append(owned, task)
		}
	}
	return owned
}

func activeOwnedTask(tasks []domain.Task) (domain.Task, bool) {
	for _, task := range tasks {
		if task.Status == "blocked" {
			return task, true
		}
	}
	for _, task := range tasks {
		if task.Status == "in_progress" {
			return task, true
		}
	}
	return domain.Task{}, false
}

func outstandingCount(tasks []domain.Task) int {
	count := 0
	for _, task := range tasks {
		if task.Status == "queued" || task.Status == "in_progress" || task.Status == "blocked" {
			count++
		}
	}
	return count
}

func allOwnedCompleted(tasks []domain.Task) bool {
	if len(tasks) == 0 {
		return false
	}
	for _, task := range tasks {
		if task.Status != "completed" {
			return false
		}
	}
	return true
}

func roleStatus(agent domain.Agent, taskStatus string) string {
	if taskStatus == "blocked" {
		return "blocked"
	}
	if strings.Contains(strings.ToLower(agent.Role), "review") || strings.Contains(strings.ToLower(agent.Role), "approval") || agent.Team == "Review" {
		return "reviewing"
	}
	if agent.Team == "Planning" || agent.Team == "Command" {
		return "thinking"
	}
	return "typing"
}

func doneStatus(agent domain.Agent) string {
	if agent.ID == "audit" {
		return "typing"
	}
	return "done"
}

func idleStatus(agent domain.Agent, step int) string {
	if agent.ID == "audit" && step < 5 {
		return "typing"
	}
	if agent.Team == "Review" && step >= 3 && step < 5 {
		return "waiting_approval"
	}
	return "idle"
}

func computerStateForStatus(agent domain.Agent, status string) string {
	switch status {
	case "blocked":
		return "approval-hold"
	case "reviewing":
		return "review-panel"
	case "thinking":
		return "planning-console"
	case "typing":
		if agent.ID == "audit" {
			return "trace-capture"
		}
		return "workstation-active"
	case "waiting_approval":
		return "approval-queue"
	case "done":
		return "archive-ready"
	default:
		return "awaiting-work"
	}
}

func waitingMessage(team string) string {
	switch team {
	case "Review":
		return "等待上游资料进入审核区。"
	case "Delivery":
		return "等待审核与治理门禁全部关闭。"
	case "Execution":
		return "等待规划区释放下一批原子任务包。"
	default:
		return "等待上游规划信号。"
	}
}

func suggestedAgentIDForStep(agents []domain.Agent, step int) string {
	switch step {
	case 0:
		return "chief"
	case 1:
		return "decomposer"
	case 2:
		return firstAgentIDByTeam(agents, "Execution")
	case 3:
		return "reviewer"
	case 4:
		if hasAgentID(agents, "security") {
			return "security"
		}
		if hasAgentID(agents, "approver") {
			return "approver"
		}
		return "delivery"
	default:
		return "delivery"
	}
}

func nextTargetForAgent(agentID string, agents []domain.Agent) string {
	switch agentID {
	case "chief":
		return agentLabelByID(agents, "intake")
	case "intake":
		return agentLabelByID(agents, "decomposer")
	case "planner", "decomposer":
		return agentLabelByID(agents, firstAgentIDByTeam(agents, "Execution"))
	case "reviewer":
		if hasAgentID(agents, "qa") {
			return agentLabelByID(agents, "qa")
		}
		if hasAgentID(agents, "security") {
			return agentLabelByID(agents, "security")
		}
		return agentLabelByID(agents, "delivery")
	case "qa":
		if hasAgentID(agents, "security") {
			return agentLabelByID(agents, "security")
		}
		return agentLabelByID(agents, "delivery")
	case "security":
		if hasAgentID(agents, "approver") {
			return agentLabelByID(agents, "approver")
		}
		return agentLabelByID(agents, "delivery")
	case "approver":
		return agentLabelByID(agents, "delivery")
	case "delivery":
		return agentLabelByID(agents, "audit")
	case "audit":
		return agentLabelByID(agents, "chief")
	default:
		return agentLabelByID(agents, "reviewer")
	}
}

func firstAgentIDByTeam(agents []domain.Agent, team string) string {
	for _, agent := range agents {
		if agent.Team == team {
			return agent.ID
		}
	}
	return ""
}

func lastAgentIDByTeam(agents []domain.Agent, team string) string {
	for i := len(agents) - 1; i >= 0; i-- {
		if agents[i].Team == team {
			return agents[i].ID
		}
	}
	return ""
}

func hasAgentID(agents []domain.Agent, id string) bool {
	for _, agent := range agents {
		if agent.ID == id {
			return true
		}
	}
	return false
}

func agentLabelByID(agents []domain.Agent, id string) string {
	for _, agent := range agents {
		if agent.ID == id {
			return fmt.Sprintf("%s / %s", agent.Name, agent.Role)
		}
	}
	return "Archive"
}

func teamCount(agents []domain.Agent, team string) int {
	count := 0
	for _, agent := range agents {
		if agent.Team == team {
			count++
		}
	}
	return count
}

func buildRunID(title string, now time.Time) string {
	return fmt.Sprintf("run-%s-%d", slugify(title), now.Unix())
}

func slugify(value string) string {
	value = strings.ToLower(value)
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "mission"
	}
	return result
}

func deriveTitle(mission string) string {
	words := strings.Fields(mission)
	if len(words) > 7 {
		words = words[:7]
	}
	if len(words) == 0 {
		return "New Maliang Swarm Run"
	}
	return titleize(strings.Join(words, " "))
}

func inferTaskType(mission string) string {
	mission = strings.ToLower(mission)
	switch {
	case containsAnyLower(mission, "research", "evidence", "literature", "analysis"):
		return "research"
	case containsAnyLower(mission, "content", "article", "copy", "write"):
		return "content"
	case containsAnyLower(mission, "product", "roadmap", "requirements", "ux"):
		return "product"
	default:
		return "engineering"
	}
}

func containsAnyLower(value string, candidates ...string) bool {
	value = strings.ToLower(value)
	for _, candidate := range candidates {
		if strings.Contains(value, strings.ToLower(candidate)) {
			return true
		}
	}
	return false
}
