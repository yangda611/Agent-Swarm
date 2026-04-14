export interface DashboardState {
    hasActiveRun: boolean;
    title: string;
    subtitle: string;
    workspaceName: string;
    activeTemplate: string;
    suggestedAgentId: string;
    runStep: number;
    maxStep: number;
    metrics: MetricCard[];
    workflow: WorkflowSummary;
    settings: SettingsSummary;
    zones: SceneZone[];
    tasks: TaskState[];
    agents: AgentState[];
    handoffs: HandoffState[];
    timeline: TimelineItem[];
}

export interface MetricCard {
    label: string;
    value: string;
    accent: string;
    detail: string;
}

export interface WorkflowSummary {
    currentStage: string;
    atomicTaskPolicy: string;
    reviewMode: string;
    plannerSource: string;
    pendingGates: string[];
    completedGates: string[];
}

export interface SettingsSummary {
    concurrencyLimit: number;
    approvalPolicy: string;
    theme: string;
    budgetMode: string;
    primaryProvider: string;
    aiProviders: AIProvider[];
    modelProfiles: ModelProfile[];
}

export interface ModelProfile {
    tier: string;
    binding: string;
    responsibility: string;
}

export interface AIProvider {
    id: string;
    name: string;
    format: string;
    baseUrl: string;
    apiPath: string;
    apiVersion: string;
    defaultModel: string;
    plannerModel: string;
    workerModel: string;
    reviewerModel: string;
    headersJson: string;
    notes: string;
    enabled: boolean;
    isPrimary: boolean;
    apiKeyConfigured: boolean;
    apiKeyPreview: string;
}

export interface SceneZone {
    id: string;
    name: string;
    purpose: string;
    accent: string;
    x: number;
    y: number;
    w: number;
    h: number;
}

export interface TaskState {
    id: string;
    title: string;
    stage: string;
    ownerRole: string;
    artifactName: string;
    artifactPath: string;
    status: string;
    riskLevel: string;
    detail: string;
    outputSummary: string;
    progress: number;
    reviewRequired: boolean;
    dependencies: number;
}

export interface AgentState {
    id: string;
    name: string;
    role: string;
    team: string;
    deskLabel: string;
    modelTier: string;
    status: string;
    computerState: string;
    currentTask: string;
    detail: string;
    artifact: string;
    nextTarget: string;
    lastAction: string;
    lastOutput: string;
    lastHandoff: string;
    lastReceiver: string;
    lastArtifactPath: string;
    lastUpdate: string;
    riskLevel: string;
    progress: number;
    queueDepth: number;
    x: number;
    y: number;
    highlights: string[];
}

export interface HandoffState {
    id: string;
    artifactName: string;
    fromAgentId: string;
    toAgentId: string;
    status: string;
}

export interface TimelineItem {
    timeLabel: string;
    kind: string;
    title: string;
    detail: string;
}

export interface SettingsUpdateInput {
    concurrencyLimit: number;
    approvalPolicy: string;
    theme: string;
    budgetMode: string;
}

export interface RunCreationInput {
    title: string;
    mission: string;
    deliverable: string;
    taskType: string;
    priority: string;
    maxAgents: number;
    requiresQA: boolean;
    requiresSecurity: boolean;
    requiresHumanApproval: boolean;
}

export interface AIProviderInput {
    id: string;
    name: string;
    format: string;
    baseUrl: string;
    apiPath: string;
    apiVersion: string;
    apiKey: string;
    defaultModel: string;
    plannerModel: string;
    workerModel: string;
    reviewerModel: string;
    headersJson: string;
    notes: string;
    enabled: boolean;
    isPrimary: boolean;
}
