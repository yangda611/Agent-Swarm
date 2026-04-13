import {useMemo} from "react";
import type {AgentState, HandoffState, SceneZone} from "../types";
import {formatGateLabel, formatStatus, localizeText, normalizeStageKey} from "../lib/format";

interface OfficeSceneProps {
    zones: SceneZone[];
    agents: AgentState[];
    handoffs: HandoffState[];
    completedGates: string[];
    focusZoneId: string;
    pendingGates: string[];
    selectedAgentId?: string;
    stageMotionKey: number;
    stageTitle: string;
    viewMode: "desks" | "routes" | "events";
    onSelectAgent: (agentId: string) => void;
}

const zoneOrder = ["command", "planning", "engineering", "review", "delivery"];
const zoneByTeam: Record<string, string> = {
    Command: "command",
    Planning: "planning",
    Execution: "engineering",
    Review: "review",
    Delivery: "delivery",
};

const zoneIcons: Record<string, string> = {
    command: "指",
    planning: "策",
    engineering: "执",
    review: "审",
    delivery: "交",
};

const zoneNarratives: Record<string, string> = {
    command: "主脑确认范围、优先级与门禁。",
    planning: "把任务拆到原子级并排布角色。",
    engineering: "并发推进实现与集成工作。",
    review: "复核质量、安全与放行条件。",
    delivery: "汇总产物、签收并归档。",
};

interface StageEvent {
    id: string;
    zoneId: string;
    tone: "alert" | "success" | "archive";
    label: string;
    title: string;
    detail: string;
}

function initialsForName(name: string): string {
    const compact = name.trim().replace(/\s+/g, "");
    if (!compact) {
        return "MS";
    }

    const first = compact[0];
    const second = compact[1] || "";

    if (!/[A-Za-z]/.test(first)) {
        return compact.slice(0, 2);
    }

    return `${first}${second}`.toUpperCase();
}

function transitState(handoffs: HandoffState[], agentId: string): "dispatching" | "receiving" | null {
    const active = handoffs.find((handoff) => handoff.status !== "completed" && handoff.status !== "done"
        && (handoff.fromAgentId === agentId || handoff.toAgentId === agentId));
    if (!active) {
        return null;
    }
    if (active.fromAgentId === agentId) {
        return "dispatching";
    }
    return "receiving";
}

function buildStageEvents(
    focusZoneId: string,
    agents: AgentState[],
    handoffs: HandoffState[],
    pendingGates: string[],
    completedGates: string[],
): StageEvent[] {
    const events: StageEvent[] = [];
    const waitingApproval = agents.some((agent) => agent.status === "blocked" || agent.status === "waiting_approval");
    const reviewBusy = pendingGates.some((gate) => gate.includes("评审") || gate.includes("QA") || gate.includes("审核") || gate.includes("审批"));
    const deliveryRunning = handoffs.some((handoff) => handoff.id.includes("delivery") || handoff.artifactName.includes("delivery"));

    if (focusZoneId === "review" && waitingApproval && reviewBusy) {
        events.push({
            id: "review-alert",
            zoneId: "review",
            tone: "alert",
            label: "审核告警",
            title: "治理门禁仍在阻塞",
            detail: "审核区仍有待放行动作，当前不会越过人工或安全门禁。",
        });
    }

    if (focusZoneId === "delivery" && completedGates.some((gate) => gate.includes("人工审批") || gate.includes("安全审核"))) {
        events.push({
            id: "delivery-clear",
            zoneId: "delivery",
            tone: "success",
            label: "审批放行",
            title: "交付通道已打开",
            detail: "关键门禁已关闭，交付区正在收口产物与签收记录。",
        });
    }

    if (focusZoneId === "delivery" && deliveryRunning) {
        events.push({
            id: "delivery-archive",
            zoneId: "delivery",
            tone: "archive",
            label: "归档中",
            title: "交付包正在封存",
            detail: "最终交付摘要与审计记录正在同步归档。",
        });
    }

    return events;
}

export function OfficeScene({
    zones,
    agents,
    handoffs,
    completedGates,
    focusZoneId,
    pendingGates,
    selectedAgentId,
    stageMotionKey,
    stageTitle,
    viewMode,
    onSelectAgent,
}: OfficeSceneProps) {
    const orderedZones = [...zones].sort((left, right) => zoneOrder.indexOf(left.id) - zoneOrder.indexOf(right.id));
    const stageEvents = buildStageEvents(focusZoneId, agents, handoffs, pendingGates, completedGates);
    const focusZone = orderedZones.find((zone) => zone.id === focusZoneId) || orderedZones[0];
    const agentsByZone = orderedZones.reduce<Record<string, AgentState[]>>((collection, zone) => {
        collection[zone.id] = agents
            .filter((agent) => zoneByTeam[agent.team] === zone.id)
            .sort((left, right) => left.deskLabel.localeCompare(right.deskLabel));
        return collection;
    }, {});

    return (
        <div className="office-scene office-scene-reboot">
            <section key={`stage-${stageMotionKey}-${focusZoneId}`} className="office-stage-board">
                <div>
                    <p className="section-kicker">当前镜头</p>
                    <h2>{localizeText(focusZone?.name || "指挥台")}</h2>
                    <p className="stage-focus-copy">{stageTitle}</p>
                </div>
                <div className="office-stage-meta">
                    <span className="stage-focus-pill">{zoneNarratives[focusZoneId] || zoneNarratives.command}</span>
                    <span className="stage-focus-pill muted">{`活跃智能体 ${agents.length}`}</span>
                    <span className="stage-focus-pill muted">{`交接 ${handoffs.length}`}</span>
                </div>
            </section>

            {viewMode === "desks" ? (
                <section className={`office-theater focus-${focusZoneId}`}>
                    <div className="office-theater-grid">
                        {orderedZones.map((zone) => (
                            <article
                                key={zone.id}
                                className={`office-room-card room-${zone.id} ${zone.id === focusZoneId ? "focused" : ""}`}
                            >
                                <div className="office-room-head">
                                    <div className="office-room-title">
                                        <span className="office-room-icon">{zoneIcons[zone.id] || "区"}</span>
                                        <div>
                                            <strong>{localizeText(zone.name)}</strong>
                                            <span>{zoneNarratives[zone.id] || localizeText(zone.purpose)}</span>
                                        </div>
                                    </div>
                                    <span className="office-room-count">{`${agentsByZone[zone.id]?.length || 0} 席`}</span>
                                </div>

                                <div className="office-room-floor">
                                    {(agentsByZone[zone.id] || []).length === 0 ? (
                                        <div className="office-room-empty">当前无人值守</div>
                                    ) : (
                                        <div className="office-pawn-grid">
                                            {(agentsByZone[zone.id] || []).map((agent) => {
                                                const transferState = transitState(handoffs, agent.id);
                                                const normalizedStage = normalizeStageKey(agent.status);

                                                return (
                                                    <button
                                                        key={agent.id}
                                                        className={`office-agent-pawn status-${agent.status} ${selectedAgentId === agent.id ? "active" : ""} ${transferState || ""}`}
                                                        onClick={() => onSelectAgent(agent.id)}
                                                        type="button"
                                                    >
                                                        <span className={`office-agent-status status-${normalizedStage}`}/>
                                                        <span className="office-agent-avatar">{initialsForName(agent.name)}</span>
                                                        <span className="office-agent-name">{agent.name}</span>
                                                        <span className="office-agent-desk">{agent.deskLabel}</span>
                                                        <span className="office-agent-progress">
                                                            <span style={{width: `${agent.progress}%`}}/>
                                                        </span>
                                                    </button>
                                                );
                                            })}
                                        </div>
                                    )}
                                </div>
                            </article>
                        ))}
                    </div>

                    <div className="office-theater-footer">
                        {handoffs.length === 0 ? (
                            <span className="soft-pill">当前没有正在交接的资料</span>
                        ) : (
                            handoffs.map((handoff) => (
                                <span key={handoff.id} className="office-transfer-pill">
                                    {handoff.artifactName}
                                </span>
                            ))
                        )}
                    </div>
                </section>
            ) : null}

            {viewMode === "routes" ? (
                <section className="office-route-board">
                    <div className="office-route-columns">
                        {handoffs.length === 0 ? (
                            <div className="empty-state-card slim">
                                <strong>当前没有资料交接</strong>
                                <p>任务推进后，这里会显示跨区流转路线。</p>
                            </div>
                        ) : (
                            handoffs.map((handoff) => {
                                const fromAgent = agents.find((agent) => agent.id === handoff.fromAgentId);
                                const toAgent = agents.find((agent) => agent.id === handoff.toAgentId);

                                if (!fromAgent || !toAgent) {
                                    return null;
                                }

                                return (
                                    <article key={handoff.id} className="office-route-card">
                                        <div className="office-route-rail"/>
                                        <div className="office-route-stops">
                                            <span>{fromAgent.name}</span>
                                            <strong>{handoff.artifactName}</strong>
                                            <span>{toAgent.name}</span>
                                        </div>
                                        <div className="office-route-foot">
                                            <span>{fromAgent.deskLabel}</span>
                                            <span>{formatStatus(handoff.status)}</span>
                                            <span>{toAgent.deskLabel}</span>
                                        </div>
                                    </article>
                                );
                            })
                        )}
                    </div>
                </section>
            ) : null}

            {viewMode === "events" ? (
                <section className="office-events-board">
                    <article className="surface-subcard">
                        <p className="section-kicker">现场事件</p>
                        {stageEvents.length === 0 ? (
                            <div className="empty-state-card slim">
                                <strong>当前没有特殊事件</strong>
                                <p>当门禁阻塞、审批放行或归档开始时，这里会出现事件卡。</p>
                            </div>
                        ) : (
                            <div className="office-event-compact-list">
                                {stageEvents.map((event) => (
                                    <article key={event.id} className={`office-event-compact ${event.tone}`}>
                                        <span>{event.label}</span>
                                        <strong>{event.title}</strong>
                                        <p>{event.detail}</p>
                                    </article>
                                ))}
                            </div>
                        )}
                    </article>

                    <article className="surface-subcard">
                        <p className="section-kicker">治理门禁</p>
                        <div className="office-gate-grid">
                            <div>
                                <span className="analytics-mini-label">待关闭</span>
                                <div className="gate-list compact">
                                    {pendingGates.length > 0 ? pendingGates.map((gate) => (
                                        <span key={gate} className="gate-chip pending">{formatGateLabel(gate)}</span>
                                    )) : <span className="soft-pill">当前无待处理门禁</span>}
                                </div>
                            </div>
                            <div>
                                <span className="analytics-mini-label">已关闭</span>
                                <div className="gate-list compact">
                                    {completedGates.length > 0 ? completedGates.map((gate) => (
                                        <span key={gate} className="gate-chip">{formatGateLabel(gate)}</span>
                                    )) : <span className="soft-pill">尚无关闭记录</span>}
                                </div>
                            </div>
                        </div>
                    </article>
                </section>
            ) : null}
        </div>
    );
}
