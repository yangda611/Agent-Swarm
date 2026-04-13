import type {EChartsOption} from "echarts";
import type {WorkflowSummary} from "../types";
import {formatStatus, localizeText} from "../lib/format";
import {EChartPanel} from "./EChartPanel";

interface GovernanceFlowPanelProps {
    workflow: WorkflowSummary;
}

interface GateNodeView {
    id: string;
    shortLabel: string;
    fullLabel: string;
    stateLabel: string;
    status: "completed" | "pending";
    active: boolean;
}

interface GateBlueprint {
    id: string;
    shortLabel: string;
    aliases: string[];
}

const canonicalGateBlueprints: GateBlueprint[] = [
    {id: "intake", shortLabel: "需求整理", aliases: ["需求标准化", "Intake normalization"]},
    {id: "plan", shortLabel: "规划评审", aliases: ["规划评审", "Planning review"]},
    {id: "staff", shortLabel: "编制检查", aliases: ["编制与任务覆盖检查", "Staffing and task coverage check", "DAG staffing check"]},
    {id: "peer", shortLabel: "同行评审", aliases: ["执行产物同行评审", "Peer review on execution bundle"]},
    {id: "qa", shortLabel: "QA验收", aliases: ["QA 验收", "QA acceptance"]},
    {id: "security", shortLabel: "安全审核", aliases: ["安全审核", "Security review", "Security approval on outbound action"]},
    {id: "approval", shortLabel: "人工审批", aliases: ["人工审批", "Human approval"]},
    {id: "delivery", shortLabel: "交付签收", aliases: ["最终交付签收", "Final delivery sign-off"]},
];

export function GovernanceFlowPanel({workflow}: GovernanceFlowPanelProps) {
    const gateNodes = buildGateNodes(workflow);
    const pendingCount = gateNodes.filter((node) => node.status === "pending").length;
    const completedCount = gateNodes.filter((node) => node.status === "completed").length;
    const activeGate = gateNodes.find((node) => node.active) || gateNodes.find((node) => node.status === "pending");

    const option = buildGateChartOption(gateNodes);

    return (
        <article className="surface-card governance-flow-card">
            <div className="panel-header">
                <div>
                    <p className="section-kicker">流程门禁</p>
                    <h3 className="surface-title">治理流图</h3>
                </div>
                <div className="stat-pill-group">
                    <span className="soft-pill">{`已通过 ${completedCount} / ${gateNodes.length}`}</span>
                    <span className="soft-pill">{`待处理 ${pendingCount}`}</span>
                </div>
            </div>

            <div className="governance-flow-summary">
                <div className="governance-flow-metric">
                    <span>当前卡点</span>
                    <strong>{activeGate ? activeGate.shortLabel : "全部通过"}</strong>
                </div>
                <div className="governance-flow-metric">
                    <span>当前状态</span>
                    <strong>{activeGate ? activeGate.stateLabel : "已收口"}</strong>
                </div>
                <div className="governance-flow-metric">
                    <span>查看方式</span>
                    <strong>悬停节点</strong>
                </div>
            </div>

            <EChartPanel className="governance-flow-chart" minHeight={250} option={option}/>

            <div className="governance-flow-note">
                <div className="flow-status-legend">
                    <span className="legend-chip completed">已通过</span>
                    <span className="legend-chip pending">待处理</span>
                    <span className="legend-chip active">当前焦点</span>
                </div>
                <span>悬停节点可查看完整门禁名称与状态。</span>
            </div>
        </article>
    );
}

function buildGateNodes(workflow: WorkflowSummary): GateNodeView[] {
    const pendingEntries = workflow.pendingGates.map(parseGateEntry);
    const completedEntries = workflow.completedGates.map((gate) => ({
        baseLabel: stripGateStatus(gate),
        fullLabel: localizeText(stripGateStatus(gate)),
        stateLabel: "已通过",
        status: "completed" as const,
    }));

    const allEntries = [...pendingEntries, ...completedEntries];

    const orderedNodes: GateNodeView[] = canonicalGateBlueprints
        .flatMap((blueprint) => {
            const pendingMatch = pendingEntries.find((entry) => gateMatches(entry.baseLabel, blueprint.aliases));
            const completedMatch = completedEntries.find((entry) => gateMatches(entry.baseLabel, blueprint.aliases));
            const resolved = pendingMatch || completedMatch;

            if (!resolved) {
                return [];
            }

            return [{
                id: blueprint.id,
                shortLabel: blueprint.shortLabel,
                fullLabel: resolved.fullLabel,
                stateLabel: resolved.stateLabel,
                status: resolved.status,
                active: resolved.status === "pending" && pendingEntries[0]?.baseLabel === resolved.baseLabel,
            }];
        });

    const extraNodes: GateNodeView[] = allEntries
        .filter((entry) => !canonicalGateBlueprints.some((blueprint) => gateMatches(entry.baseLabel, blueprint.aliases)))
        .map((entry, index) => ({
            id: `extra-${index}`,
            shortLabel: entry.fullLabel,
            fullLabel: entry.fullLabel,
            stateLabel: entry.stateLabel,
            status: entry.status,
            active: entry.status === "pending" && pendingEntries[0]?.baseLabel === entry.baseLabel,
        }));

    return orderedNodes.concat(extraNodes);
}

function buildGateChartOption(gates: GateNodeView[]): EChartsOption {
    const nodes = gates.map((gate, index) => {
        const isCompleted = gate.status === "completed";
        const isActive = gate.active;
        const x = 70 + index * 120;
        const y = index % 2 === 0 ? 74 : 144;

        return {
            name: gate.shortLabel,
            x,
            y,
            symbol: "roundRect",
            symbolSize: [96, 44],
            value: gate.stateLabel,
            fullLabel: gate.fullLabel,
            itemStyle: {
                color: isCompleted ? "#ecfdf5" : "#fff7ed",
                borderColor: isCompleted ? "#34d399" : isActive ? "#f97316" : "#fdba74",
                borderWidth: isActive ? 2.5 : 1.5,
                shadowColor: isActive ? "rgba(249, 115, 22, 0.16)" : "rgba(148, 163, 184, 0.08)",
                shadowBlur: isActive ? 18 : 8,
            },
            label: {
                show: true,
                color: isCompleted ? "#047857" : "#9a3412",
                fontSize: 12,
                fontWeight: 700,
                lineHeight: 14,
            },
        };
    });

    const links = gates.slice(1).map((gate, index) => ({
        source: gates[index].shortLabel,
        target: gate.shortLabel,
        lineStyle: {
            color: gate.status === "completed" ? "#86efac" : "#fed7aa",
            width: gate.active ? 3 : 2,
            opacity: 1,
        },
    }));

    return {
        animationDuration: 260,
        tooltip: {
            trigger: "item",
            backgroundColor: "rgba(15, 23, 42, 0.94)",
            borderWidth: 0,
            textStyle: {
                color: "#f8fafc",
                fontSize: 12,
            },
            formatter: (params: any) => {
                const fullLabel = params?.data?.fullLabel || "";
                const stateLabel = params?.data?.value || "";
                return `${fullLabel}<br/>${stateLabel}`;
            },
        },
        series: [
            {
                type: "graph",
                layout: "none",
                roam: false,
                draggable: false,
                data: nodes,
                links,
                edgeSymbol: ["none", "arrow"],
                edgeSymbolSize: [0, 8],
                lineStyle: {
                    curveness: 0.08,
                },
                emphasis: {
                    scale: false,
                },
            },
        ],
    };
}

function parseGateEntry(value: string) {
    const match = value.match(/^(.*)\s+\(([^()]+)\)$/);
    if (!match) {
        return {
            baseLabel: value.trim(),
            fullLabel: localizeText(value.trim()),
            stateLabel: "待处理",
            status: "pending" as const,
        };
    }

    return {
        baseLabel: match[1].trim(),
        fullLabel: localizeText(match[1].trim()),
        stateLabel: formatStatus(match[2].trim()),
        status: "pending" as const,
    };
}

function stripGateStatus(value: string): string {
    const match = value.match(/^(.*)\s+\(([^()]+)\)$/);
    if (!match) {
        return value.trim();
    }
    return match[1].trim();
}

function gateMatches(label: string, aliases: readonly string[]): boolean {
    const normalizedLabel = stripGateStatus(label).trim().toLowerCase();
    return aliases.some((alias) => normalizedLabel === alias.trim().toLowerCase());
}
