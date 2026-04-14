import type {SettingsSummary, WorkflowSummary} from "../types";
import {
    formatPlannerSource,
    formatStageHeadline,
    formatTheme,
    localizeText,
} from "../lib/format";
import {GovernanceFlowPanel} from "./GovernanceFlowPanel";

interface OverviewPanelProps {
    hasActiveRun: boolean;
    settings: SettingsSummary;
    workflow: WorkflowSummary;
    onCreateRun: () => void;
    onOpenProviders: () => void;
}

export function OverviewPanel({hasActiveRun, settings, workflow, onCreateRun, onOpenProviders}: OverviewPanelProps) {
    const headline = hasActiveRun ? formatStageHeadline(workflow.currentStage) : "空项目";
    const summary = hasActiveRun
        ? localizeText(workflow.reviewMode)
        : "创建第一个任务后，这里会显示拆解、执行和审查进度；设置和 AI 接口可以先配置好。";

    return (
        <section className="page-stack">
            <article className="surface-card overview-spotlight-card">
                <div className="overview-spotlight-copy">
                    <p className="section-kicker">运行概览</p>
                    <h2 className="page-section-title">{headline}</h2>
                    <p className="page-section-copy">{summary}</p>
                </div>

                <div className="overview-spotlight-side">
                    <div className="overview-spotlight-meta">
                        <span className="soft-pill">{formatPlannerSource(workflow.plannerSource)}</span>
                        <span className="soft-pill">{localizeText(settings.primaryProvider)}</span>
                        <span className="soft-pill">{formatTheme(settings.theme)}</span>
                    </div>
                    <div className="overview-spotlight-actions">
                        <button className="primary-button" onClick={onCreateRun} type="button">
                            新建任务
                        </button>
                        <button className="ghost-button" onClick={onOpenProviders} type="button">
                            接口配置
                        </button>
                    </div>
                </div>
            </article>

            {hasActiveRun ? (
                <GovernanceFlowPanel workflow={workflow}/>
            ) : (
                <article className="surface-card">
                    <div className="empty-state-card">
                        <strong>还没有流程门禁</strong>
                        <p>创建第一个任务后，这里会显示治理流程。</p>
                    </div>
                </article>
            )}
        </section>
    );
}
