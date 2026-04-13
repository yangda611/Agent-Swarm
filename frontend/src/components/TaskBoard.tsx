import {useMemo} from "react";
import {EChartPanel} from "./EChartPanel";
import type {TaskState} from "../types";
import {formatRisk, formatStage, formatStatus, localizeText} from "../lib/format";

interface TaskBoardProps {
    tasks: TaskState[];
}

const statusPalette: Record<string, string> = {
    queued: "#94a3b8",
    in_progress: "#2563eb",
    reviewing: "#8b5cf6",
    blocked: "#ef4444",
    completed: "#10b981",
};

const riskPalette: Record<string, string> = {
    low: "#10b981",
    medium: "#f59e0b",
    high: "#ef4444",
};

function shortTitle(value: string): string {
    const normalized = localizeText(value);
    return normalized.length > 14 ? `${normalized.slice(0, 14)}…` : normalized;
}

export function TaskBoard({tasks}: TaskBoardProps) {
    const summary = useMemo(() => {
        const active = tasks.filter((task) => task.status === "in_progress" || task.status === "reviewing" || task.status === "queued").length;
        const blocked = tasks.filter((task) => task.status === "blocked").length;
        const completed = tasks.filter((task) => task.status === "completed").length;
        return {active, blocked, completed};
    }, [tasks]);

    const pieOption = useMemo(() => {
        const counts = tasks.reduce<Record<string, number>>((collection, task) => {
            collection[task.status] = (collection[task.status] || 0) + 1;
            return collection;
        }, {});

        const data = Object.entries(counts).map(([status, value]) => ({
            name: formatStatus(status),
            value,
            itemStyle: {color: statusPalette[status] || "#64748b"},
        }));

        return {
            tooltip: {trigger: "item"},
            series: [
                {
                    type: "pie",
                    radius: ["54%", "78%"],
                    center: ["50%", "50%"],
                    label: {
                        formatter: "{b}\n{c}",
                        color: "#334155",
                        fontSize: 12,
                        lineHeight: 18,
                    },
                    labelLine: {length: 14, length2: 10},
                    data,
                },
            ],
        } as any;
    }, [tasks]);

    const progressOption = useMemo(() => {
        const sorted = [...tasks]
            .sort((left, right) => right.progress - left.progress)
            .slice(0, 8);

        return {
            grid: {left: 88, right: 24, top: 24, bottom: 18},
            tooltip: {trigger: "axis", axisPointer: {type: "shadow"}},
            xAxis: {
                type: "value",
                max: 100,
                splitLine: {lineStyle: {color: "rgba(148, 163, 184, 0.16)"}},
                axisLabel: {color: "#64748b", formatter: "{value}%"},
            },
            yAxis: {
                type: "category",
                data: sorted.map((task) => shortTitle(task.title)),
                axisLabel: {color: "#475569"},
                axisTick: {show: false},
                axisLine: {show: false},
            },
            series: [
                {
                    type: "bar",
                    barWidth: 14,
                    itemStyle: {
                        borderRadius: 999,
                        color: "#60a5fa",
                    },
                    data: sorted.map((task) => task.progress),
                },
            ],
        } as any;
    }, [tasks]);

    const stageRiskOption = useMemo(() => {
        const stages = Array.from(new Set(tasks.map((task) => formatStage(task.stage))));
        const risks: Array<"low" | "medium" | "high"> = ["low", "medium", "high"];

        return {
            legend: {bottom: 0, textStyle: {color: "#64748b"}},
            tooltip: {trigger: "axis", axisPointer: {type: "shadow"}},
            grid: {left: 48, right: 18, top: 28, bottom: 48},
            xAxis: {
                type: "category",
                data: stages,
                axisLabel: {color: "#475569"},
                axisLine: {lineStyle: {color: "#dbe5f0"}},
            },
            yAxis: {
                type: "value",
                splitLine: {lineStyle: {color: "rgba(148, 163, 184, 0.16)"}},
                axisLabel: {color: "#64748b"},
            },
            series: risks.map((risk) => ({
                type: "bar",
                name: formatRisk(risk),
                stack: "risk",
                barWidth: 18,
                itemStyle: {
                    borderRadius: risk === "high" ? [8, 8, 0, 0] : 0,
                    color: riskPalette[risk],
                },
                data: stages.map((stage) => tasks.filter((task) => formatStage(task.stage) === stage && task.riskLevel === risk).length),
            })),
        } as any;
    }, [tasks]);

    const blockedTasks = useMemo(() => tasks
        .filter((task) => task.status === "blocked")
        .slice(0, 4), [tasks]);

    return (
        <section className="page-stack">
            <section className="analytics-summary-row">
                <article className="surface-card analytics-kpi-card">
                    <span className="section-kicker">活跃任务</span>
                    <strong>{summary.active}</strong>
                </article>
                <article className="surface-card analytics-kpi-card">
                    <span className="section-kicker">阻塞任务</span>
                    <strong>{summary.blocked}</strong>
                </article>
                <article className="surface-card analytics-kpi-card">
                    <span className="section-kicker">已完成</span>
                    <strong>{summary.completed}</strong>
                </article>
            </section>

            <section className="analytics-grid">
                <article className="surface-card chart-card">
                    <div className="analytics-card-head">
                        <div>
                            <p className="section-kicker">状态分布</p>
                            <h2 className="surface-title">任务状态总览</h2>
                        </div>
                    </div>
                    <EChartPanel minHeight={300} option={pieOption}/>
                </article>

                <article className="surface-card chart-card">
                    <div className="analytics-card-head">
                        <div>
                            <p className="section-kicker">阶段风险</p>
                            <h2 className="surface-title">阶段与风险堆叠</h2>
                        </div>
                    </div>
                    <EChartPanel minHeight={300} option={stageRiskOption}/>
                </article>

                <article className="surface-card chart-card chart-card-wide">
                    <div className="analytics-card-head">
                        <div>
                            <p className="section-kicker">进度排行</p>
                            <h2 className="surface-title">任务推进速度</h2>
                        </div>
                        <span className="soft-pill">{`${tasks.length} 个原子任务`}</span>
                    </div>
                    <EChartPanel minHeight={340} option={progressOption}/>
                </article>

                <article className="surface-card chart-card">
                    <div className="analytics-card-head">
                        <div>
                            <p className="section-kicker">阻塞提醒</p>
                            <h2 className="surface-title">需要立即关注</h2>
                        </div>
                    </div>

                    {blockedTasks.length === 0 ? (
                        <div className="empty-state-card slim">
                            <strong>当前没有阻塞</strong>
                            <p>任务链路处于可推进状态。</p>
                        </div>
                    ) : (
                        <div className="analytics-chip-stack">
                            {blockedTasks.map((task) => (
                                <article key={task.id} className="analytics-chip-card">
                                    <strong>{shortTitle(task.title)}</strong>
                                    <span>{formatStage(task.stage)}</span>
                                </article>
                            ))}
                        </div>
                    )}
                </article>
            </section>
        </section>
    );
}
