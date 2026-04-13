package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"maliangswarm/internal/domain"

	_ "modernc.org/sqlite"
)

// ErrNoSnapshot indicates that the runtime store is empty.
var ErrNoSnapshot = errors.New("no runtime snapshot found")

// Store persists runtime state in SQLite.
type Store struct {
	db *sql.DB
}

// Close releases the underlying SQLite connection.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// NewSQLiteStore opens or creates a SQLite database and initializes its schema.
func NewSQLiteStore(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	store := &Store{db: db}
	if err := store.init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) init(ctx context.Context) error {
	statements := []string{
		`PRAGMA journal_mode = WAL;`,
		`CREATE TABLE IF NOT EXISTS runtime_settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			concurrency_limit INTEGER NOT NULL,
			approval_policy TEXT NOT NULL,
			theme TEXT NOT NULL,
			budget_mode TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS model_profiles (
			tier TEXT PRIMARY KEY,
			binding TEXT NOT NULL,
			responsibility TEXT NOT NULL,
			sort_order INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS ai_providers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			format TEXT NOT NULL,
			base_url TEXT NOT NULL,
			api_path TEXT NOT NULL,
			api_version TEXT NOT NULL,
			api_key TEXT NOT NULL,
			default_model TEXT NOT NULL,
			planner_model TEXT NOT NULL,
			worker_model TEXT NOT NULL,
			reviewer_model TEXT NOT NULL,
			headers_json TEXT NOT NULL,
			notes TEXT NOT NULL,
			enabled INTEGER NOT NULL,
			is_primary INTEGER NOT NULL,
			sort_order INTEGER NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS runs (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			mission TEXT NOT NULL DEFAULT '',
			deliverable TEXT NOT NULL DEFAULT '',
			priority TEXT NOT NULL DEFAULT '',
			task_type TEXT NOT NULL DEFAULT '',
			subtitle TEXT NOT NULL,
			workspace_name TEXT NOT NULL,
			active_template TEXT NOT NULL,
			current_stage TEXT NOT NULL,
			atomic_task_policy TEXT NOT NULL,
			review_mode TEXT NOT NULL,
			planner_source TEXT NOT NULL DEFAULT '',
			suggested_agent_id TEXT NOT NULL,
			step INTEGER NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS workflow_gates (
			id TEXT PRIMARY KEY,
			run_id TEXT NOT NULL,
			label TEXT NOT NULL,
			status TEXT NOT NULL,
			sort_order INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS scene_zones (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			purpose TEXT NOT NULL,
			accent TEXT NOT NULL,
			x INTEGER NOT NULL,
			y INTEGER NOT NULL,
			w INTEGER NOT NULL,
			h INTEGER NOT NULL,
			sort_order INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			run_id TEXT NOT NULL,
			title TEXT NOT NULL,
			stage TEXT NOT NULL,
			owner_agent_id TEXT NOT NULL,
			owner_role TEXT NOT NULL,
			artifact_name TEXT NOT NULL DEFAULT '',
			artifact_path TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			risk_level TEXT NOT NULL,
			detail TEXT NOT NULL,
			output_summary TEXT NOT NULL DEFAULT '',
			progress INTEGER NOT NULL,
			review_required INTEGER NOT NULL,
			dependencies INTEGER NOT NULL,
			sort_order INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			run_id TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL,
			team TEXT NOT NULL,
			desk_label TEXT NOT NULL,
			model_tier TEXT NOT NULL,
			status TEXT NOT NULL,
			computer_state TEXT NOT NULL,
			current_task TEXT NOT NULL,
			detail TEXT NOT NULL,
			artifact TEXT NOT NULL,
			next_target TEXT NOT NULL,
			last_action TEXT NOT NULL DEFAULT '',
			last_output TEXT NOT NULL DEFAULT '',
			last_handoff TEXT NOT NULL DEFAULT '',
			last_receiver TEXT NOT NULL DEFAULT '',
			last_artifact_path TEXT NOT NULL DEFAULT '',
			last_update TEXT NOT NULL,
			risk_level TEXT NOT NULL,
			progress INTEGER NOT NULL,
			queue_depth INTEGER NOT NULL,
			x INTEGER NOT NULL,
			y INTEGER NOT NULL,
			highlights_json TEXT NOT NULL,
			sort_order INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS handoffs (
			id TEXT PRIMARY KEY,
			run_id TEXT NOT NULL,
			artifact_name TEXT NOT NULL,
			from_agent_id TEXT NOT NULL,
			to_agent_id TEXT NOT NULL,
			status TEXT NOT NULL,
			sort_order INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS timeline_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id TEXT NOT NULL,
			time_label TEXT NOT NULL,
			kind TEXT NOT NULL,
			title TEXT NOT NULL,
			detail TEXT NOT NULL,
			sort_order INTEGER NOT NULL
		);`,
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	migrations := []struct {
		table      string
		columnName string
		columnDef  string
	}{
		{table: "runs", columnName: "mission", columnDef: "mission TEXT NOT NULL DEFAULT ''"},
		{table: "runs", columnName: "deliverable", columnDef: "deliverable TEXT NOT NULL DEFAULT ''"},
		{table: "runs", columnName: "priority", columnDef: "priority TEXT NOT NULL DEFAULT ''"},
		{table: "runs", columnName: "task_type", columnDef: "task_type TEXT NOT NULL DEFAULT ''"},
		{table: "runs", columnName: "planner_source", columnDef: "planner_source TEXT NOT NULL DEFAULT ''"},
		{table: "tasks", columnName: "artifact_name", columnDef: "artifact_name TEXT NOT NULL DEFAULT ''"},
		{table: "tasks", columnName: "artifact_path", columnDef: "artifact_path TEXT NOT NULL DEFAULT ''"},
		{table: "tasks", columnName: "output_summary", columnDef: "output_summary TEXT NOT NULL DEFAULT ''"},
		{table: "agents", columnName: "last_action", columnDef: "last_action TEXT NOT NULL DEFAULT ''"},
		{table: "agents", columnName: "last_output", columnDef: "last_output TEXT NOT NULL DEFAULT ''"},
		{table: "agents", columnName: "last_handoff", columnDef: "last_handoff TEXT NOT NULL DEFAULT ''"},
		{table: "agents", columnName: "last_receiver", columnDef: "last_receiver TEXT NOT NULL DEFAULT ''"},
		{table: "agents", columnName: "last_artifact_path", columnDef: "last_artifact_path TEXT NOT NULL DEFAULT ''"},
	}

	for _, migration := range migrations {
		if err := s.ensureColumn(ctx, migration.table, migration.columnName, migration.columnDef); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) ensureColumn(ctx context.Context, table string, columnName string, columnDef string) error {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s);`, table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if strings.EqualFold(name, columnName) {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s;`, table, columnDef))
	return err
}

// LoadSnapshot returns the current persisted runtime snapshot.
func (s *Store) LoadSnapshot(ctx context.Context) (domain.Snapshot, error) {
	var snapshot domain.Snapshot

	if err := s.db.QueryRowContext(ctx, `
		SELECT concurrency_limit, approval_policy, theme, budget_mode
		FROM runtime_settings
		WHERE id = 1
	`).Scan(
		&snapshot.Settings.ConcurrencyLimit,
		&snapshot.Settings.ApprovalPolicy,
		&snapshot.Settings.Theme,
		&snapshot.Settings.BudgetMode,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Snapshot{}, ErrNoSnapshot
		}
		return domain.Snapshot{}, err
	}

	if err := s.db.QueryRowContext(ctx, `
		SELECT id, title, mission, deliverable, priority, task_type, subtitle, workspace_name, active_template, current_stage,
		       atomic_task_policy, review_mode, planner_source, suggested_agent_id, step
		FROM runs
		LIMIT 1
	`).Scan(
		&snapshot.Run.ID,
		&snapshot.Run.Title,
		&snapshot.Run.Mission,
		&snapshot.Run.Deliverable,
		&snapshot.Run.Priority,
		&snapshot.Run.TaskType,
		&snapshot.Run.Subtitle,
		&snapshot.Run.WorkspaceName,
		&snapshot.Run.ActiveTemplate,
		&snapshot.Run.CurrentStage,
		&snapshot.Run.AtomicTaskPolicy,
		&snapshot.Run.ReviewMode,
		&snapshot.Run.PlannerSource,
		&snapshot.Run.SuggestedAgentID,
		&snapshot.Run.Step,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Snapshot{}, ErrNoSnapshot
		}
		return domain.Snapshot{}, err
	}

	var err error
	snapshot.ModelProfiles, err = s.loadModelProfiles(ctx)
	if err != nil {
		return domain.Snapshot{}, err
	}
	snapshot.AIProviders, err = s.loadAIProviders(ctx)
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot.Zones, err = s.loadZones(ctx)
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot.Gates, err = s.loadGates(ctx, snapshot.Run.ID)
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot.Tasks, err = s.loadTasks(ctx, snapshot.Run.ID)
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot.Agents, err = s.loadAgents(ctx, snapshot.Run.ID)
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot.Handoffs, err = s.loadHandoffs(ctx, snapshot.Run.ID)
	if err != nil {
		return domain.Snapshot{}, err
	}

	snapshot.Timeline, err = s.loadTimeline(ctx, snapshot.Run.ID)
	if err != nil {
		return domain.Snapshot{}, err
	}

	return snapshot, nil
}

// SaveSnapshot replaces the current persisted runtime snapshot.
func (s *Store) SaveSnapshot(ctx context.Context, snapshot domain.Snapshot) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC().Format(time.RFC3339)

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO runtime_settings (id, concurrency_limit, approval_policy, theme, budget_mode, updated_at)
		VALUES (1, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			concurrency_limit = excluded.concurrency_limit,
			approval_policy = excluded.approval_policy,
			theme = excluded.theme,
			budget_mode = excluded.budget_mode,
			updated_at = excluded.updated_at
	`, snapshot.Settings.ConcurrencyLimit, snapshot.Settings.ApprovalPolicy, snapshot.Settings.Theme, snapshot.Settings.BudgetMode, now); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM model_profiles`); err != nil {
		return err
	}
	for _, profile := range snapshot.ModelProfiles {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO model_profiles (tier, binding, responsibility, sort_order)
			VALUES (?, ?, ?, ?)
		`, profile.Tier, profile.Binding, profile.Responsibility, profile.SortOrder); err != nil {
			return err
		}
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM ai_providers`); err != nil {
		return err
	}
	for _, provider := range snapshot.AIProviders {
		enabled := 0
		if provider.Enabled {
			enabled = 1
		}
		isPrimary := 0
		if provider.IsPrimary {
			isPrimary = 1
		}

		if _, err = tx.ExecContext(ctx, `
			INSERT INTO ai_providers (
				id, name, format, base_url, api_path, api_version, api_key, default_model,
				planner_model, worker_model, reviewer_model, headers_json, notes,
				enabled, is_primary, sort_order, updated_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, provider.ID, provider.Name, provider.Format, provider.BaseURL, provider.APIPath, provider.APIVersion,
			provider.APIKey, provider.DefaultModel, provider.PlannerModel, provider.WorkerModel,
			provider.ReviewerModel, provider.HeadersJSON, provider.Notes, enabled, isPrimary,
			provider.SortOrder, now); err != nil {
			return err
		}
	}

	// Persist only one active runtime snapshot. Clear previous run-scoped rows
	// so stable task and agent IDs can be reused across newly created runs.
	for _, statement := range []string{
		`DELETE FROM timeline_events`,
		`DELETE FROM handoffs`,
		`DELETE FROM agents`,
		`DELETE FROM tasks`,
		`DELETE FROM workflow_gates`,
		`DELETE FROM runs`,
	} {
		if _, err = tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO runs (
			id, title, mission, deliverable, priority, task_type, subtitle, workspace_name, active_template, current_stage,
			atomic_task_policy, review_mode, planner_source, suggested_agent_id, step, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			mission = excluded.mission,
			deliverable = excluded.deliverable,
			priority = excluded.priority,
			task_type = excluded.task_type,
			subtitle = excluded.subtitle,
			workspace_name = excluded.workspace_name,
			active_template = excluded.active_template,
			current_stage = excluded.current_stage,
			atomic_task_policy = excluded.atomic_task_policy,
			review_mode = excluded.review_mode,
			planner_source = excluded.planner_source,
			suggested_agent_id = excluded.suggested_agent_id,
			step = excluded.step,
			updated_at = excluded.updated_at
	`, snapshot.Run.ID, snapshot.Run.Title, snapshot.Run.Mission, snapshot.Run.Deliverable, snapshot.Run.Priority, snapshot.Run.TaskType,
		snapshot.Run.Subtitle, snapshot.Run.WorkspaceName, snapshot.Run.ActiveTemplate, snapshot.Run.CurrentStage,
		snapshot.Run.AtomicTaskPolicy, snapshot.Run.ReviewMode, snapshot.Run.PlannerSource, snapshot.Run.SuggestedAgentID,
		snapshot.Run.Step, now); err != nil {
		return err
	}

	for _, gate := range snapshot.Gates {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO workflow_gates (id, run_id, label, status, sort_order)
			VALUES (?, ?, ?, ?, ?)
		`, gate.ID, snapshot.Run.ID, gate.Label, gate.Status, gate.SortOrder); err != nil {
			return err
		}
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM scene_zones`); err != nil {
		return err
	}
	for _, zone := range snapshot.Zones {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO scene_zones (id, name, purpose, accent, x, y, w, h, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, zone.ID, zone.Name, zone.Purpose, zone.Accent, zone.X, zone.Y, zone.W, zone.H, zone.SortOrder); err != nil {
			return err
		}
	}

	for _, task := range snapshot.Tasks {
		reviewRequired := 0
		if task.ReviewRequired {
			reviewRequired = 1
		}

		if _, err = tx.ExecContext(ctx, `
			INSERT INTO tasks (
				id, run_id, title, stage, owner_agent_id, owner_role, artifact_name, artifact_path, status,
				risk_level, detail, output_summary, progress, review_required, dependencies, sort_order
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, task.ID, snapshot.Run.ID, task.Title, task.Stage, task.OwnerAgentID, task.OwnerRole, task.ArtifactName, task.ArtifactPath, task.Status,
			task.RiskLevel, task.Detail, task.OutputSummary, task.Progress, reviewRequired, task.Dependencies, task.SortOrder); err != nil {
			return err
		}
	}

	for _, agent := range snapshot.Agents {
		highlightsJSON, marshalErr := json.Marshal(agent.Highlights)
		if marshalErr != nil {
			return marshalErr
		}

		if _, err = tx.ExecContext(ctx, `
			INSERT INTO agents (
				id, run_id, name, role, team, desk_label, model_tier, status, computer_state,
				current_task, detail, artifact, next_target, last_action, last_output, last_handoff,
				last_receiver, last_artifact_path, last_update, risk_level,
				progress, queue_depth, x, y, highlights_json, sort_order
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, agent.ID, snapshot.Run.ID, agent.Name, agent.Role, agent.Team, agent.DeskLabel, agent.ModelTier,
			agent.Status, agent.ComputerState, agent.CurrentTask, agent.Detail, agent.Artifact, agent.NextTarget,
			agent.LastAction, agent.LastOutput, agent.LastHandoff, agent.LastReceiver, agent.LastArtifactPath,
			agent.LastUpdate, agent.RiskLevel, agent.Progress, agent.QueueDepth, agent.X, agent.Y,
			string(highlightsJSON), agent.SortOrder); err != nil {
			return err
		}
	}

	for _, handoff := range snapshot.Handoffs {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO handoffs (id, run_id, artifact_name, from_agent_id, to_agent_id, status, sort_order)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, handoff.ID, snapshot.Run.ID, handoff.ArtifactName, handoff.FromAgentID, handoff.ToAgentID, handoff.Status, handoff.SortOrder); err != nil {
			return err
		}
	}

	for _, event := range snapshot.Timeline {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO timeline_events (run_id, time_label, kind, title, detail, sort_order)
			VALUES (?, ?, ?, ?, ?, ?)
		`, snapshot.Run.ID, event.TimeLabel, event.Kind, event.Title, event.Detail, event.SortOrder); err != nil {
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (s *Store) loadModelProfiles(ctx context.Context) ([]domain.ModelProfile, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT sort_order, tier, binding, responsibility
		FROM model_profiles
		ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []domain.ModelProfile
	for rows.Next() {
		var profile domain.ModelProfile
		if err := rows.Scan(&profile.SortOrder, &profile.Tier, &profile.Binding, &profile.Responsibility); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}

func (s *Store) loadAIProviders(ctx context.Context) ([]domain.AIProviderConfig, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, format, base_url, api_path, api_version, api_key, default_model,
		       planner_model, worker_model, reviewer_model, headers_json, notes,
		       enabled, is_primary, sort_order
		FROM ai_providers
		ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []domain.AIProviderConfig
	for rows.Next() {
		var provider domain.AIProviderConfig
		var enabled int
		var isPrimary int
		if err := rows.Scan(
			&provider.ID,
			&provider.Name,
			&provider.Format,
			&provider.BaseURL,
			&provider.APIPath,
			&provider.APIVersion,
			&provider.APIKey,
			&provider.DefaultModel,
			&provider.PlannerModel,
			&provider.WorkerModel,
			&provider.ReviewerModel,
			&provider.HeadersJSON,
			&provider.Notes,
			&enabled,
			&isPrimary,
			&provider.SortOrder,
		); err != nil {
			return nil, err
		}
		provider.Enabled = enabled == 1
		provider.IsPrimary = isPrimary == 1
		providers = append(providers, provider)
	}

	return providers, rows.Err()
}

func (s *Store) loadZones(ctx context.Context) ([]domain.SceneZone, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, purpose, accent, x, y, w, h, sort_order
		FROM scene_zones
		ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var zones []domain.SceneZone
	for rows.Next() {
		var zone domain.SceneZone
		if err := rows.Scan(&zone.ID, &zone.Name, &zone.Purpose, &zone.Accent, &zone.X, &zone.Y, &zone.W, &zone.H, &zone.SortOrder); err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}

	return zones, rows.Err()
}

func (s *Store) loadGates(ctx context.Context, runID string) ([]domain.WorkflowGate, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, label, status, sort_order
		FROM workflow_gates
		WHERE run_id = ?
		ORDER BY sort_order ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gates []domain.WorkflowGate
	for rows.Next() {
		var gate domain.WorkflowGate
		if err := rows.Scan(&gate.ID, &gate.Label, &gate.Status, &gate.SortOrder); err != nil {
			return nil, err
		}
		gates = append(gates, gate)
	}

	return gates, rows.Err()
}

func (s *Store) loadTasks(ctx context.Context, runID string) ([]domain.Task, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, stage, owner_agent_id, owner_role, artifact_name, artifact_path, status, risk_level, detail,
		       output_summary, progress, review_required, dependencies, sort_order
		FROM tasks
		WHERE run_id = ?
		ORDER BY sort_order ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		var reviewRequired int
		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Stage,
			&task.OwnerAgentID,
			&task.OwnerRole,
			&task.ArtifactName,
			&task.ArtifactPath,
			&task.Status,
			&task.RiskLevel,
			&task.Detail,
			&task.OutputSummary,
			&task.Progress,
			&reviewRequired,
			&task.Dependencies,
			&task.SortOrder,
		); err != nil {
			return nil, err
		}
		task.ReviewRequired = reviewRequired == 1
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (s *Store) loadAgents(ctx context.Context, runID string) ([]domain.Agent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, role, team, desk_label, model_tier, status, computer_state,
		       current_task, detail, artifact, next_target, last_action, last_output, last_handoff,
		       last_receiver, last_artifact_path, last_update, risk_level,
		       progress, queue_depth, x, y, highlights_json, sort_order
		FROM agents
		WHERE run_id = ?
		ORDER BY sort_order ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []domain.Agent
	for rows.Next() {
		var agent domain.Agent
		var highlightsJSON string
		if err := rows.Scan(
			&agent.ID,
			&agent.Name,
			&agent.Role,
			&agent.Team,
			&agent.DeskLabel,
			&agent.ModelTier,
			&agent.Status,
			&agent.ComputerState,
			&agent.CurrentTask,
			&agent.Detail,
			&agent.Artifact,
			&agent.NextTarget,
			&agent.LastAction,
			&agent.LastOutput,
			&agent.LastHandoff,
			&agent.LastReceiver,
			&agent.LastArtifactPath,
			&agent.LastUpdate,
			&agent.RiskLevel,
			&agent.Progress,
			&agent.QueueDepth,
			&agent.X,
			&agent.Y,
			&highlightsJSON,
			&agent.SortOrder,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(highlightsJSON), &agent.Highlights); err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	return agents, rows.Err()
}

func (s *Store) loadHandoffs(ctx context.Context, runID string) ([]domain.Handoff, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, artifact_name, from_agent_id, to_agent_id, status, sort_order
		FROM handoffs
		WHERE run_id = ?
		ORDER BY sort_order ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var handoffs []domain.Handoff
	for rows.Next() {
		var handoff domain.Handoff
		if err := rows.Scan(&handoff.ID, &handoff.ArtifactName, &handoff.FromAgentID, &handoff.ToAgentID, &handoff.Status, &handoff.SortOrder); err != nil {
			return nil, err
		}
		handoffs = append(handoffs, handoff)
	}

	return handoffs, rows.Err()
}

func (s *Store) loadTimeline(ctx context.Context, runID string) ([]domain.TimelineEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT time_label, kind, title, detail, sort_order
		FROM timeline_events
		WHERE run_id = ?
		ORDER BY sort_order ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timeline []domain.TimelineEvent
	for rows.Next() {
		var event domain.TimelineEvent
		if err := rows.Scan(&event.TimeLabel, &event.Kind, &event.Title, &event.Detail, &event.SortOrder); err != nil {
			return nil, err
		}
		timeline = append(timeline, event)
	}

	return timeline, rows.Err()
}
