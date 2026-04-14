import {useMemo, useState, type FormEvent} from "react";
import type {AIProvider, AIProviderInput} from "../types";
import {formatProviderFormat, localizeText} from "../lib/format";

interface AIProviderPanelProps {
    busyLabel: string;
    providers: AIProvider[];
    primaryProvider: string;
    onDeleteProvider: (id: string) => Promise<void>;
    onProbeProvider: (input: AIProviderInput) => Promise<string>;
    onUpsertProvider: (input: AIProviderInput) => Promise<void>;
    onNotify?: (tone: "success" | "warning" | "error", title: string, description?: string) => void;
}

const baseUrlPlaceholders: Record<string, string> = {
    "openai-compatible": "https://api.openai.com/v1",
    anthropic: "https://api.anthropic.com",
    gemini: "https://generativelanguage.googleapis.com",
    "azure-openai": "https://YOUR-RESOURCE.openai.azure.com",
    ollama: "http://127.0.0.1:11434",
    "custom-http": "https://your-gateway.example.com",
};

const pathPlaceholders: Record<string, string> = {
    "openai-compatible": "/chat/completions",
    anthropic: "/v1/messages",
    gemini: "/v1beta/models",
    "azure-openai": "/openai/responses",
    ollama: "/api/chat",
    "custom-http": "/v1/dispatch",
};

const modelPlaceholders: Record<string, string> = {
    "openai-compatible": "gpt-5.4",
    anthropic: "claude-sonnet-4-5",
    gemini: "gemini-2.5-pro",
    "azure-openai": "gpt-4.1",
    ollama: "qwen2.5-coder:14b",
    "custom-http": "custom-model",
};

const inheritedModelPlaceholder = "留空则沿用默认模型";

function emptyProviderInput(makePrimary: boolean): AIProviderInput {
    return {
        id: "",
        name: "",
        format: "openai-compatible",
        baseUrl: "",
        apiPath: "",
        apiVersion: "",
        apiKey: "",
        defaultModel: "",
        plannerModel: "",
        workerModel: "",
        reviewerModel: "",
        headersJson: "",
        notes: "",
        enabled: true,
        isPrimary: makePrimary,
    };
}

export function AIProviderPanel({
    busyLabel,
    providers,
    primaryProvider,
    onDeleteProvider,
    onProbeProvider,
    onUpsertProvider,
    onNotify,
}: AIProviderPanelProps) {
    const [form, setForm] = useState<AIProviderInput>(() => emptyProviderInput(true));
    const [editingName, setEditingName] = useState("");
    const [probeReport, setProbeReport] = useState("");
    const isBusy = busyLabel.length > 0;

    const liveCount = useMemo(() => providers.filter((provider) => provider.enabled).length, [providers]);
    const localizedProviders = useMemo(() => providers.map((provider) => ({
        ...provider,
        name: localizeText(provider.name),
        notes: localizeText(provider.notes),
    })), [providers]);

    const requiredChecks = useMemo(() => {
        const checks = [
            {label: "接口名称", ok: form.name.trim().length > 0},
            {label: "基础地址", ok: form.baseUrl.trim().length > 0},
            {label: "默认模型", ok: form.defaultModel.trim().length > 0},
        ];

        if (form.baseUrl.trim() && !/^https?:\/\//i.test(form.baseUrl.trim())) {
            checks.push({label: "基础地址需以 http:// 或 https:// 开头", ok: false});
        }

        return checks;
    }, [form.baseUrl, form.defaultModel, form.name]);

    const blockingChecks = requiredChecks.filter((item) => !item.ok);
    const canSave = !isBusy && blockingChecks.length === 0;
    const canProbe = !isBusy && form.baseUrl.trim().length > 0;
    const primarySummary = primaryProvider.trim() ? "已绑定主路由" : "未设置主路由";

    async function submit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        if (!canSave) {
            onNotify?.("warning", "无法保存", "请先补全必填字段后再保存。");
            return;
        }

        try {
            const finalForm = {
                ...form,
                apiPath: form.apiPath || pathPlaceholders[form.format] || "",
            };
            await onUpsertProvider(finalForm);
            setForm(emptyProviderInput(providers.length === 0));
            setEditingName("");
            setProbeReport("");
            onNotify?.("success", "保存成功", "接口配置已更新。");
        } catch {
            onNotify?.("error", "保存失败", "接口配置未能保存，请稍后重试。");
        }
    }

    function editProvider(provider: AIProvider) {
        setForm({
            id: provider.id,
            name: localizeText(provider.name),
            format: provider.format,
            baseUrl: provider.baseUrl,
            apiPath: provider.apiPath,
            apiVersion: provider.apiVersion,
            apiKey: "",
            defaultModel: provider.defaultModel,
            plannerModel: provider.plannerModel,
            workerModel: provider.workerModel,
            reviewerModel: provider.reviewerModel,
            headersJson: provider.headersJson || "",
            notes: localizeText(provider.notes),
            enabled: provider.enabled,
            isPrimary: provider.isPrimary,
        });
        setEditingName(localizeText(provider.name));
        setProbeReport("");
    }

    function resetForm() {
        setForm(emptyProviderInput(providers.length === 0));
        setEditingName("");
        setProbeReport("");
    }

    async function removeProvider(id: string) {
        if (!window.confirm("确定要删除这个接口配置吗？")) {
            return;
        }

        await onDeleteProvider(id);
        if (form.id === id) {
            resetForm();
        }
    }

    async function probeCurrentProvider() {
        if (!canProbe) {
            onNotify?.("warning", "无法测试", "请先填写基础地址。");
            return;
        }

        try {
            const finalForm = {
                ...form,
                apiPath: form.apiPath || pathPlaceholders[form.format] || "",
            };
            const report = await onProbeProvider(finalForm);
            setProbeReport(report);

            const normalized = report.toLowerCase();
            if (normalized.includes("总结果：成功") || normalized.includes("总结果: 成功")) {
                onNotify?.("success", "测试通过", "所有已配置链路都返回了可用结果。");
                return;
            }

            if (normalized.includes("总结果：部分成功") || normalized.includes("总结果: 部分成功")) {
                onNotify?.("warning", "测试部分成功", "部分链路返回了可用结果，请检查失败项。");
                return;
            }

            onNotify?.("error", "测试失败", "当前路由未通过测试，请检查地址、路径和模型配置。");
        } catch {
            onNotify?.("error", "测试失败", "测试请求未完成，请稍后重试。");
        }
    }

    const baseUrlPlaceholder = baseUrlPlaceholders[form.format];
    const modelPlaceholder = modelPlaceholders[form.format];
    const isOpenAICompatible = form.format === "openai-compatible";

    return (
        <section className="page-stack">
            <div className="split-layout provider-layout provider-layout-compact">
                <article className="surface-card provider-sidebar-card modal-form-card">
                    <div className="panel-header provider-sidebar-head">
                        <div>
                            <p className="section-kicker">接口注册表</p>
                            <h2 className="page-section-title">已配置接口</h2>
                            <p className="page-section-copy">这里只看名称和状态。</p>
                        </div>

                        {editingName ? (
                            <button className="ghost-button" disabled={isBusy} onClick={resetForm} type="button">
                                新建接口
                            </button>
                        ) : null}
                    </div>

                    <div className="provider-overview-row provider-overview-row-simple">
                        <span className="soft-pill">{primarySummary}</span>
                        <span className="soft-pill">{`在线 ${liveCount} / ${providers.length}`}</span>
                    </div>

                    {localizedProviders.length === 0 ? (
                        <div className="empty-state-card slim">
                            <strong>还没有接口</strong>
                            <p>先在右侧新增一条路由，再测试连通性。</p>
                        </div>
                    ) : (
                        <div className="provider-list provider-list-compact">
                            {localizedProviders.map((provider) => (
                                <article key={provider.id} className={`provider-list-card provider-compact-card ${provider.isPrimary ? "primary" : ""}`}>
                                    <div className="provider-list-head">
                                        <div>
                                            <strong>{provider.name}</strong>
                                            <span>{formatProviderFormat(provider.format)}</span>
                                        </div>
                                    </div>

                                    <div className="provider-compact-tags">
                                        {provider.isPrimary ? <span className="soft-pill">主路由</span> : null}
                                        <span className={`provider-status ${provider.enabled ? "online" : "offline"}`}>
                                            {provider.enabled ? "在线" : "停用"}
                                        </span>
                                    </div>

                                    <div className="inline-actions">
                                        <button className="text-button" disabled={isBusy} onClick={() => editProvider(provider)} type="button">
                                            编辑
                                        </button>
                                        <button className="text-button danger" disabled={isBusy} onClick={() => removeProvider(provider.id)} type="button">
                                            删除
                                        </button>
                                    </div>
                                </article>
                            ))}
                        </div>
                    )}
                </article>

                <article className="surface-card provider-editor-card modal-form-card">
                    <div className="panel-header provider-editor-head">
                        <div>
                            <p className="section-kicker">接口编辑</p>
                            <h2 className="page-section-title">{editingName ? `编辑接口：${editingName}` : "新建接口"}</h2>
                            <p className="page-section-copy">左侧选中接口后，在这里完成编辑。</p>
                        </div>
                    </div>

                    <form className="provider-editor-form" onSubmit={submit}>
                        <section className="provider-form-section">
                            <div className="provider-section-head">
                                <h3>基础信息</h3>
                                <p>名称、格式和请求地址为必填项。</p>
                            </div>

                            <div className="form-grid compact provider-form-grid">
                                <label className="field-group">
                                    <span className="field-label">接口名称</span>
                                    <input
                                        className="field-input"
                                        disabled={isBusy}
                                        onChange={(event) => setForm((current) => ({...current, name: event.target.value}))}
                                        placeholder="例如：生产环境主路由"
                                        value={form.name}
                                    />
                                </label>

                                <label className="field-group">
                                    <span className="field-label">接口格式</span>
                                    <select
                                        className="field-input"
                                        disabled={isBusy}
                                        onChange={(event) => setForm((current) => ({...current, format: event.target.value}))}
                                        value={form.format}
                                    >
                                        <option value="openai-compatible">兼容格式</option>
                                        <option value="anthropic">Anthropic</option>
                                        <option value="gemini">Gemini</option>
                                        <option value="azure-openai">Azure OpenAI</option>
                                        <option value="ollama">Ollama</option>
                                        <option value="custom-http">自定义接口</option>
                                    </select>
                                </label>

                                <label className="field-group span-two">
                                    <span className="field-label">基础地址</span>
                                    <input
                                        className="field-input"
                                        disabled={isBusy}
                                        onChange={(event) => setForm((current) => ({...current, baseUrl: event.target.value}))}
                                        placeholder={baseUrlPlaceholder}
                                        value={form.baseUrl}
                                    />
                                    {isOpenAICompatible ? (
                                        <small className="field-help">OpenAI 兼容格式请直接填到版本层级，例如 `https://openclawroot.com/v1`；大多数兼容网关推荐把请求路径设为 `/chat/completions`。</small>
                                    ) : null}
                                </label>

                                <label className="field-group">
                                    <span className="field-label">请求路径</span>
                                    <input
                                        className="field-input"
                                        disabled={isBusy}
                                        onChange={(event) => setForm((current) => ({...current, apiPath: event.target.value}))}
                                        placeholder={pathPlaceholders[form.format]}
                                        value={form.apiPath}
                                    />
                                    <small className="field-help">兼容格式默认建议使用 `/chat/completions`；只有明确支持时再改成别的路径。</small>
                                </label>

                                {!isOpenAICompatible ? (
                                    <label className="field-group">
                                        <span className="field-label">接口版本</span>
                                        <input
                                            className="field-input"
                                            disabled={isBusy}
                                            onChange={(event) => setForm((current) => ({...current, apiVersion: event.target.value}))}
                                            placeholder="可选"
                                            value={form.apiVersion}
                                        />
                                    </label>
                                ) : null}
                            </div>
                        </section>

                        <section className="provider-form-section">
                            <div className="provider-section-head">
                                <h3>模型绑定</h3>
                                <p>为默认、主脑、执行和审核分别指定模型。</p>
                            </div>

                            <div className="form-grid compact provider-form-grid">
                                <label className="field-group">
                                    <span className="field-label">默认模型</span>
                                    <input
                                        className="field-input"
                                        disabled={isBusy}
                                        onChange={(event) => setForm((current) => ({...current, defaultModel: event.target.value}))}
                                        placeholder={modelPlaceholder}
                                        value={form.defaultModel}
                                    />
                                </label>

                                <label className="field-group">
                                    <span className="field-label">主脑模型</span>
                                    <input
                                        className="field-input"
                                        disabled={isBusy}
                                        onChange={(event) => setForm((current) => ({...current, plannerModel: event.target.value}))}
                                        placeholder={inheritedModelPlaceholder}
                                        value={form.plannerModel}
                                    />
                                </label>

                                <label className="field-group">
                                    <span className="field-label">执行模型</span>
                                    <input
                                        className="field-input"
                                        disabled={isBusy}
                                        onChange={(event) => setForm((current) => ({...current, workerModel: event.target.value}))}
                                        placeholder={inheritedModelPlaceholder}
                                        value={form.workerModel}
                                    />
                                </label>

                                <label className="field-group">
                                    <span className="field-label">审核模型</span>
                                    <input
                                        className="field-input"
                                        disabled={isBusy}
                                        onChange={(event) => setForm((current) => ({...current, reviewerModel: event.target.value}))}
                                        placeholder={inheritedModelPlaceholder}
                                        value={form.reviewerModel}
                                    />
                                </label>
                            </div>
                        </section>

                        <details className="provider-advanced-panel">
                            <summary>高级配置</summary>
                            <div className="provider-advanced-body">
                                <div className="form-grid compact provider-form-grid">
                                    <label className="field-group span-two">
                                        <span className="field-label">接口密钥</span>
                                        <input
                                            className="field-input"
                                            disabled={isBusy}
                                            onChange={(event) => setForm((current) => ({...current, apiKey: event.target.value}))}
                                            placeholder={editingName ? "留空表示保留原有密钥" : "输入当前接口密钥"}
                                            type="password"
                                            value={form.apiKey}
                                        />
                                    </label>

                                    <label className="field-group span-two">
                                        <span className="field-label">请求头 JSON</span>
                                        <textarea
                                            className="field-input field-area"
                                            disabled={isBusy}
                                            onChange={(event) => setForm((current) => ({...current, headersJson: event.target.value}))}
                                            placeholder='可选，例如 {"x-tenant":"prod"}'
                                            rows={4}
                                            value={form.headersJson}
                                        />
                                    </label>

                                    <label className="field-group span-two">
                                        <span className="field-label">备注</span>
                                        <textarea
                                            className="field-input field-area"
                                            disabled={isBusy}
                                            onChange={(event) => setForm((current) => ({...current, notes: event.target.value}))}
                                            placeholder="可选"
                                            rows={3}
                                            value={form.notes}
                                        />
                                    </label>
                                </div>

                                <div className="toggle-row provider-toggle-row">
                                    <label className="toggle-card">
                                        <input checked={form.enabled} disabled={isBusy} onChange={(event) => setForm((current) => ({...current, enabled: event.target.checked}))} type="checkbox"/>
                                        <span>启用接口</span>
                                    </label>
                                    <label className="toggle-card">
                                        <input checked={form.isPrimary} disabled={isBusy} onChange={(event) => setForm((current) => ({...current, isPrimary: event.target.checked}))} type="checkbox"/>
                                        <span>设为主路由</span>
                                    </label>
                                </div>
                            </div>
                        </details>

                        <div className="provider-feedback-bar">
                            <div className="provider-feedback-copy">
                                {blockingChecks.length > 0 ? (
                                    <div className="provider-feedback-list">
                                        {blockingChecks.map((item) => (
                                            <span key={item.label} className="provider-feedback-item">
                                                {item.label}
                                            </span>
                                        ))}
                                    </div>
                                ) : (
                                    <p>当前表单已满足保存条件，可以直接保存或先测试。</p>
                                )}
                            </div>

                            <div className="form-actions provider-actions">
                                <button className="primary-button" disabled={!canSave} type="submit">
                                    {editingName ? "保存修改" : "新增接口"}
                                </button>
                                <button className="secondary-button" disabled={!canProbe} onClick={probeCurrentProvider} type="button">
                                    测试路由
                                </button>
                                <button className="ghost-button" disabled={isBusy} onClick={resetForm} type="button">
                                    清空
                                </button>
                            </div>
                        </div>
                    </form>

                    {probeReport ? (
                        <div className="result-box provider-result-box">
                            <p className="field-label">测试结果</p>
                            <pre>{localizeText(probeReport)}</pre>
                        </div>
                    ) : null}
                </article>
            </div>
        </section>
    );
}
