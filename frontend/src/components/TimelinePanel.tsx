import {useMemo, useState} from "react";
import {EChartPanel} from "./EChartPanel";
import type {TimelineItem} from "../types";
import {formatTimelineKind, localizeText} from "../lib/format";

interface TimelinePanelProps {
    items: TimelineItem[];
}

type TimelineFilter = "all" | "planning" | "execution" | "review" | "delivery";

const filterMap: Record<TimelineFilter, string[]> = {
    all: [],
    planning: ["intake", "planning", "decomposition", "staffing"],
    execution: ["execution", "artifact", "handoff"],
    review: ["review", "security", "audit"],
    delivery: ["delivery", "archive"],
};

const timelineFilters: Array<{id: TimelineFilter; label: string}> = [
    {id: "all", label: "全部"},
    {id: "planning", label: "规划"},
    {id: "execution", label: "执行"},
    {id: "review", label: "审核"},
    {id: "delivery", label: "交付"},
];

function shortText(value: string, size = 16): string {
    const normalized = localizeText(value);
    return normalized.length > size ? `${normalized.slice(0, size)}…` : normalized;
}

export function TimelinePanel({items}: TimelinePanelProps) {
    const [activeFilter, setActiveFilter] = useState<TimelineFilter>("all");

    const filteredItems = useMemo(() => {
        const acceptedKinds = filterMap[activeFilter];
        if (acceptedKinds.length === 0) {
            return items;
        }
        return items.filter((item) => acceptedKinds.includes(item.kind));
    }, [activeFilter, items]);

    const replayOption = useMemo(() => {
        const kindLabels = Array.from(new Set(filteredItems.map((item) => formatTimelineKind(item.kind))));
        const categories = filteredItems.map((item, index) => `${item.timeLabel} #${index + 1}`);

        return {
            grid: {left: 78, right: 22, top: 24, bottom: 40},
            tooltip: {
                trigger: "item",
                formatter: (params: any) => {
                    const [, yIndex, title, detail, timeLabel, kindLabel] = params.data;
                    return [
                        `<strong>${title}</strong>`,
                        `${kindLabel} · ${timeLabel}`,
                        detail,
                    ].join("<br/>");
                },
            },
            xAxis: {
                type: "category",
                data: categories,
                axisLabel: {
                    color: "#64748b",
                    interval: 0,
                    formatter: (value: string) => value.split(" #")[0],
                },
                axisLine: {lineStyle: {color: "#dbe5f0"}},
            },
            yAxis: {
                type: "category",
                data: kindLabels,
                axisLabel: {color: "#475569"},
                axisTick: {show: false},
                axisLine: {show: false},
            },
            series: [
                {
                    type: "scatter",
                    symbolSize: 20,
                    itemStyle: {
                        color: "#2563eb",
                        shadowBlur: 12,
                        shadowColor: "rgba(37, 99, 235, 0.22)",
                    },
                    data: filteredItems.map((item, index) => [
                        index,
                        kindLabels.indexOf(formatTimelineKind(item.kind)),
                        shortText(item.title, 20),
                        shortText(item.detail, 28),
                        item.timeLabel,
                        formatTimelineKind(item.kind),
                    ]),
                },
            ],
        } as any;
    }, [filteredItems]);

    const densityOption = useMemo(() => {
        const grouped = filteredItems.reduce<Record<string, number>>((collection, item) => {
            const kind = formatTimelineKind(item.kind);
            collection[kind] = (collection[kind] || 0) + 1;
            return collection;
        }, {});

        const entries = Object.entries(grouped);

        return {
            grid: {left: 52, right: 18, top: 18, bottom: 20},
            tooltip: {trigger: "axis", axisPointer: {type: "shadow"}},
            xAxis: {
                type: "value",
                splitLine: {lineStyle: {color: "rgba(148, 163, 184, 0.16)"}},
                axisLabel: {color: "#64748b"},
            },
            yAxis: {
                type: "category",
                data: entries.map(([kind]) => kind),
                axisLabel: {color: "#475569"},
                axisTick: {show: false},
                axisLine: {show: false},
            },
            series: [
                {
                    type: "bar",
                    data: entries.map(([, value]) => value),
                    barWidth: 14,
                    itemStyle: {
                        borderRadius: 999,
                        color: "#8b5cf6",
                    },
                },
            ],
        } as any;
    }, [filteredItems]);

    const recentItems = filteredItems.slice(-4).reverse();

    return (
        <section className="page-stack">
            <div className="filter-chip-row compact-top">
                {timelineFilters.map((filter) => (
                    <button
                        key={filter.id}
                        className={`filter-chip ${activeFilter === filter.id ? "active" : ""}`}
                        onClick={() => setActiveFilter(filter.id)}
                        type="button"
                    >
                        {filter.label}
                    </button>
                ))}
            </div>

            <section className="analytics-grid">
                <article className="surface-card chart-card chart-card-wide">
                    <div className="analytics-card-head">
                        <div>
                            <p className="section-kicker">事件轨迹</p>
                            <h2 className="surface-title">运行事件地图</h2>
                        </div>
                        <span className="soft-pill">{`${filteredItems.length} 条事件`}</span>
                    </div>
                    <EChartPanel minHeight={360} option={replayOption}/>
                </article>

                <article className="surface-card chart-card">
                    <div className="analytics-card-head">
                        <div>
                            <p className="section-kicker">事件密度</p>
                            <h2 className="surface-title">阶段聚集度</h2>
                        </div>
                    </div>
                    <EChartPanel minHeight={360} option={densityOption}/>
                </article>

                <article className="surface-card chart-card">
                    <div className="analytics-card-head">
                        <div>
                            <p className="section-kicker">最近回放</p>
                            <h2 className="surface-title">关键节点</h2>
                        </div>
                    </div>

                    {recentItems.length === 0 ? (
                        <div className="empty-state-card slim">
                            <strong>当前没有事件</strong>
                            <p>推进流程后，这里会出现新的回放节点。</p>
                        </div>
                    ) : (
                        <div className="analytics-chip-stack">
                            {recentItems.map((item) => (
                                <article key={`${item.timeLabel}-${item.title}`} className="analytics-chip-card timeline-chip-card">
                                    <strong>{shortText(item.title, 18)}</strong>
                                    <span>{`${item.timeLabel} · ${formatTimelineKind(item.kind)}`}</span>
                                </article>
                            ))}
                        </div>
                    )}
                </article>
            </section>
        </section>
    );
}
