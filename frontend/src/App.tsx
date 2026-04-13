import {useEffect, useMemo, useState} from "react";
import "./App.css";
import {
    AdvanceDemoRun,
    CreateRun,
    DeleteAIProvider,
    GetDashboardState,
    ProbeAIProvider,
    ResetDemoRun,
    UpdateRuntimeSettings,
    UpsertAIProvider,
} from "../wailsjs/go/main/App";
import {AIProviderPanel} from "./components/AIProviderPanel";
import {DialogModal} from "./components/DialogModal";
import {InspectorDrawer} from "./components/InspectorDrawer";
import {NavigationRail} from "./components/NavigationRail";
import {OfficeScene} from "./components/OfficeScene";
import {OverviewPanel} from "./components/OverviewPanel";
import {RunComposerPanel} from "./components/RunComposerPanel";
import {SettingsPanel} from "./components/SettingsPanel";
import {TaskBoard} from "./components/TaskBoard";
import {TimelinePanel} from "./components/TimelinePanel";
import type {AIProviderInput, DashboardState, RunCreationInput, SettingsUpdateInput} from "./types";
import {formatStageHeadline, localizeText, normalizeStageKey} from "./lib/format";

type OfficeViewMode = "desks" | "routes" | "events";
type ToastTone = "success" | "warning" | "error";

interface AppToast {
    id: number;
    tone: ToastTone;
    title: string;
    description: string;
}

const navigationItems = [
    {id: "overview", label: "概览", hint: "总览", icon: "总"},
    {id: "office", label: "办公室", hint: "现场", icon: "办"},
    {id: "tasks", label: "任务板", hint: "原子", icon: "板"},
    {id: "timeline", label: "时间线", hint: "回放", icon: "线"},
    {id: "settings", label: "设置", hint: "控制", icon: "设"},
];

const officeTabs: Array<{id: OfficeViewMode; label: string; description: string}> = [
    {id: "desks", label: "工位视图", description: "查看智能体当前工位和任务负载"},
    {id: "routes", label: "交接路线", description: "查看资料流转路径和在途交接"},
    {id: "events", label: "事件回放", description: "查看当前阶段门禁、告警和治理状态"},
];

const pageMeta: Record<string, {title: string; description: string}> = {
    overview: {
        title: "概览",
        description: "查看运行状态、治理门禁、最近事件和工作台配置。",
    },
    office: {
        title: "办公室现场",
        description: "从工位、路线和事件三个视角查看多智能体协作现场。",
    },
    tasks: {
        title: "任务板",
        description: "查看原子任务状态、负责人、依赖关系和当前产物。",
    },
    timeline: {
        title: "时间线",
        description: "按时间回看需求接入、规划、执行、审核与交付过程。",
    },
    settings: {
        title: "设置",
        description: "调整运行参数并执行工作台级控制动作。",
    },
};

interface RunActionOptions {
    forceStagePulse?: boolean;
    jumpToSectionId?: string;
}

function App() {
    const [snapshot, setSnapshot] = useState<DashboardState | null>(null);
    const [selectedAgentId, setSelectedAgentId] = useState<string | null>(null);
    const [busyLabel, setBusyLabel] = useState("");
    const [errorMessage, setErrorMessage] = useState("");
    const [activeSectionId, setActiveSectionId] = useState("overview");
    const [officeViewMode, setOfficeViewMode] = useState<OfficeViewMode>("desks");
    const [isAgentDetailOpen, setAgentDetailOpen] = useState(false);
    const [isComposerOpen, setComposerOpen] = useState(false);
    const [isProviderOpen, setProviderOpen] = useState(false);
    const [stageMotionKey, setStageMotionKey] = useState(0);
    const [toasts, setToasts] = useState<AppToast[]>([]);
    const [isAutoRunning, setAutoRunning] = useState(false);

    useEffect(() => {
        void loadDashboard();
    }, []);

    useEffect(() => {
        if (!isAutoRunning || !snapshot) {
            return;
        }
        if (snapshot.runStep >= snapshot.maxStep) {
            setAutoRunning(false);
            return;
        }
        if (busyLabel.length > 0 || errorMessage.length > 0 || isComposerOpen || isProviderOpen) {
            return;
        }

        const timer = window.setTimeout(() => {
            void runAction("自动推进流程", AdvanceDemoRun, {forceStagePulse: true});
        }, snapshot.runStep <= 1 ? 1800 : 2400);

        return () => window.clearTimeout(timer);
    }, [
        busyLabel,
        errorMessage.length,
        isAutoRunning,
        isComposerOpen,
        isProviderOpen,
        snapshot,
    ]);

    function pushToast(tone: ToastTone, title: string, description = "") {
        const id = Date.now() + Math.floor(Math.random() * 1000);
        setToasts((current) => [...current, {id, tone, title, description}]);

        window.setTimeout(() => {
            setToasts((current) => current.filter((toast) => toast.id !== id));
        }, 4200);
    }

    async function loadDashboard() {
        try {
            const data = await GetDashboardState() as unknown as DashboardState;
            applySnapshot(data);
        } catch (error) {
            const message = extractErrorMessage(error);
            setErrorMessage(message);
            pushToast("error", "加载失败", message);
        }
    }

    function applySnapshot(data: DashboardState) {
        setSnapshot(data);
        setSelectedAgentId((current) => {
            if (current && data.agents.some((agent) => agent.id === current)) {
                return current;
            }
            return data.suggestedAgentId || data.agents[0]?.id || null;
        });
    }

    async function runAction(label: string, action: () => Promise<unknown>, options: RunActionOptions = {}) {
        const previousStageKey = snapshot ? normalizeStageKey(snapshot.workflow.currentStage) : "";

        setBusyLabel(label);
        setErrorMessage("");

        try {
            const data = await action() as DashboardState;
            applySnapshot(data);

            const nextStageKey = normalizeStageKey(data.workflow.currentStage);
            if (options.forceStagePulse || previousStageKey !== nextStageKey) {
                setStageMotionKey((current) => current + 1);
            }
            if (options.jumpToSectionId) {
                setActiveSectionId(options.jumpToSectionId);
            }
            return true;
        } catch (error) {
            const message = extractErrorMessage(error);
            setErrorMessage(message);
            pushToast("error", `${label}失败`, message);
            return false;
        } finally {
            setBusyLabel("");
        }
    }

    async function runProbe(label: string, action: () => Promise<string>) {
        setBusyLabel(label);
        setErrorMessage("");

        try {
            return await action();
        } catch (error) {
            const message = extractErrorMessage(error);
            setErrorMessage(message);
            pushToast("error", "测试请求失败", message);
            return "接口测试失败，当前路由返回了异常运行错误。";
        } finally {
            setBusyLabel("");
        }
    }

    const page = useMemo(() => pageMeta[activeSectionId] || pageMeta.overview, [activeSectionId]);

    if (!snapshot) {
        return (
            <div className="loading-shell">
                <div className="loading-card light">
                    <div className="loading-signal"/>
                    <h1>正在启动 maliang swarm</h1>
                    <p>系统正在恢复运行数据与接口配置，请稍候。</p>
                </div>
            </div>
        );
    }

    const selectedAgent = snapshot.agents.find((agent) => agent.id === selectedAgentId) || snapshot.agents[0];
    const currentStageKey = normalizeStageKey(snapshot.workflow.currentStage);
    const focusZoneId = zoneForStage(currentStageKey);
    const focusedZone = snapshot.zones.find((zone) => zone.id === focusZoneId);
    const activeOfficeTab = officeTabs.find((tab) => tab.id === officeViewMode) || officeTabs[0];
    return (
        <div className="app-shell app-shell-modern">
            <NavigationRail
                activeId={activeSectionId}
                items={navigationItems}
                onNavigate={setActiveSectionId}
            />

            <main className="workspace workspace-modern">
                <div className="workspace-scroll modern">
                    <header className="page-header">
                        <div>
                            <p className="page-tag">maliang swarm</p>
                            <h1>{page.title}</h1>
                            <p className="page-description">{page.description}</p>
                        </div>

                        <div className="page-toolbar">
                            <button className="ghost-button" disabled={busyLabel.length > 0} onClick={() => void loadDashboard()} type="button">
                                刷新
                            </button>
                            <button className="ghost-button" disabled={busyLabel.length > 0} onClick={() => setProviderOpen(true)} type="button">
                                接口配置
                            </button>
                            <button className="primary-button" disabled={busyLabel.length > 0} onClick={() => setComposerOpen(true)} type="button">
                                新建任务
                            </button>
                        </div>
                    </header>

                    {errorMessage ? <div className="status-banner light">{localizeText(errorMessage)}</div> : null}

                    {activeSectionId === "overview" ? (
                        <OverviewPanel
                            onCreateRun={() => setComposerOpen(true)}
                            onOpenProviders={() => setProviderOpen(true)}
                            settings={snapshot.settings}
                            workflow={snapshot.workflow}
                        />
                    ) : null}

                    {activeSectionId === "office" ? (
                        <section className="page-stack">
                            <article className="surface-card tight">
                                <div className="panel-header">
                                    <div>
                                        <p className="section-kicker">办公室视角</p>
                                        <h2 className="page-section-title">{activeOfficeTab.label}</h2>
                                        <p className="page-section-copy">{activeOfficeTab.description}</p>
                                    </div>
                                    <div className="stat-pill-group">
                                        <span className="soft-pill">{formatStageHeadline(snapshot.workflow.currentStage)}</span>
                                        <span className="soft-pill">{localizeText(focusedZone?.name || "指挥区")}</span>
                                        <span className="soft-pill">点击智能体查看详情与提示词</span>
                                    </div>
                                </div>

                                <div className="page-tab-row">
                                    {officeTabs.map((tab) => (
                                        <button
                                            key={tab.id}
                                            className={`page-tab-button ${officeViewMode === tab.id ? "active" : ""}`}
                                            onClick={() => setOfficeViewMode(tab.id)}
                                            type="button"
                                        >
                                            <span>{tab.label}</span>
                                            <small>{tab.description}</small>
                                        </button>
                                    ))}
                                </div>
                            </article>

                            <article className="surface-card office-surface office-surface-full">
                                <div className="panel-header">
                                    <div>
                                        <p className="section-kicker">当前现场</p>
                                        <h2 className="page-section-title">{formatStageHeadline(snapshot.workflow.currentStage)}</h2>
                                        <p className="page-section-copy">{localizeText(focusedZone?.purpose || "主脑正在协调当前阶段的办公室现场。")}</p>
                                    </div>
                                    <div className="stat-pill-group office-inline-tip-row">
                                        <span className="soft-pill">{`当前任务：${localizeText(snapshot.title)}`}</span>
                                        <span className="soft-pill muted-soft-pill">详情改为点击弹窗查看</span>
                                    </div>
                                </div>

                                <OfficeScene
                                    agents={snapshot.agents}
                                    completedGates={snapshot.workflow.completedGates}
                                    focusZoneId={focusZoneId}
                                    handoffs={snapshot.handoffs}
                                    pendingGates={snapshot.workflow.pendingGates}
                                    selectedAgentId={selectedAgent?.id}
                                    stageMotionKey={stageMotionKey}
                                    stageTitle={formatStageHeadline(snapshot.workflow.currentStage)}
                                    viewMode={officeViewMode}
                                    zones={snapshot.zones}
                                    onSelectAgent={(agentId) => {
                                        setSelectedAgentId(agentId);
                                        setAgentDetailOpen(true);
                                    }}
                                />
                            </article>
                        </section>
                    ) : null}

                    {activeSectionId === "tasks" ? <TaskBoard tasks={snapshot.tasks}/> : null}
                    {activeSectionId === "timeline" ? <TimelinePanel items={snapshot.timeline}/> : null}

                    {activeSectionId === "settings" ? (
                        <SettingsPanel
                            busyLabel={busyLabel}
                            onAdvance={async () => {
                                const success = await runAction("推进流程阶段", AdvanceDemoRun);
                                if (success) {
                                    if (snapshot.runStep + 1 >= snapshot.maxStep) {
                                        setAutoRunning(false);
                                    }
                                    pushToast("success", "流程已推进", "办公室现场和任务状态已同步更新。")
                                }
                            }}
                            onOpenProviders={() => setProviderOpen(true)}
                            onReset={async () => {
                                const success = await runAction("重置当前流程", ResetDemoRun, {forceStagePulse: true});
                                if (success) {
                                    setAutoRunning(false);
                                    pushToast("success", "流程已重置", "已回到初始阶段，保留当前设置与接口。")
                                }
                            }}
                            onSaveSettings={async (input: SettingsUpdateInput) => {
                                const success = await runAction("保存主设置", () => UpdateRuntimeSettings(input));
                                if (success) {
                                    pushToast("success", "设置已保存", "新的运行参数已经生效。")
                                }
                            }}
                            settings={snapshot.settings}
                            workflow={snapshot.workflow}
                        />
                    ) : null}
                </div>
            </main>

            <DialogModal
                open={isAgentDetailOpen && activeSectionId === "office" && Boolean(selectedAgent)}
                onClose={() => setAgentDetailOpen(false)}
                title={selectedAgent ? `${selectedAgent.name} · 智能体详情` : "智能体详情"}
                width="880px"
            >
                <InspectorDrawer agent={selectedAgent} variant="modal"/>
            </DialogModal>

            <DialogModal
                open={isComposerOpen}
                onClose={() => {
                    if (!busyLabel) {
                        setComposerOpen(false);
                    }
                }}
                title="新建任务"
                width="980px"
            >
                <RunComposerPanel
                    busyLabel={busyLabel}
                    onCreateRun={async (input: RunCreationInput) => {
                        const success = await runAction("编译新的任务运行", () => CreateRun(input), {
                            forceStagePulse: true,
                            jumpToSectionId: "office",
                        });
                        if (success) {
                            setAutoRunning(true);
                            pushToast("success", "任务已创建", "已自动进入办公室现场视图。可继续推进流程阶段。")
                            setComposerOpen(false);
                        }
                    }}
                />
            </DialogModal>

            <DialogModal
                open={isProviderOpen}
                onClose={() => {
                    if (!busyLabel) {
                        setProviderOpen(false);
                    }
                }}
                title="接口配置"
                width="1240px"
            >
                <AIProviderPanel
                    busyLabel={busyLabel}
                    onDeleteProvider={async (id: string) => {
                        const success = await runAction("移除接口配置", () => DeleteAIProvider(id));
                        if (success) {
                            pushToast("success", "接口已移除", "配置列表已完成同步更新。")
                        }
                    }}
                    onNotify={pushToast}
                    onProbeProvider={(input: AIProviderInput) => runProbe("测试接口路由", () => ProbeAIProvider(input))}
                    onUpsertProvider={async (input: AIProviderInput) => {
                        await runAction("保存接口配置", () => UpsertAIProvider(input));
                    }}
                    primaryProvider={snapshot.settings.primaryProvider}
                    providers={snapshot.settings.aiProviders}
                />
            </DialogModal>

            {toasts.length > 0 ? (
                <div className="toast-stack" aria-live="polite" aria-atomic="false">
                    {toasts.map((toast) => (
                        <article key={toast.id} className={`toast-card ${toast.tone}`}>
                            <strong>{toast.title}</strong>
                            {toast.description ? <p>{toast.description}</p> : null}
                        </article>
                    ))}
                </div>
            ) : null}
        </div>
    );
}

export default App;

function zoneForStage(stageKey: string): string {
    switch (stageKey) {
        case "intake":
            return "command";
        case "planning":
            return "planning";
        case "execution":
            return "engineering";
        case "review":
            return "review";
        case "delivery":
            return "delivery";
        default:
            return "command";
    }
}

function extractErrorMessage(error: unknown): string {
    if (error instanceof Error && error.message.trim()) {
        return error.message;
    }

    if (typeof error === "string" && error.trim()) {
        return error;
    }

    if (error && typeof error === "object") {
        for (const value of Object.values(error as Record<string, unknown>)) {
            if (typeof value === "string" && value.trim()) {
                return value;
            }
        }
    }

    return "控制台遇到了未预期的运行错误。";
}
