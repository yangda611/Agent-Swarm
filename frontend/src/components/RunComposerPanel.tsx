import {useMemo, useState, type FormEvent} from "react";
import type {RunCreationInput} from "../types";

interface RunComposerPanelProps {
    busyLabel: string;
    onCreateRun: (input: RunCreationInput) => Promise<void>;
}

export function RunComposerPanel({busyLabel, onCreateRun}: RunComposerPanelProps) {
    const [form, setForm] = useState<RunCreationInput>({
        title: "",
        mission: "",
        deliverable: "",
        taskType: "engineering",
        priority: "high",
        maxAgents: 12,
        requiresQA: true,
        requiresSecurity: true,
        requiresHumanApproval: true,
    });

    const isBusy = busyLabel.length > 0;
    const trimmedTitle = form.title.trim();
    const trimmedMission = form.mission.trim();
    const trimmedDeliverable = form.deliverable.trim();

    const validation = useMemo(() => {
        const issues: string[] = [];

        if (trimmedMission.length < 12) {
            issues.push("任务目标至少填写 12 个字。");
        }
        if (trimmedDeliverable.length < 6) {
            issues.push("请补充清晰的交付物说明。");
        }
        if (form.maxAgents < 1 || form.maxAgents > 100) {
            issues.push("最大智能体数必须在 1 到 100 之间。");
        }
        if (form.requiresHumanApproval && !form.requiresSecurity) {
            issues.push("建议人工审批与安全审查同时开启。");
        }

        return issues;
    }, [form.maxAgents, form.requiresHumanApproval, form.requiresSecurity, trimmedDeliverable.length, trimmedMission.length]);

    const canSubmit = validation.filter((issue) => !issue.includes("建议")).length === 0 && !isBusy;

    async function submit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        if (!canSubmit) {
            return;
        }
        await onCreateRun(form);
    }

    return (
        <section className="page-stack">
            <article className="surface-card modal-form-card">
                <div className="panel-header">
                    <div>
                        <p className="section-kicker">任务发布</p>
                        <h2 className="page-section-title">创建新的多智能体运行任务</h2>
                        <p className="page-section-copy">只填目标与交付物，主脑会自动拆解并编排。</p>
                    </div>
                </div>

                <div className="intake-hint-strip">
                    <span className="soft-pill intake-hint-chip">自动规划角色与审核链路</span>
                    <span className="soft-pill intake-hint-chip">建议写清边界与验收标准</span>
                    <span className="soft-pill intake-hint-chip">并发上限 100</span>
                </div>

                <form className="form-grid intake-form-grid" onSubmit={submit}>
                    <label className="field-group">
                        <span className="field-label">任务标题</span>
                        <input
                            className="field-input"
                            disabled={isBusy}
                            onChange={(event) => setForm((current) => ({...current, title: event.target.value}))}
                            placeholder="例如：maliang swarm 正式版收口"
                            value={form.title}
                        />
                        <small className="field-help">{trimmedTitle ? "用于运行记录和时间线命名。" : "可留空，系统会自动生成标题。"}</small>
                    </label>

                    <label className="field-group">
                        <span className="field-label">最大智能体数</span>
                        <input
                            className="field-input"
                            disabled={isBusy}
                            max={100}
                            min={1}
                            onChange={(event) => setForm((current) => ({...current, maxAgents: Number(event.target.value)}))}
                            type="number"
                            value={form.maxAgents}
                        />
                        <small className="field-help">按真实复杂度填写即可。</small>
                    </label>

                    <label className="field-group span-two">
                        <span className="field-label">任务目标</span>
                        <textarea
                            className="field-input field-area"
                            disabled={isBusy}
                            onChange={(event) => setForm((current) => ({...current, mission: event.target.value}))}
                            placeholder="描述要完成什么、边界在哪、有哪些约束和必须满足的要求"
                            rows={5}
                            value={form.mission}
                        />
                        <small className={`field-help ${trimmedMission.length > 0 && trimmedMission.length < 12 ? "error" : ""}`}>
                            {trimmedMission.length > 0 && trimmedMission.length < 12 ? "当前描述偏短，建议补充关键信息。" : "这里最影响主脑拆解质量。"}
                        </small>
                    </label>

                    <label className="field-group span-two">
                        <span className="field-label">期望交付物</span>
                        <textarea
                            className="field-input field-area"
                            disabled={isBusy}
                            onChange={(event) => setForm((current) => ({...current, deliverable: event.target.value}))}
                            placeholder="写清最终要交付什么，例如程序、后台、文档、页面或配置中心"
                            rows={3}
                            value={form.deliverable}
                        />
                        <small className={`field-help ${trimmedDeliverable.length > 0 && trimmedDeliverable.length < 6 ? "error" : ""}`}>
                            {trimmedDeliverable.length > 0 && trimmedDeliverable.length < 6 ? "交付说明过短，建议补充验收对象。" : "交付物越具体，后续审核越容易对齐。"}
                        </small>
                    </label>

                    <label className="field-group">
                        <span className="field-label">任务类型</span>
                        <select
                            className="field-input"
                            disabled={isBusy}
                            onChange={(event) => setForm((current) => ({...current, taskType: event.target.value}))}
                            value={form.taskType}
                        >
                            <option value="engineering">工程研发</option>
                            <option value="product">产品设计</option>
                            <option value="research">研究分析</option>
                            <option value="content">内容生产</option>
                        </select>
                    </label>

                    <label className="field-group">
                        <span className="field-label">优先级</span>
                        <select
                            className="field-input"
                            disabled={isBusy}
                            onChange={(event) => setForm((current) => ({...current, priority: event.target.value}))}
                            value={form.priority}
                        >
                            <option value="critical">紧急</option>
                            <option value="high">高</option>
                            <option value="medium">中</option>
                            <option value="low">低</option>
                        </select>
                    </label>

                    <div className="toggle-row span-two">
                        <label className="toggle-card">
                            <input checked={form.requiresQA} disabled={isBusy} onChange={(event) => setForm((current) => ({...current, requiresQA: event.target.checked}))} type="checkbox"/>
                            <span>质量门禁</span>
                        </label>
                        <label className="toggle-card">
                            <input checked={form.requiresSecurity} disabled={isBusy} onChange={(event) => setForm((current) => ({...current, requiresSecurity: event.target.checked}))} type="checkbox"/>
                            <span>安全审查</span>
                        </label>
                        <label className="toggle-card">
                            <input checked={form.requiresHumanApproval} disabled={isBusy} onChange={(event) => setForm((current) => ({...current, requiresHumanApproval: event.target.checked}))} type="checkbox"/>
                            <span>人工审批</span>
                        </label>
                    </div>

                    <div className="form-status-bar span-two">
                        {validation.length > 0 ? (
                            <ul className="summary-bullet-list">
                                {validation.map((issue) => (
                                    <li key={issue} className={issue.includes("建议") ? "soft" : "error"}>
                                        {issue}
                                    </li>
                                ))}
                            </ul>
                        ) : (
                            <p className="field-help">表单信息已满足发布条件，可以交给主脑开始编排。</p>
                        )}
                    </div>

                    <div className="form-actions span-two">
                        <button className="primary-button" disabled={!canSubmit} type="submit">
                            {isBusy ? busyLabel : "发布任务"}
                        </button>
                    </div>
                </form>
            </article>
        </section>
    );
}
