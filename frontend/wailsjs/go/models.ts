export namespace orchestrator {
	
	export class AIProvider {
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
	
	    static createFrom(source: any = {}) {
	        return new AIProvider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.format = source["format"];
	        this.baseUrl = source["baseUrl"];
	        this.apiPath = source["apiPath"];
	        this.apiVersion = source["apiVersion"];
	        this.defaultModel = source["defaultModel"];
	        this.plannerModel = source["plannerModel"];
	        this.workerModel = source["workerModel"];
	        this.reviewerModel = source["reviewerModel"];
	        this.headersJson = source["headersJson"];
	        this.notes = source["notes"];
	        this.enabled = source["enabled"];
	        this.isPrimary = source["isPrimary"];
	        this.apiKeyConfigured = source["apiKeyConfigured"];
	        this.apiKeyPreview = source["apiKeyPreview"];
	    }
	}
	export class AIProviderInput {
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
	
	    static createFrom(source: any = {}) {
	        return new AIProviderInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.format = source["format"];
	        this.baseUrl = source["baseUrl"];
	        this.apiPath = source["apiPath"];
	        this.apiVersion = source["apiVersion"];
	        this.apiKey = source["apiKey"];
	        this.defaultModel = source["defaultModel"];
	        this.plannerModel = source["plannerModel"];
	        this.workerModel = source["workerModel"];
	        this.reviewerModel = source["reviewerModel"];
	        this.headersJson = source["headersJson"];
	        this.notes = source["notes"];
	        this.enabled = source["enabled"];
	        this.isPrimary = source["isPrimary"];
	    }
	}
	export class AgentState {
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
	
	    static createFrom(source: any = {}) {
	        return new AgentState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.role = source["role"];
	        this.team = source["team"];
	        this.deskLabel = source["deskLabel"];
	        this.modelTier = source["modelTier"];
	        this.status = source["status"];
	        this.computerState = source["computerState"];
	        this.currentTask = source["currentTask"];
	        this.detail = source["detail"];
	        this.artifact = source["artifact"];
	        this.nextTarget = source["nextTarget"];
	        this.lastAction = source["lastAction"];
	        this.lastOutput = source["lastOutput"];
	        this.lastHandoff = source["lastHandoff"];
	        this.lastReceiver = source["lastReceiver"];
	        this.lastArtifactPath = source["lastArtifactPath"];
	        this.lastUpdate = source["lastUpdate"];
	        this.riskLevel = source["riskLevel"];
	        this.progress = source["progress"];
	        this.queueDepth = source["queueDepth"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.highlights = source["highlights"];
	    }
	}
	export class TimelineItem {
	    timeLabel: string;
	    kind: string;
	    title: string;
	    detail: string;
	
	    static createFrom(source: any = {}) {
	        return new TimelineItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timeLabel = source["timeLabel"];
	        this.kind = source["kind"];
	        this.title = source["title"];
	        this.detail = source["detail"];
	    }
	}
	export class HandoffState {
	    id: string;
	    artifactName: string;
	    fromAgentId: string;
	    toAgentId: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new HandoffState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.artifactName = source["artifactName"];
	        this.fromAgentId = source["fromAgentId"];
	        this.toAgentId = source["toAgentId"];
	        this.status = source["status"];
	    }
	}
	export class TaskState {
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
	
	    static createFrom(source: any = {}) {
	        return new TaskState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.stage = source["stage"];
	        this.ownerRole = source["ownerRole"];
	        this.artifactName = source["artifactName"];
	        this.artifactPath = source["artifactPath"];
	        this.status = source["status"];
	        this.riskLevel = source["riskLevel"];
	        this.detail = source["detail"];
	        this.outputSummary = source["outputSummary"];
	        this.progress = source["progress"];
	        this.reviewRequired = source["reviewRequired"];
	        this.dependencies = source["dependencies"];
	    }
	}
	export class SceneZone {
	    id: string;
	    name: string;
	    purpose: string;
	    accent: string;
	    x: number;
	    y: number;
	    w: number;
	    h: number;
	
	    static createFrom(source: any = {}) {
	        return new SceneZone(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.purpose = source["purpose"];
	        this.accent = source["accent"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.w = source["w"];
	        this.h = source["h"];
	    }
	}
	export class ModelProfile {
	    tier: string;
	    binding: string;
	    responsibility: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tier = source["tier"];
	        this.binding = source["binding"];
	        this.responsibility = source["responsibility"];
	    }
	}
	export class SettingsSummary {
	    concurrencyLimit: number;
	    approvalPolicy: string;
	    theme: string;
	    budgetMode: string;
	    primaryProvider: string;
	    aiProviders: AIProvider[];
	    modelProfiles: ModelProfile[];
	
	    static createFrom(source: any = {}) {
	        return new SettingsSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.concurrencyLimit = source["concurrencyLimit"];
	        this.approvalPolicy = source["approvalPolicy"];
	        this.theme = source["theme"];
	        this.budgetMode = source["budgetMode"];
	        this.primaryProvider = source["primaryProvider"];
	        this.aiProviders = this.convertValues(source["aiProviders"], AIProvider);
	        this.modelProfiles = this.convertValues(source["modelProfiles"], ModelProfile);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorkflowSummary {
	    currentStage: string;
	    atomicTaskPolicy: string;
	    reviewMode: string;
	    plannerSource: string;
	    pendingGates: string[];
	    completedGates: string[];
	
	    static createFrom(source: any = {}) {
	        return new WorkflowSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentStage = source["currentStage"];
	        this.atomicTaskPolicy = source["atomicTaskPolicy"];
	        this.reviewMode = source["reviewMode"];
	        this.plannerSource = source["plannerSource"];
	        this.pendingGates = source["pendingGates"];
	        this.completedGates = source["completedGates"];
	    }
	}
	export class MetricCard {
	    label: string;
	    value: string;
	    accent: string;
	    detail: string;
	
	    static createFrom(source: any = {}) {
	        return new MetricCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.value = source["value"];
	        this.accent = source["accent"];
	        this.detail = source["detail"];
	    }
	}
	export class DashboardState {
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
	
	    static createFrom(source: any = {}) {
	        return new DashboardState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hasActiveRun = source["hasActiveRun"];
	        this.title = source["title"];
	        this.subtitle = source["subtitle"];
	        this.workspaceName = source["workspaceName"];
	        this.activeTemplate = source["activeTemplate"];
	        this.suggestedAgentId = source["suggestedAgentId"];
	        this.runStep = source["runStep"];
	        this.maxStep = source["maxStep"];
	        this.metrics = this.convertValues(source["metrics"], MetricCard);
	        this.workflow = this.convertValues(source["workflow"], WorkflowSummary);
	        this.settings = this.convertValues(source["settings"], SettingsSummary);
	        this.zones = this.convertValues(source["zones"], SceneZone);
	        this.tasks = this.convertValues(source["tasks"], TaskState);
	        this.agents = this.convertValues(source["agents"], AgentState);
	        this.handoffs = this.convertValues(source["handoffs"], HandoffState);
	        this.timeline = this.convertValues(source["timeline"], TimelineItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class RunCreationInput {
	    title: string;
	    mission: string;
	    deliverable: string;
	    taskType: string;
	    priority: string;
	    maxAgents: number;
	    requiresQA: boolean;
	    requiresSecurity: boolean;
	    requiresHumanApproval: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RunCreationInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.mission = source["mission"];
	        this.deliverable = source["deliverable"];
	        this.taskType = source["taskType"];
	        this.priority = source["priority"];
	        this.maxAgents = source["maxAgents"];
	        this.requiresQA = source["requiresQA"];
	        this.requiresSecurity = source["requiresSecurity"];
	        this.requiresHumanApproval = source["requiresHumanApproval"];
	    }
	}
	
	export class SettingsInput {
	    concurrencyLimit: number;
	    approvalPolicy: string;
	    theme: string;
	    budgetMode: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingsInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.concurrencyLimit = source["concurrencyLimit"];
	        this.approvalPolicy = source["approvalPolicy"];
	        this.theme = source["theme"];
	        this.budgetMode = source["budgetMode"];
	    }
	}
	
	
	

}

