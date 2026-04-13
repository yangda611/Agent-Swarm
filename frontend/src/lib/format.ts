const stageLabels: Record<string, string> = {
    intake: "需求接入",
    planning: "规划拆解",
    execution: "执行开发",
    review: "审核复核",
    delivery: "交付归档",
};

const statusLabels: Record<string, string> = {
    queued: "排队中",
    in_progress: "进行中",
    completed: "已完成",
    blocked: "已阻塞",
    reviewing: "审核中",
    thinking: "规划中",
    typing: "执行中",
    idle: "待命",
    done: "已收口",
    handoff: "交接中",
    waiting_approval: "待审批",
    in_transit: "运送中",
};

const riskLabels: Record<string, string> = {
    low: "低风险",
    medium: "中风险",
    high: "高风险",
};

const roleLabels: Record<string, string> = {
    "Chief Orchestrator": "主脑指挥官",
    "Intake Analyst": "需求分析员",
    "Solution Planner": "方案规划师",
    "Task Decomposer": "任务拆解师",
    "Implementation Agent": "执行智能体",
    "Peer Reviewer": "同行评审员",
    "QA Verifier": "质量验证员",
    "Security Reviewer": "安全审核员",
    "Human Approval Proxy": "人工审批代理",
    "Delivery Manager": "交付经理",
    "Audit Recorder": "审计记录员",
    Chief: "主脑",
    Archive: "归档区",
    "Builder A": "执行智能体 A",
};

const teamLabels: Record<string, string> = {
    Command: "指挥区",
    Planning: "规划区",
    Execution: "执行区",
    Review: "审核区",
    Delivery: "交付区",
};

const tierLabels: Record<string, string> = {
    "tier-strategic": "战略模型",
    "tier-review": "审核模型",
    "tier-execution": "执行模型",
    "tier-routing": "路由模型",
};

const computerStateLabels: Record<string, string> = {
    "approval-hold": "审批挂起",
    "review-panel": "审核面板",
    "planning-console": "规划控制台",
    "trace-capture": "审计追踪",
    "workstation-active": "工作站运行中",
    "approval-queue": "审批队列",
    "archive-ready": "归档就绪",
    "awaiting-work": "等待任务",
    "workflow-selection": "流程选择中",
    "enterprise-guardrails": "企业门禁校验",
    "scheduler-overview": "调度总览",
    "review-pressure": "评审负载监控",
    "approval-watch": "审批等待中",
    "release-readiness": "交付签收准备",
    "requirements-normalization": "需求标准化",
    "brief-packaging": "简报打包中",
    "brief-archived": "简报已归档",
    "awaiting-plan": "等待规划信号",
    "atomic-task-check": "原子任务检查",
    "task-graph-export": "任务图导出中",
    "task-graph-archived": "任务图已归档",
    "awaiting-task": "等待任务分派",
    "task-queue": "任务队列",
    "frontend-shell": "前端外壳开发",
    "review-bundle-prep": "评审包准备中",
    "ui-shell-archived": "界面产物已归档",
    "sqlite-runtime": "运行时持久化",
    "runtime-archived": "运行时产物已归档",
    "review-queue-empty": "审核队列为空",
    "incoming-bundle-wait": "等待执行产物",
    "diff-review": "差异审查中",
    "review-complete": "评审完成",
    "policy-queue-empty": "策略队列为空",
    "approval-verified": "审批已核验",
    "release-waiting": "等待交付放行",
    "delivery-package": "交付打包中",
};

const providerFormatLabels: Record<string, string> = {
    "openai-compatible": "兼容格式",
    anthropic: "Anthropic",
    gemini: "Gemini",
    "azure-openai": "Azure OpenAI",
    ollama: "Ollama",
    "custom-http": "自定义接口",
};

const themeLabels: Record<string, string> = {
    "pixel-hq-amber": "像素总部·琥珀",
    "pixel-hq-mint": "像素总部·薄荷",
    "night-shift": "夜班模式",
};

const budgetModeLabels: Record<string, string> = {
    guardrailed: "稳健受控",
    balanced: "均衡模式",
    "throughput-first": "吞吐优先",
};

const timelineKindLabels: Record<string, string> = {
    run: "运行",
    intake: "接入",
    planning: "规划",
    decomposition: "拆解",
    staffing: "排班",
    execution: "执行",
    review: "审核",
    handoff: "交接",
    security: "安全",
    audit: "审计",
    delivery: "交付",
    archive: "归档",
    artifact: "产物",
};

const stageHeadlineLabels: Record<string, string> = {
    intake: "需求接入与流程选择",
    planning: "原子拆解与角色排布",
    execution: "并行执行阶段",
    review: "审核复核与治理收口",
    delivery: "交付打包与归档",
    "Intake and workflow selection": "需求接入与流程选择",
    "Atomic decomposition and staffing": "原子拆解与角色排布",
    "Parallel implementation": "并行执行阶段",
    "Peer review and governance packaging": "审核复核与治理收口",
    "Security and approval gate": "安全审批与治理门禁",
    "Delivery packaging and archive": "交付打包与归档",
    "Peer review": "同行评审",
    "Security approval hold": "安全审批挂起",
    "Delivery packaging": "交付打包",
};

const highlightLabels: Record<string, string> = {
    "Owns the workflow choice": "负责最终流程选型",
    "Reserves review capacity": "为审核环节预留产能",
    "Keeps enterprise gates visible": "保持企业门禁全程可见",
    "Maintains review reserve lanes": "持续保留审核通道",
    "Publishes stage transitions": "发布阶段切换信号",
    "Treats approvals as hard gates": "把审批视作硬门禁",
    "Normalizes the mission into structured constraints": "把任务整理成结构化约束",
    "Captures concurrency and review requirements": "记录并发上限与审核要求",
    "Feeds the planning lane": "把标准化需求送入规划区",
    "Captured the 100-agent ceiling": "记录了 100 个智能体的并发上限",
    "Flagged audit coverage": "标记了审计覆盖要求",
    "Fed a structured scope brief into planning": "已将结构化范围简报送入规划区",
    "Maps the mission into staged workstreams": "把任务映射为阶段化工作流",
    "Keeps review lanes explicit": "明确标注评审链路",
    "Prepares the staffing outline": "准备角色编制方案",
    "Splits work into atomic tasks": "把工作拆到原子任务级别",
    "Assigns one owner per output": "为每个产物指定唯一负责人",
    "Marks review boundaries": "标记审核边界",
    "Expanded work into atomic nodes": "已把工作扩展成原子节点",
    "Inserted review gates": "已插入评审门禁",
    "Marked one owner per task": "每个任务都绑定唯一负责人",
    "Checks task atomicity and integration seams": "检查任务原子性与集成接缝",
    "Verifies reviewer visibility": "确保审核视图保持可见",
    "Packages findings for governance lanes": "将发现的问题送入治理链路",
    "Verifies tasks are atomic": "核验任务已经足够原子化",
    "Checks review hooks remain visible": "检查评审挂钩依然可见",
    "Packages findings for security": "把结论打包送往安全审核",
    "Aggregates final outputs": "汇总最终交付产物",
    "Waits for every hard gate to close": "等待全部硬门禁关闭",
    "Packages the release handoff": "打包最终交付交接",
    "Prepares the release summary": "准备最终发布摘要",
    "Captures approvals and lineage": "记录审批与血缘信息",
    "Emits replayable stage history": "输出可回放的阶段历史",
    "Seals the delivery ledger": "封存交付台账",
    "Runs acceptance against persisted workflow": "基于持久化流程执行验收",
    "Checks operator controls": "检查操作台控制项",
    "Feeds validated bundles into delivery": "把验收通过的包送入交付区",
    "Reviews risky outbound actions": "审核高风险外发动作",
    "Treats policy holds as hard stops": "把策略挂起视作硬阻塞",
    "Waits for explicit clearance": "等待明确放行",
    "Blocks risky actions instead of warning": "对高风险动作直接阻断",
    "Waits on explicit human approval": "等待人工明确审批",
    "Writes the hold state into the floor": "把挂起状态写入运行现场",
    "Turns blocked actions into approval tickets": "把阻塞动作转成人工审批单",
    "Waits on explicit human decision": "等待人工明确决策",
    "Releases the run once approval lands": "审批通过后释放运行",
    "Owns one or more atomic execution tasks": "负责一个或多个原子执行任务",
    "Packages artifacts for review": "把产物打包交给审核区",
    "Works inside the execution bay": "在执行区持续推进工作",
    "Replaced the starter UI": "已替换初始界面骨架",
    "Added control desk and task board": "已补上控制台与任务板",
    "Hands artifacts into review": "正在把产物送入审核链路",
    "Stores run, task, gate, worker, and provider state": "持久化运行、任务、门禁、智能体与接口状态",
    "Supports reset and settings persistence": "支持重置与设置持久化",
    "Packages runtime output for review": "把运行时产物打包送审",
};

const textTranslations: Record<string, string> = {
    "A persisted enterprise workflow where planning, execution, review, approvals, and AI routing all share the same office floor.":
        "这是一个可持久化的企业级工作流运行空间，规划、执行、审核、审批与接口路由都在同一间办公室里协作。",
    "Enterprise delivery command room": "企业交付指挥台",
    "Enterprise Task Delivery v1": "企业任务交付流程",
    "Engineering Enterprise Workflow": "工程研发流程",
    "maliang swarm foundation": "maliang swarm 基础工程",
    "OpenAI-Compatible Gateway": "兼容接口",
    "OpenAI-compatible gateway": "兼容接口",
    "OpenAI-Compatible Gateway (OpenAI 兼容，密钥待配置)": "兼容接口（密钥待配置）",
    "Not configured": "未配置",
    "Intake, orchestration, and workflow selection": "需求接入与流程选择",
    "Decomposition and staffing": "任务拆解与编排",
    "Peer review and security": "同行评审与安全治理",
    "Parallel implementation and integration": "并行执行与集成",
    "Release packaging and audit": "发布打包与审计",
    "Decompose until every task has one owner, one artifact, and an isolated retry path.":
        "必须拆解到每个任务都只有一个负责人、一个产物和一个独立重试边界。",
    "Four-eyes review by default, human approval for high-risk actions, and persistent audit on every stage.":
        "默认双人复核，高风险动作必须人工审批，并在每个阶段保留持久化审计轨迹。",
    "Four-eyes review by default, human approval for high-risk actions, and persistent audit at every stage.":
        "默认双人复核，高风险动作必须人工审批，并在每个阶段保留持久化审计轨迹。",
    "Normalize the user request": "整理用户任务输入",
    "Capture concurrency ceiling, audit requirements, and approval coverage.": "提取并发上限、审计要求和审批覆盖范围。",
    "Select the enterprise workflow template": "选择合适的企业流程模板",
    "Choose the standard enterprise flow before execution begins.": "在执行开始前，先选择标准企业流程路径。",
    "Decompose work into atomic tasks": "拆解为原子任务",
    "Split work until every node has one owner and one retry boundary.": "持续拆解，直到每个节点都有唯一负责人和唯一重试边界。",
    "Build the office UI shell": "构建办公室可视化外壳",
    "Lay out the office floor, task board, inspector, and controls.": "搭建办公室场景、任务板、抽屉和控制台。",
    "Persist runtime state in SQLite": "将运行时状态写入本地存储",
    "Store run, task, gate, and worker state locally.": "把运行、任务、门禁与智能体状态持久化到本地。",
    "Review, secure, and package delivery": "完成评审、安全核验与交付打包",
    "Peer review opens first, then security hold, then delivery packaging.": "先进行同行评审，再处理安全挂起，最后进入交付打包。",
    "Reserve review capacity": "预留审核产能",
    "Watch worker pools and gate load": "监控执行池与门禁负载",
    "Watch peer review progress": "观察同行评审进度",
    "Hold the run on security approval": "在安全审批前挂起运行",
    "Prepare final delivery sign-off": "准备最终交付签收",
    "Controls the enterprise path and prevents execution from starving review lanes.":
        "负责控制整体企业流程，避免执行阶段挤占审核产能。",
    "Normalize goals, constraints, and delivery signals": "标准化目标、约束与交付信号",
    "Package the normalized brief for planning": "将标准化简报送往规划区",
    "Intake work complete": "需求接入工作已完成",
    "Turns freeform intent into a structured brief the planning lane can decompose.":
        "把自由描述转成规划区可继续拆解的结构化简报。",
    "Waiting for the planning blueprint": "等待规划蓝图",
    "Split work into atomic tasks": "将工作拆为原子任务",
    "Deliver the task graph to execution": "把任务图送入执行区",
    "Task graph complete": "任务图已完成",
    "Refuses compound tasks until each node has one owner, one artifact, and one retry path.":
        "拒绝复合任务，直到每个节点都只有一个负责人、一个产物和一个重试路径。",
    "Waiting for decomposition output": "等待拆解结果",
    "Queued for the UI slice": "界面工作流已进入队列",
    "Build the office shell and controls": "构建办公室场景与控制台",
    "Package the UI slice for review": "将界面产物打包送审",
    "UI slice approved": "界面产物已通过审核",
    "Owns the office floor, task board, and the visible operator shell.":
        "负责办公室现场、任务板以及操作者可见的主界面外壳。",
    "Queued for the runtime slice": "运行时工作流已进入队列",
    "Persist run, task, gate, worker, and provider state": "持久化运行、任务、门禁、智能体与接口状态",
    "Package the runtime slice for review": "将运行时产物打包送审",
    "Runtime slice approved": "运行时产物已通过审核",
    "Owns the local state store that makes the workflow replayable.": "负责让整套流程可回放的本地状态存储。",
    "Waiting for reviewable artifacts": "等待可审核产物",
    "Waiting for execution bundles": "等待执行产物包",
    "Check atomic-task fidelity and UI coverage": "检查原子任务质量与界面覆盖",
    "Review notes forwarded to security": "评审结论已送往安全审核",
    "Review work complete": "评审工作已完成",
    "Looks for brittle assumptions, missing review hooks, and unclear ownership before delivery opens.":
        "在交付开启前，重点检查脆弱假设、缺失评审挂钩与责任边界不清的问题。",
    "Waiting for risky actions to review": "等待高风险动作进入审核",
    "Pause risky outbound actions until approval arrives": "在审批到来前暂停高风险外发动作",
    "Record approval and close the security gate": "记录审批结果并关闭安全门禁",
    "Treats risky outbound actions as hard stops until a human explicitly clears them.":
        "对高风险外发动作采取硬阻断，直到人工明确放行。",
    "Waiting for upstream gates to close": "等待上游门禁关闭",
    "Waiting for security approval": "等待安全审批",
    "Assemble the final delivery summary": "组装最终交付摘要",
    "Cannot package delivery until review and security both report green.": "只有评审与安全都放行后，才会开始打包交付。",
    "Waiting for upstream planning signals.": "等待上游规划信号。",
    "Waiting for upstream artifacts to enter the review lane.": "等待上游产物进入审核链路。",
    "Run created in the command deck": "运行已在指挥台创建",
    "The system opened a persisted enterprise run and started intake normalization.":
        "系统已创建持久化运行实例，并开始执行需求标准化。",
    "Chief Orchestrator started workflow selection": "主脑开始选择工作流模板",
    "The control deck is choosing the standard enterprise path before dispatching workers.":
        "主控台正在选择标准企业流程路径，然后才会派发执行智能体。",
    "Scope brief sent to decomposition": "范围简报已送入拆解区",
    "Intake delivered the normalized brief into the planning room.": "需求接入阶段已把标准化简报送入规划区。",
    "Atomic task expansion started": "原子任务扩展已开始",
    "The decomposer is splitting the run until every node has one owner and one retry boundary.":
        "任务拆解师正在持续拆分，直到每个节点都有唯一负责人和唯一重试边界。",
    "Execution lanes reserved": "执行工位已预留",
    "Worker slots were allocated without starving review capacity.": "系统已分配执行工位，同时保留了足够的审核产能。",
    "UI and runtime slices entered the execution bay": "界面与运行时工作流已进入执行区",
    "The system is now writing the office shell and the SQLite runtime state.": "系统正在同步开发办公室界面与本地运行时。",
    "Execution bundle opened in peer review": "执行产物包已进入同行评审",
    "The review corridor started checking task atomicity, UI hooks, and runtime seams.": "审核区开始检查任务原子性、界面挂钩与运行时接缝。",
    "Review packet prepared for security": "评审包已准备送往安全审核",
    "Peer review bundled findings and flagged one outbound path for security attention.":
        "同行评审已打包发现的问题，并标记了一条需要安全关注的外发路径。",
    "Security gate blocked a risky action": "安全门禁阻断了高风险动作",
    "A human approval is required before the run can continue to delivery.": "必须获得人工审批后，运行才能继续进入交付阶段。",
    "Approval hold written to the runtime ledger": "审批挂起已写入运行时台账",
    "The runtime now persists this hold as a replayable event.": "运行时已将这次挂起记录为可回放事件。",
    "Security approval recorded": "安全审批已记录",
    "The blocked action was cleared and the gate was closed.": "被阻断的动作已放行，对应门禁也已关闭。",
    "Delivery package assembly started": "交付包组装已开始",
    "The delivery manager began aggregating the release summary and audit trail.": "交付经理开始汇总发布摘要与审计轨迹。",
    "Intake normalization": "需求标准化",
    "Planning review": "规划评审",
    "Staffing and task coverage check": "编制与任务覆盖检查",
    "DAG staffing check": "编制与任务覆盖检查",
    "Peer review on execution bundle": "执行产物同行评审",
    "QA acceptance": "QA 验收",
    "Security review": "安全审核",
    "Security approval on outbound action": "高风险外发动作安全审核",
    "Human approval": "人工审批",
    "Final delivery sign-off": "最终交付签收",
};

const phraseTranslations: Array<[string, string]> = [
    ["Enterprise delivery command room", "企业交付指挥台"],
    ["Engineering control room", "工程指挥台"],
    ["Command Deck", "指挥台"],
    ["Planning Room", "规划室"],
    ["Review Corridor", "审核走廊"],
    ["Execution Bay", "执行区"],
    ["Delivery Desk", "交付台"],
    ["OpenAI-Compatible Gateway", "兼容接口"],
    ["OpenAI-compatible gateway", "兼容接口"],
    ["OpenAI compatible, key pending", "密钥待配置"],
    ["key pending", "密钥待配置"],
    ["Engineering Enterprise Workflow", "工程研发流程"],
    ["Enterprise Task Delivery v1", "企业任务交付流程"],
    ["maliang swarm foundation", "maliang swarm 基础工程"],
];

const exactTranslations = new Map<string, string>(Object.entries(textTranslations));
const insensitiveTranslations = new Map<string, string>(
    Object.entries(textTranslations).map(([key, value]) => [key.trim().toLowerCase(), value]),
);

export function normalizeStageKey(value: string): string {
    const normalized = value.trim().toLowerCase();

    if (normalized in stageLabels) {
        return normalized;
    }
    if (normalized.includes("intake") || normalized.includes("接入")) {
        return "intake";
    }
    if (normalized.includes("planning") || normalized.includes("decomposition") || normalized.includes("拆解") || normalized.includes("规划")) {
        return "planning";
    }
    if (normalized.includes("execution") || normalized.includes("implementation") || normalized.includes("执行")) {
        return "execution";
    }
    if (normalized.includes("review") || normalized.includes("approval") || normalized.includes("审核") || normalized.includes("评审") || normalized.includes("审批")) {
        return "review";
    }
    if (normalized.includes("delivery") || normalized.includes("archive") || normalized.includes("交付") || normalized.includes("归档")) {
        return "delivery";
    }

    return value;
}

function translatePhrases(value: string): string {
    return phraseTranslations.reduce((current, [source, target]) => current.split(source).join(target), value);
}

export function localizeText(value: string): string {
    const normalized = value.trim();
    if (!normalized) {
        return "";
    }

    const exact = exactTranslations.get(normalized);
    if (exact) {
        return exact;
    }

    const lowered = insensitiveTranslations.get(normalized.toLowerCase());
    if (lowered) {
        return lowered;
    }

    return translatePhrases(normalized);
}

export function formatStage(value: string): string {
    const key = normalizeStageKey(value);
    return stageLabels[key] || localizeText(value);
}

export function formatStatus(value: string): string {
    return statusLabels[value] || localizeText(value);
}

export function formatRisk(value: string): string {
    return riskLabels[value] || localizeText(value);
}

export function formatRole(value: string): string {
    return roleLabels[value] || localizeText(value);
}

export function formatTeam(value: string): string {
    return teamLabels[value] || localizeText(value);
}

export function formatModelTier(value: string): string {
    return tierLabels[value] || localizeText(value);
}

export function formatComputerState(value: string): string {
    return computerStateLabels[value] || localizeText(value);
}

export function formatProviderFormat(value: string): string {
    return providerFormatLabels[value] || localizeText(value);
}

export function formatTheme(value: string): string {
    return themeLabels[value] || localizeText(value);
}

export function formatBudgetMode(value: string): string {
    return budgetModeLabels[value] || localizeText(value);
}

export function formatTimelineKind(value: string): string {
    return timelineKindLabels[value] || localizeText(value);
}

export function formatStageHeadline(value: string): string {
    return stageHeadlineLabels[value] || stageHeadlineLabels[normalizeStageKey(value)] || localizeText(value);
}

export function formatPlannerSource(value: string): string {
    const trimmed = value.trim();

    if (trimmed.startsWith("live via ")) {
        return `实时规划 / ${trimmed.slice("live via ".length)}`;
    }
    if (trimmed === "rules fallback") {
        return "规则回退";
    }
    if (trimmed === "rules fallback / no live planner route") {
        return "规则回退 / 未找到可用主脑路由";
    }
    if (trimmed.startsWith("rules fallback / ")) {
        return `规则回退 / ${localizeText(trimmed.slice("rules fallback / ".length))}`;
    }

    return localizeText(trimmed);
}

export function formatGateLabel(value: string): string {
    const match = value.match(/^(.*)\s+\(([^()]+)\)$/);
    if (!match) {
        return localizeText(value);
    }

    return `${localizeText(match[1])}（${formatStatus(match[2])}）`;
}

export function formatHighlight(value: string): string {
    return highlightLabels[value] || localizeText(value);
}

export function formatAgentLabel(value: string): string {
    if (value === "Archive") {
        return "归档区";
    }

    const [name, role] = value.split(" / ");
    if (!role) {
        return localizeText(value);
    }

    return `${name} / ${formatRole(role)}`;
}
