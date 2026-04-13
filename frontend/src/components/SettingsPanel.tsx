import {useEffect, useMemo, useState, type FormEvent} from "react";
import type {SettingsSummary, SettingsUpdateInput, WorkflowSummary} from "../types";
import {
    formatBudgetMode,
    formatPlannerSource,
    formatStageHeadline,
    formatTheme,
    localizeText,
} from "../lib/format";

interface SettingsPanelProps {
    busyLabel: string;
    onAdvance: () => Promise<void>;
    onOpenProviders: () => void;
    onReset: () => Promise<void>;
    onSaveSettings: (input: SettingsUpdateInput) => Promise<void>;
    settings: SettingsSummary;
    workflow: WorkflowSummary;
}

export function SettingsPanel({
    busyLabel,
    onAdvance,
    onOpenProviders,
    onReset,
    onSaveSettings,
    settings,
    workflow,
}: SettingsPanelProps) {
    const [form, setForm] = useState<SettingsUpdateInput>({
        concurrencyLimit: settings.concurrencyLimit,
        approvalPolicy: localizeText(settings.approvalPolicy),
        theme: settings.theme,
        budgetMode: settings.budgetMode,
    });

    useEffect(() => {
        setForm({
            concurrencyLimit: settings.concurrencyLimit,
            approvalPolicy: localizeText(settings.approvalPolicy),
            theme: settings.theme,
            budgetMode: settings.budgetMode,
        });
    }, [settings]);

    const isBusy = busyLabel.length > 0;
    const validation = useMemo(() => {
        const issues: string[] = [];
        if (form.concurrencyLimit < 1 || form.concurrencyLimit > 100) {
            issues.push("并发上限必须在 1 到 100 之间。");
        }
        if (form.approvalPolicy.trim().length < 8) {
            issues.push("审批策略建议至少写清 8 个字，方便团队理解治理规则。");
        }
        return issues;
    }, [form.approvalPolicy, form.concurrencyLimit]);

    async function submit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        if (validation.length > 0) {
            return;
        }
        await onSaveSettings(form);
    }

    return (
        <section className="page-stack">
            <article className="surface-card settings-studio-card">
                <div className="panel-header settings-studio-head">
                    <div>
                        <p className="section-kicker">系统设置</p>
                        <h2 className="page-section-title">运行控制台</h2>
                        <p className="page-section-copy">把控制动作和运行参数集中到同一块面板里，减少来回切换。</p>
                    </div>
                    <div className="stat-pill-group settings-studio-pills">
                        <span className="soft-pill">{formatStageHeadline(workflow.currentStage)}</span>
                        <span className="soft-pill">{formatPlannerSource(workflow.plannerSource)}</span>
                        <span className="soft-pill">{busyLabel || "当前就绪"}</span>
                    </div>
                </div>

                <div className="settings-studio-grid">
                    <aside className="settings-control-rail">
                        <section className="surface-subcard settings-action-card">
                            <div>
                                <p className="section-kicker">快捷操作</p>
                                <h3 className="surface-title">运行控制</h3>
                            </div>

                            <div className="settings-action-stack">
                                <button className="quick-action-button primary" disabled={isBusy} onClick={() => void onAdvance()} type="button">
                                    <span className="quick-action-icon">▶</span>
                                    <span>推进一阶段</span>
                                </button>
                                <button className="quick-action-button" disabled={isBusy} onClick={onOpenProviders} type="button">
                                    <span className="quick-action-icon">◎</span>
                                    <span>管理接口</span>
                                </button>
                                <button className="quick-action-button" disabled={isBusy} onClick={() => void onReset()} type="button">
                                    <span className="quick-action-icon">↺</span>
                                    <span>重置流程</span>
                                </button>
                            </div>
                        </section>

                        <section className="surface-subcard settings-runtime-card">
                            <p className="section-kicker">运行概况</p>
                            <div className="detail-list compact">
                                <div className="detail-row multi">
                                    <span>当前阶段</span>
                                    <strong>{formatStageHeadline(workflow.currentStage)}</strong>
                                </div>
                                <div className="detail-row multi">
                                    <span>规划来源</span>
                                    <strong>{formatPlannerSource(workflow.plannerSource)}</strong>
                                </div>
                                <div className="detail-row">
                                    <span>主路由</span>
                                    <strong>{localizeText(settings.primaryProvider)}</strong>
                                </div>
                            </div>
                        </section>
                    </aside>

                    <section className="settings-editor-shell">
                        <div className="settings-current-strip">
                            <div className="settings-current-item">
                                <span>当前主题</span>
                                <strong>{formatTheme(settings.theme)}</strong>
                            </div>
                            <div className="settings-current-item">
                                <span>预算策略</span>
                                <strong>{formatBudgetMode(settings.budgetMode)}</strong>
                            </div>
                            <div className="settings-current-item">
                                <span>并发上限</span>
                                <strong>{settings.concurrencyLimit}</strong>
                            </div>
                        </div>

                        <form className="settings-form-shell" onSubmit={submit}>
                            <section className="surface-subcard settings-form-card">
                                <div className="settings-form-head">
                                    <div>
                                        <p className="section-kicker">运行参数</p>
                                        <h3 className="surface-title">基础配置</h3>
                                    </div>
                                    <p className="page-section-copy">这部分会直接影响并发规模、主题和预算策略。</p>
                                </div>

                                <div className="form-grid compact settings-form-grid">
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
                                        <small className="field-help">建议按照机器能力和接口配额设置。</small>
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

                                    <label className="field-group span-two">
                                        <span className="field-label">预算模式</span>
                                        <select
                                            className="field-input"
                                            disabled={isBusy}
                                            onChange={(event) => setForm((current) => ({...current, budgetMode: event.target.value}))}
                                            value={form.budgetMode}
                                        >
                                            <option value="guardrailed">稳健受控</option>
                                            <option value="balanced">均衡模式</option>
                                            <option value="throughput-first">吞吐优先</option>
                                        </select>
                                    </label>
                                </div>
                            </section>

                            <section className="surface-subcard settings-form-card">
                                <div className="settings-form-head">
                                    <div>
                                        <p className="section-kicker">治理规则</p>
                                        <h3 className="surface-title">审批策略</h3>
                                    </div>
                                    <p className="page-section-copy">写清哪些动作必须人工审批，哪些情况需要直接阻塞。</p>
                                </div>

                                <label className="field-group">
                                    <textarea
                                        className="field-input field-area settings-policy-area"
                                        disabled={isBusy}
                                        onChange={(event) => setForm((current) => ({...current, approvalPolicy: event.target.value}))}
                                        rows={6}
                                        value={form.approvalPolicy}
                                    />
                                </label>
                            </section>

                            <div className="surface-subcard settings-save-bar">
                                <div className="settings-save-copy">
                                    {validation.length > 0 ? (
                                        <ul className="summary-bullet-list">
                                            {validation.map((issue) => (
                                                <li key={issue} className="error">
                                                    {issue}
                                                </li>
                                            ))}
                                        </ul>
                                    ) : (
                                        <p className="field-help">当前设置已经满足保存要求，可以直接写入运行时。</p>
                                    )}
                                </div>
                                <div className="form-actions">
                                    <button className="primary-button" disabled={isBusy || validation.length > 0} type="submit">
                                        保存设置
                                    </button>
                                </div>
                            </div>
                        </form>
                    </section>
                </div>
            </article>
        </section>
    );
}
