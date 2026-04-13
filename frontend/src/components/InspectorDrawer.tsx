import {useEffect, useMemo, useState} from "react";
import type {AgentState} from "../types";
import {
    formatAgentLabel,
    formatComputerState,
    formatHighlight,
    formatModelTier,
    formatRisk,
    formatRole,
    formatStatus,
    formatTeam,
    localizeText,
} from "../lib/format";

interface InspectorDrawerProps {
    agent?: AgentState;
    variant?: "sidebar" | "modal";
}

function statusTone(status: string): "active" | "review" | "warning" | "danger" | "success" | "idle" {
    switch (status) {
        case "in_progress":
        case "typing":
        case "handoff":
        case "in_transit":
            return "active";
        case "reviewing":
        case "thinking":
            return "review";
        case "queued":
        case "waiting_approval":
            return "warning";
        case "blocked":
            return "danger";
        case "completed":
        case "done":
            return "success";
        default:
            return "idle";
    }
}

function statusNarrative(agent: AgentState): string {
    switch (agent.status) {
        case "in_progress":
        case "typing":
            return "正在处理当前任务，产物会在完成后进入交接。";
        case "reviewing":
            return "正在复核产物质量和流程边界。";
        case "thinking":
        case "queued":
            return "等待上游规划或任务分配。";
        case "handoff":
        case "in_transit":
            return "资料正在向下一位协作者交接。";
        case "waiting_approval":
            return "当前动作需要审批放行。";
        case "blocked":
            return "遇到阻塞项，等待治理门禁关闭。";
        case "completed":
        case "done":
            return "当前职责已完成，等待归档或下一次调度。";
        default:
            return "当前待命，等待主脑分配工作。";
    }
}

function initialsForName(name: string): string {
    const normalized = name.trim();
    if (!normalized) {
        return "MS";
    }
    const compact = normalized.replace(/\s+/g, "");
    const first = compact[0] || "M";
    const second = compact[1] || "";
    if (!/[A-Za-z]/.test(first)) {
        return compact.slice(0, 2);
    }
    return `${first}${second}`.toUpperCase();
}

export function InspectorDrawer({agent, variant = "sidebar"}: InspectorDrawerProps) {
    const [isPromptVisible, setPromptVisible] = useState(false);
    const promptText = useMemo(() => (agent ? buildAgentPrompt(agent) : ""), [agent]);

    useEffect(() => {
        setPromptVisible(false);
    }, [agent?.id]);

    if (!agent) {
        return (
            <div className={`inspector-drawer inspector-drawer-modern ${variant === "modal" ? "inspector-drawer-modal" : ""}`}>
                <div className="empty-state-card slim inspector-empty-card">
                    <p className="section-kicker">智能体详情</p>
                    <h2 className="page-section-title">尚未选中智能体</h2>
                    <p className="page-section-copy">在办公室里点击一个智能体，这里会展示它的职责、进度和交接信息。</p>
                </div>
            </div>
        );
    }

    const tone = statusTone(agent.status);
    const artifactLabel = agent.artifact?.trim() ? localizeText(agent.artifact) : "暂无资料";
    const nextTargetLabel = agent.nextTarget?.trim() ? formatAgentLabel(agent.nextTarget) : "暂无下一站";
    const lastAction = agent.lastAction?.trim() ? localizeText(agent.lastAction) : localizeText(agent.currentTask);
    const lastOutput = agent.lastOutput?.trim() ? localizeText(agent.lastOutput) : "暂无返回内容";
    const lastHandoff = agent.lastHandoff?.trim() ? localizeText(agent.lastHandoff) : artifactLabel;
    const lastReceiver = agent.lastReceiver?.trim() ? formatAgentLabel(agent.lastReceiver) : nextTargetLabel;
    const lastArtifactPath = agent.lastArtifactPath?.trim() ? agent.lastArtifactPath : "未生成本地文件";
    const highlightList = agent.highlights.length > 0
        ? agent.highlights
        : ["当前没有额外关注点，系统会在状态变化后自动补充。"];

    return (
        <div className={`inspector-drawer inspector-drawer-modern ${variant === "modal" ? "inspector-drawer-modal" : ""}`}>
            <section className={`inspector-hero tone-${tone}`}>
                <div className="inspector-hero-top">
                    <div className="inspector-avatar" aria-hidden="true">{initialsForName(agent.name)}</div>
                    <div className="inspector-hero-copy">
                        <p className="section-kicker">智能体详情</p>
                        <h2 className="page-section-title">{agent.name}</h2>
                        <p className="page-section-copy">{formatRole(agent.role)}，当前位于 {formatTeam(agent.team)}。</p>
                    </div>
                </div>

                <div className="inspector-chip-row">
                    <span className={`inspector-chip tone-${tone}`}>{formatStatus(agent.status)}</span>
                    <span className="inspector-chip neutral">{agent.deskLabel}</span>
                    <span className="inspector-chip neutral">{formatModelTier(agent.modelTier)}</span>
                    <span className={`inspector-chip risk-${agent.riskLevel}`}>{formatRisk(agent.riskLevel)}</span>
                </div>
            </section>

            <article className="surface-subcard inspector-progress-card">
                <div className="inspector-progress-head">
                    <div>
                        <p className="section-kicker">执行进度</p>
                        <h3 className="surface-title">{agent.progress}%</h3>
                    </div>
                    <span className={`inspector-inline-tag tone-${tone}`}>{statusNarrative(agent)}</span>
                </div>

                <div className="progress-track">
                    <div className="progress-fill" style={{width: `${agent.progress}%`}}/>
                </div>

                <div className="inspector-progress-meta">
                    <div>
                        <span>队列深度</span>
                        <strong>{agent.queueDepth}</strong>
                    </div>
                    <div>
                        <span>最近更新</span>
                        <strong>{agent.lastUpdate}</strong>
                    </div>
                </div>
            </article>

            <section className="inspector-signal-grid">
                <article className="surface-subcard inspector-signal-card">
                    <span className="inspector-signal-label">电脑状态</span>
                    <strong>{formatComputerState(agent.computerState)}</strong>
                    <p>{statusNarrative(agent)}</p>
                </article>

                <article className="surface-subcard inspector-signal-card">
                    <span className="inspector-signal-label">当前职责</span>
                    <strong>{localizeText(agent.currentTask)}</strong>
                    <p>{localizeText(agent.detail)}</p>
                </article>
            </section>

            <article className="surface-subcard">
                <div className="panel-header">
                    <div>
                        <p className="section-kicker">资料流向</p>
                        <h3 className="surface-title">当前交接面板</h3>
                    </div>
                    <span className={`inspector-inline-tag tone-${tone}`}>{formatTeam(agent.team)}</span>
                </div>

                <div className="detail-list compact">
                    <div className="detail-row">
                        <span>当前资料</span>
                        <strong>{artifactLabel}</strong>
                    </div>
                    <div className="detail-row multi">
                        <span>下一目标</span>
                        <strong>{nextTargetLabel}</strong>
                    </div>
                    <div className="detail-row">
                        <span>风险级别</span>
                        <strong>{formatRisk(agent.riskLevel)}</strong>
                    </div>
                </div>
            </article>

            <article className="surface-subcard inspector-trace-card">
                <div className="panel-header">
                    <div>
                        <p className="section-kicker">执行追踪</p>
                        <h3 className="surface-title">最近一次协作明细</h3>
                    </div>
                </div>

                <div className="inspector-trace-grid">
                    <div className="inspector-trace-item">
                        <span>做了什么</span>
                        <p>{lastAction}</p>
                    </div>
                    <div className="inspector-trace-item">
                        <span>返回内容</span>
                        <p>{lastOutput}</p>
                    </div>
                    <div className="inspector-trace-item">
                        <span>向下传递</span>
                        <p>{lastHandoff}</p>
                    </div>
                    <div className="inspector-trace-item">
                        <span>接收对象</span>
                        <p>{lastReceiver}</p>
                    </div>
                    <div className="inspector-trace-item inspector-trace-path">
                        <span>本地文档路径</span>
                        <p>{lastArtifactPath}</p>
                    </div>
                </div>
            </article>

            <article className="surface-subcard">
                <p className="section-kicker">当前关注点</p>
                <div className="highlight-list modern">
                    {highlightList.map((highlight) => (
                        <div key={highlight} className="highlight-item">
                            {formatHighlight(highlight)}
                        </div>
                    ))}
                </div>
            </article>

            <article className="surface-subcard inspector-prompt-card">
                <div className="panel-header inspector-prompt-head">
                    <div>
                        <p className="section-kicker">运行提示词</p>
                        <h3 className="surface-title">当前智能体指令</h3>
                        <p className="page-section-copy">默认折叠，需要时展开查看。</p>
                    </div>
                    <button
                        className="ghost-button"
                        onClick={() => setPromptVisible((current) => !current)}
                        type="button"
                    >
                        {isPromptVisible ? "隐藏提示词" : "查看提示词"}
                    </button>
                </div>

                {isPromptVisible ? (
                    <pre className="inspector-prompt-block">{promptText}</pre>
                ) : (
                    <div className="inspector-prompt-placeholder">
                        <span className="soft-pill">按需展开</span>
                        <span className="soft-pill">不常驻占用空间</span>
                    </div>
                )}
            </article>
        </div>
    );
}

function buildAgentPrompt(agent: AgentState): string {
    const highlightLines = agent.highlights.length > 0
        ? agent.highlights.map((highlight) => `- ${formatHighlight(highlight)}`).join("\n")
        : "- 当前没有额外关注点，保持标准执行。";

    const artifactLine = agent.artifact?.trim() ? `当前资料：${localizeText(agent.artifact)}` : "当前资料：暂无";
    const nextTargetLine = agent.nextTarget?.trim() ? `下一目标：${formatAgentLabel(agent.nextTarget)}` : "下一目标：暂无";
    const lastActionLine = agent.lastAction?.trim() ? `最近动作：${localizeText(agent.lastAction)}` : "";
    const lastOutputLine = agent.lastOutput?.trim() ? `最近返回：${localizeText(agent.lastOutput)}` : "";
    const lastHandoffLine = agent.lastHandoff?.trim() ? `最近交接：${localizeText(agent.lastHandoff)}` : "";
    const lastReceiverLine = agent.lastReceiver?.trim() ? `交接对象：${formatAgentLabel(agent.lastReceiver)}` : "";

    return [
        `你是 maliang swarm 办公室中的 ${formatRole(agent.role)}，位于 ${formatTeam(agent.team)}，工位编号为 ${agent.deskLabel}。`,
        `当前模型档位：${formatModelTier(agent.modelTier)}。当前状态：${formatStatus(agent.status)}。电脑状态：${formatComputerState(agent.computerState)}。`,
        `你的当前任务是：${localizeText(agent.currentTask)}。`,
        `职责说明：${localizeText(agent.detail)}。`,
        artifactLine,
        nextTargetLine,
        lastActionLine,
        lastOutputLine,
        lastHandoffLine,
        lastReceiverLine,
        "执行原则：",
        highlightLines,
        "输出要求：只返回与当前职责直接相关的中文结果，内容要简洁、可执行、可审计，并在需要交接时明确指出下一位协作者与产物名称。",
    ].filter(Boolean).join("\n");
}
