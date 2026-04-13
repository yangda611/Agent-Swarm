import type {SettingsSummary, WorkflowSummary} from "../types";
import {
    formatPlannerSource,
    formatStageHeadline,
    formatTheme,
    localizeText,
} from "../lib/format";
import {GovernanceFlowPanel} from "./GovernanceFlowPanel";

interface OverviewPanelProps {
    settings: SettingsSummary;
    workflow: WorkflowSummary;
    onCreateRun: () => void;
    onOpenProviders: () => void;
}

export function OverviewPanel({settings, workflow, onCreateRun, onOpenProviders}: OverviewPanelProps) {
    return (
        <section className="page-stack">
            <article className="surface-card overview-spotlight-card">
                <div className="overview-spotlight-copy">
                    <p className="section-kicker">运行概览</p>
                    <h2 className="page-section-title">{formatStageHeadline(workflow.currentStage)}</h2>
                    <p className="page-section-copy">{localizeText(workflow.reviewMode)}</p>
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

            <GovernanceFlowPanel workflow={workflow}/>
        </section>
    );
}
