import {useEffect, useState, type FormEvent} from "react";
import type {MetricCard, SettingsSummary, SettingsUpdateInput, WorkflowSummary} from "../types";
import {formatBudgetMode, formatModelTier, formatPlannerSource, formatStageHeadline, formatTheme} from "../lib/format";

interface SidebarProps {
    metrics: MetricCard[];
    workflow: WorkflowSummary;
    settings: SettingsSummary;
    busyLabel: string;
    onAdvance: () => void;
    onReset: () => void;
    onSaveSettings: (input: SettingsUpdateInput) => Promise<void>;
}

export function Sidebar({metrics, workflow, settings, busyLabel, onAdvance, onReset, onSaveSettings}: SidebarProps) {
    const [form, setForm] = useState<SettingsUpdateInput>({
        concurrencyLimit: settings.concurrencyLimit,
        approvalPolicy: settings.approvalPolicy,
        theme: settings.theme,
        budgetMode: settings.budgetMode,
    });

    useEffect(() => {
        setForm({
            concurrencyLimit: settings.concurrencyLimit,
            approvalPolicy: settings.approvalPolicy,
            theme: settings.theme,
            budgetMode: settings.budgetMode,
        });
    }, [settings]);

    const isBusy = busyLabel.length > 0;

    async function submitSettings(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        await onSaveSettings(form);
    }

    return (
        <aside className="sidebar">
            <section className="panel-block">
                <p className="panel-label">控制台</p>
                <div className="control-grid">
                    <button className="control-button" disabled={isBusy} onClick={onAdvance} type="button">
                        推进一阶段
                    </button>
                    <button className="control-button secondary" disabled={isBusy} onClick={onReset} type="button">
                        重置流程
                    </button>
                </div>
                <p className="settings-copy">{busyLabel || "所有控制项都会写回本地 SQLite 运行时存储。"}</p>
            </section>

            <section className="panel-block">
                <p className="panel-label">运行指标</p>
                <div className="metric-grid">
                    {metrics.map((metric) => (
                        <article key={metric.label} className={`metric-card ${metric.accent}`}>
                            <p className="metric-title">{metric.label}</p>
                            <p className="metric-value">{metric.value}</p>
                            <p className="metric-detail">{metric.detail}</p>
                        </article>
                    ))}
                </div>
            </section>

            <section className="panel-block">
                <p className="panel-label">流程门禁</p>
                <article className="workflow-card">
                    <h2 className="panel-title">{formatStageHeadline(workflow.currentStage)}</h2>
                    <p className="workflow-copy">{workflow.atomicTaskPolicy}</p>
                    <p className="workflow-copy">{workflow.reviewMode}</p>
                    <p className="workflow-copy">规划来源：{formatPlannerSource(workflow.plannerSource)}</p>
                </article>
                <div className="gate-list">
                    {workflow.completedGates.map((gate) => (
                        <div key={gate} className="gate-chip">
                            {gate}
                        </div>
                    ))}
                    {workflow.pendingGates.map((gate) => (
                        <div key={gate} className="gate-chip pending">
                            {gate}
                        </div>
                    ))}
                </div>
            </section>

            <section className="panel-block">
                <p className="panel-label">主设置</p>
                <form className="settings-form" onSubmit={submitSettings}>
                    <label className="field-group">
                        <span className="field-label">并发上限</span>
                        <input
                            className="field-input"
                            disabled={isBusy}
                            max={100}
                            min={1}
                            onChange={(event) => setForm((current) => ({...current, concurrencyLimit: Number(event.target.value)}))}
                            type="number"
                            value={form.concurrencyLimit}
                        />
                    </label>

                    <label className="field-group">
                        <span className="field-label">主题风格</span>
                        <select
                            className="field-input"
                            disabled={isBusy}
                            onChange={(event) => setForm((current) => ({...current, theme: event.target.value}))}
                            value={form.theme}
                        >
                            <option value="pixel-hq-amber">像素总部·琥珀</option>
                            <option value="pixel-hq-mint">像素总部·薄荷</option>
                            <option value="night-shift">夜班模式</option>
                        </select>
                    </label>

                    <label className="field-group">
                        <span className="field-label">预算模式</span>
                        <select
                            className="field-input"
                            disabled={isBusy}
                            onChange={(event) => setForm((current) => ({...current, budgetMode: event.target.value}))}
                            value={form.budgetMode}
                        >
                            <option value="guardrailed">稳健受控</option>
                            <option value="balanced">平衡模式</option>
                            <option value="throughput-first">吞吐优先</option>
                        </select>
                    </label>

                    <label className="field-group">
                        <span className="field-label">审批策略</span>
                        <textarea
                            className="field-input field-area"
                            disabled={isBusy}
                            onChange={(event) => setForm((current) => ({...current, approvalPolicy: event.target.value}))}
                            rows={4}
                            value={form.approvalPolicy}
                        />
                    </label>

                    <button className="control-button" disabled={isBusy} type="submit">
                        保存设置
                    </button>
                </form>

                <article className="settings-card">
                    <p className="metric-title">并发上限</p>
                    <p className="metric-value">{settings.concurrencyLimit}</p>
                    <p className="settings-copy">{settings.approvalPolicy}</p>
                    <p className="settings-copy">主题：{formatTheme(settings.theme)}</p>
                    <p className="settings-copy">预算模式：{formatBudgetMode(settings.budgetMode)}</p>
                    <p className="settings-copy">主路由：{settings.primaryProvider}</p>
                    <p className="settings-copy">已配置接口：{settings.aiProviders.length}</p>
                </article>
                <div className="profile-list">
                    {settings.modelProfiles.map((profile) => (
                        <article key={profile.tier} className="profile-card">
                            <div className="profile-row">
                                <span className="profile-tier">{formatModelTier(profile.tier)}</span>
                                <span className="profile-binding">{profile.binding}</span>
                                <span className="settings-copy">{profile.responsibility}</span>
                            </div>
                        </article>
                    ))}
                </div>
            </section>
        </aside>
    );
}
