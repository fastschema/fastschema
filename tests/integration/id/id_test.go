package id_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	u "github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	schemaDir    = "../../../tests/integration/id/data/schemas"
	migrationDir = "../../../tests/integration/id/data/migrations"
	sqliteDSN    = "../../../tests/integration/id/data/id_test.db"
)

var idTestCases = []struct {
	name string
	fn   func(t *testing.T, client h.DBClient)
}{
	{"CustomPrimaryLifecycle", testCustomPrimaryLifecycle},
	{"PrimaryKeyFilters", testPrimaryKeyFilters},
	{"RelationFiltersAndSelects", testRelationFiltersAndSelects},
	{"SystemPrimaryLifecycle", testSystemPrimaryLifecycle},
	{"SystemRelationQueries", testSystemRelationQueries},
}

var systemSchemaTypes = []any{
	systemLab{},
	systemScientist{},
	systemExperiment{},
	systemSample{},
}

func newIDSchemaBuilder(t *testing.T) *schema.Builder {
	t.Helper()
	return u.Must(schema.NewBuilderFromDir(schemaDir, systemSchemaTypes...))
}

func TestIDMySQL(t *testing.T) {
	runIDTests(t, u.Map(h.MysqlConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := newIDSchemaBuilder(t)
		return h.NewMySQLClient(t, cfg, sb, migrationDir)
	}))
}

func TestIDPostgres(t *testing.T) {
	runIDTests(t, u.Map(h.PostgresConfigs, func(cfg h.DBConfig) h.DBClient {
		sb := newIDSchemaBuilder(t)
		return h.NewPostgresClient(t, cfg, sb, migrationDir)
	}))
}

func TestIDSQLite(t *testing.T) {
	sb := newIDSchemaBuilder(t)
	client := h.NewSQLiteClient(t, "sqlite", sqliteDSN, migrationDir, sb)
	runIDTests(t, []h.DBClient{client})
}

func runIDTests(t *testing.T, clients []h.DBClient) {
	for _, client := range clients {
		clientCopy := client
		t.Run(clientCopy.Name, func(t *testing.T) {
			for _, tc := range idTestCases {
				testCase := tc
				t.Run(testCase.name, func(t *testing.T) {
					testCase.fn(t, clientCopy)
				})
			}
		})
	}
}

func testCustomPrimaryLifecycle(t *testing.T, client h.DBClient) {
	f := seedIDGraph(t, client)
	ctx := t.Context()
	projectModel := u.Must(client.C.Model("project"))
	engineerModel := u.Must(client.C.Model("engineer"))
	deploymentModel := u.Must(client.C.Model("deployment"))
	artifactModel := u.Must(client.C.Model("artifact"))
	taskModel := u.Must(client.C.Model("task"))

	extraProjectCode := uuid.New()
	// Create an extra project to with custom PK name and type=uuid
	u.Must(projectModel.CreateFromJSON(
		ctx,
		fmt.Sprintf(
			`{"code":"%s","name":"Beta","status":"draft"}`,
			extraProjectCode,
		),
	))

	// Update existing entities using custom PK name and type=uuid
	affected := u.Must(projectModel.
		Mutation().
		Where(db.EQ("code", f.projectCode)).
		Update(ctx, entity.New().Set("status", "shipped")))
	require.Equal(t, 1, affected)

	// Verify updates
	project := u.Must(projectModel.
		Query(db.EQ("code", f.projectCode)).
		First(ctx))
	assert.Equal(t, "shipped", project.Get("status"))
	assert.Equal(t, f.projectCode, project.Get("code"))
	assert.Equal(t, f.projectCode, project.ID())

	// Update engineer using custom PK name and type=string
	affected = u.Must(engineerModel.
		Mutation().
		Where(db.EQ("handle", f.engineerHandle)).
		Update(ctx, entity.New().Set("level", "principal")))
	require.Equal(t, 1, affected)

	// Verify updates
	engineer := u.Must(engineerModel.
		Query(db.EQ("handle", f.engineerHandle)).
		First(ctx))
	assert.Equal(t, "principal", engineer.Get("level"))
	assert.Equal(t, f.engineerHandle, engineer.ID())

	// Update deployment using custom PK name and type=uuid
	affected = u.Must(deploymentModel.
		Mutation().
		Where(db.EQ("deployment_id", f.deploymentID)).
		Update(ctx, entity.New().Set("environment", "prod")))
	require.Equal(t, 1, affected)

	// Verify updates
	deployment := u.Must(deploymentModel.
		Query(db.EQ("deployment_id", f.deploymentID)).
		First(ctx))
	assert.Equal(t, "prod", deployment.Get("environment"))
	assert.Equal(t, f.deploymentID, deployment.ID())

	// Update artifact using custom PK name and type=int
	affected = u.Must(artifactModel.
		Mutation().
		Where(db.EQ("artifact_no", f.artifactNo)).
		Update(ctx, entity.New().Set("description", "rebuilt")))
	require.Equal(t, 1, affected)

	// Verify updates
	artifact := u.Must(artifactModel.
		Query(db.EQ("artifact_no", f.artifactNo)).
		First(ctx))
	assert.Equal(t, "rebuilt", artifact.Get("description"))
	assert.EqualValues(t, f.artifactNo, artifact.ID())

	// Attempt to create duplicate project with same custom PK value
	_, err := projectModel.CreateFromJSON(
		ctx,
		fmt.Sprintf(`{"code":"%s","name":"Duplicate"}`, f.projectCode),
	)
	require.Error(t, err)

	// Update task using custom PK name and type=uint64
	affected = u.Must(taskModel.
		Mutation().
		Where(db.EQ("id", f.taskID)).
		Update(ctx, entity.New().Set("status", "done")))
	require.Equal(t, 1, affected)

	// Verify updates
	task := u.Must(taskModel.Query(db.EQ("id", f.taskID)).First(ctx))
	assert.Equal(t, "done", task.Get("status"))
	assert.EqualValues(t, f.taskID, task.ID())
}

func testPrimaryKeyFilters(t *testing.T, client h.DBClient) {
	f := seedIDGraph(t, client)
	ctx := h.Ctx()
	projectModel := u.Must(client.C.Model("project"))
	engineerModel := u.Must(client.C.Model("engineer"))
	teamModel := u.Must(client.C.Model("team"))
	deploymentModel := u.Must(client.C.Model("deployment"))
	artifactModel := u.Must(client.C.Model("artifact"))
	taskModel := u.Must(client.C.Model("task"))

	// Query project using custom PK name and type=uuid
	project := u.Must(projectModel.
		Query(db.EQ("code", f.projectCode)).
		First(ctx))
	assert.Equal(t, f.projectCode, project.ID())

	// Query engineer using custom PK name and type=string
	engineer := u.Must(engineerModel.
		Query(db.EQ("handle", f.engineerHandle)).
		First(ctx))
	assert.Equal(t, f.engineerHandle, engineer.ID())

	// Query team using custom PK name and type=string
	team := u.Must(teamModel.
		Query(db.EQ("slug", f.teamSlug)).
		First(ctx))
	assert.Equal(t, f.teamSlug, team.ID())

	// Query deployment using custom PK name and type=uuid
	deployment := u.Must(deploymentModel.
		Query(db.EQ("deployment_id", f.deploymentID)).
		First(ctx))
	assert.Equal(t, f.deploymentID, deployment.ID())

	// Query artifact using custom PK name and type=int
	artifact := u.Must(artifactModel.
		Query(db.EQ("artifact_no", f.artifactNo)).
		First(ctx))
	assert.EqualValues(t, f.artifactNo, artifact.ID())

	// Filter by relation using custom PK names and type=uuid
	tasksByProject1 := u.Must(taskModel.
		Query(db.EQ("project_code", f.projectCode)).
		Get(ctx))
	require.Len(t, tasksByProject1, 1)
	assert.EqualValues(t, f.taskID, tasksByProject1[0].ID())

	tasksByProject2 := u.Must(taskModel.
		Query(db.EQ("code", f.projectCode, "project")).
		Get(ctx))
	require.Len(t, tasksByProject2, 1)
	assert.EqualValues(t, f.taskID, tasksByProject2[0].ID())

	// Filter by relation using custom PK names and type=string
	tasksByEngineer1 := u.Must(taskModel.
		Query(db.EQ("assignee_handle", f.engineerHandle)).
		Get(ctx))
	require.Len(t, tasksByEngineer1, 1)
	assert.EqualValues(t, f.taskID, tasksByEngineer1[0].ID())

	tasksByEngineer2 := u.Must(taskModel.
		Query(db.EQ("handle", f.engineerHandle, "assignee")).
		Get(ctx))
	require.Len(t, tasksByEngineer2, 1)
	assert.EqualValues(t, f.taskID, tasksByEngineer2[0].ID())

	// Filter by relation using custom PK names and type=uuid
	deployments1 := u.Must(deploymentModel.
		Query(db.EQ("project_code", f.projectCode)).
		Get(ctx))
	require.Len(t, deployments1, 1)
	assert.Equal(t, f.projectCode, deployments1[0].Get("project_code"))

	deployments2 := u.Must(deploymentModel.
		Query(db.EQ("code", f.projectCode, "project")).
		Get(ctx))
	require.Len(t, deployments2, 1)
	assert.Equal(t, f.projectCode, deployments2[0].Get("project_code"))

	// Filter by relation using custom PK names and type=int
	artifacts1 := u.Must(artifactModel.
		Query(db.EQ("project_code", f.projectCode)).
		Get(ctx))
	require.Len(t, artifacts1, 1)
	assert.Equal(t, f.projectCode, artifacts1[0].Get("project_code"))

	artifacts2 := u.Must(artifactModel.
		Query(db.EQ("code", f.projectCode, "project")).
		Get(ctx))
	require.Len(t, artifacts2, 1)
	assert.Equal(t, f.projectCode, artifacts2[0].Get("project_code"))
}

func testRelationFiltersAndSelects(t *testing.T, client h.DBClient) {
	f := seedIDGraph(t, client)
	ctx := h.Ctx()
	projectModel := u.Must(client.C.Model("project"))
	teamModel := u.Must(client.C.Model("team"))
	engineerModel := u.Must(client.C.Model("engineer"))
	taskModel := u.Must(client.C.Model("task"))

	// Create additional team and engineer to test m2m and fk relations
	secondTeamSlug := "team-" + uuid.NewString()[:6]
	u.Must(teamModel.CreateFromJSON(
		ctx,
		fmt.Sprintf(
			`{"slug":"%s","name":"API"}`,
			secondTeamSlug,
		),
	))

	// Create additional engineer to test m2m and fk relations
	secondEngineerHandle := "eng-" + uuid.NewString()[:8]
	u.Must(engineerModel.CreateFromJSON(
		ctx,
		fmt.Sprintf(
			`{"handle":"%s","name":"Bea","level":"mid"}`,
			secondEngineerHandle,
		),
	))

	// Link second engineer to second team
	u.Must(teamModel.
		Mutation().
		Where(db.EQ("slug", secondTeamSlug)).
		Update(ctx, entity.New().Set("members", []*entity.Entity{
			refEntity("handle", secondEngineerHandle),
		})))

	// Link second team to existing project
	u.Must(projectModel.
		Mutation().
		Where(db.EQ("code", f.projectCode)).
		Update(ctx, entity.New().Set("teams", []*entity.Entity{
			refEntity("slug", f.teamSlug),
			refEntity("slug", secondTeamSlug),
		})))

	// Create extra task assigned to second engineer and linked to existing project
	extraTaskID := h.IDUint64(t, u.Must(taskModel.CreateFromJSON(
		ctx,
		fmt.Sprintf(
			`{"title":"migrate","status":"doing","project":{"code":"%s"},"assignee":{"handle":"%s"}}`,
			f.projectCode, secondEngineerHandle,
		),
	)))

	// Query project using teams relation with custom PK names and type=string
	projectsForTeam := u.Must(projectModel.
		Query(db.EQ("slug", secondTeamSlug, "teams")).
		Get(ctx))
	require.Len(t, projectsForTeam, 1)
	assert.Equal(t, f.projectCode, projectsForTeam[0].Get("code"))
	assert.Equal(t, f.projectCode, projectsForTeam[0].ID())

	// Query team using projects relation with custom PK names and type=uuid
	teamsForProject := u.Must(teamModel.
		Query(db.EQ("code", f.projectCode, "projects")).
		Get(ctx))
	require.Len(t, teamsForProject, 2)
	slugs := u.Map(teamsForProject, func(e *entity.Entity) string {
		return e.Get("slug").(string)
	})
	assert.ElementsMatch(t, []string{f.teamSlug, secondTeamSlug}, slugs)

	// Query tasks using assignee relation with custom PK names and type=string
	tasksForEngineer := u.Must(taskModel.
		Query(db.EQ("handle", secondEngineerHandle, "assignee")).
		Get(ctx))
	require.Len(t, tasksForEngineer, 1)
	assert.EqualValues(t, extraTaskID, tasksForEngineer[0].ID())

	// Query engineer using tasks relation with custom PK names and type=string
	teamsByMember := u.Must(teamModel.
		Query(db.EQ("handle", secondEngineerHandle, "members")).
		Get(ctx))
	require.Len(t, teamsByMember, 1)
	assert.Equal(t, secondTeamSlug, teamsByMember[0].Get("slug"))

	// Select relations from engineer with custom PK names and type=string
	engineerTeams := u.Must(engineerModel.
		Query(db.EQ("slug", f.teamSlug, "teams")).
		Select("teams").First(ctx))
	joinedTeams, ok := engineerTeams.Get("teams").([]*entity.Entity)
	require.True(t, ok)
	require.Len(t, joinedTeams, 1)
	assert.Equal(t, f.teamSlug, joinedTeams[0].ID())

	// Select relations from project with custom PK names and type=uuid
	project := u.Must(projectModel.
		Query(db.EQ("code", f.projectCode)).
		Select("teams", "tasks", "deployments", "artifacts").
		First(ctx))

	// Verify selected teams
	selectedTeams, ok := project.Get("teams").([]*entity.Entity)
	require.True(t, ok)
	teamIDs := u.Map(selectedTeams, func(e *entity.Entity) string {
		return e.ID().(string)
	})
	assert.ElementsMatch(t, []string{f.teamSlug, secondTeamSlug}, teamIDs)

	// Verify selected tasks
	selectedTasks, ok := project.Get("tasks").([]*entity.Entity)
	require.True(t, ok)
	taskIDs := u.Map(selectedTasks, func(e *entity.Entity) uint64 {
		return h.IDUint64(t, e.ID())
	})
	assert.ElementsMatch(t, []uint64{f.taskID, extraTaskID}, taskIDs)

	// Verify selected deployments
	deployments, ok := project.Get("deployments").([]*entity.Entity)
	require.True(t, ok)
	require.Len(t, deployments, 1)
	assert.Equal(t, f.deploymentID, deployments[0].ID())

	// Verify selected artifacts
	artifacts, ok := project.Get("artifacts").([]*entity.Entity)
	require.True(t, ok)
	require.Len(t, artifacts, 1)
	assert.EqualValues(t, f.artifactNo, artifacts[0].ID())
}

func testSystemPrimaryLifecycle(t *testing.T, client h.DBClient) {
	f := seedSystemIDGraph(t, client)
	ctx := h.Ctx()
	labModel := u.Must(client.C.Model("system_lab"))
	scientistModel := u.Must(client.C.Model("system_scientist"))
	experimentModel := u.Must(client.C.Model("system_experiment"))
	sampleModel := u.Must(client.C.Model("system_sample"))

	// Create an extra lab with a unique code
	extraLabCode := uuid.New()
	u.Must(labModel.CreateFromJSON(
		ctx,
		fmt.Sprintf(`{"code":"%s","name":"Orion"}`, extraLabCode),
	))

	// Update existing entities using custom PK names and type=uuid
	affected := u.Must(labModel.
		Mutation().
		Where(db.EQ("code", f.labCode)).
		Update(ctx, entity.New().Set("focus", "gravity")))
	require.Equal(t, 1, affected)

	// Verify updates
	lab := u.Must(labModel.
		Query(db.EQ("code", f.labCode)).
		First(ctx))
	assert.Equal(t, "gravity", lab.Get("focus"))
	assert.Equal(t, f.labCode, lab.ID())

	// Update scientist using custom PK name and type=string
	affected = u.Must(scientistModel.
		Mutation().
		Where(db.EQ("handle", f.scientistHandle)).
		Update(ctx, entity.New().Set("discipline", "physics")))
	require.Equal(t, 1, affected)

	// Verify updates
	scientist := u.Must(scientistModel.
		Query(db.EQ("handle", f.scientistHandle)).
		First(ctx))
	assert.Equal(t, "physics", scientist.Get("discipline"))
	assert.Equal(t, f.scientistHandle, scientist.ID())

	// Update experiment using custom PK name and type=string
	affected = u.Must(experimentModel.
		Mutation().
		Where(db.EQ("experiment_id", f.experimentID)).
		Update(ctx, entity.New().Set("stage", "final")))
	require.Equal(t, 1, affected)

	// Verify updates
	experiment := u.Must(experimentModel.
		Query(db.EQ("experiment_id", f.experimentID)).
		First(ctx))
	assert.Equal(t, "final", experiment.Get("stage"))
	assert.Equal(t, f.experimentID, experiment.ID())

	// Update sample using custom PK name and type=uint64
	affected = u.Must(sampleModel.
		Mutation().
		Where(db.EQ("sample_no", f.sampleNo)).
		Update(ctx, entity.New().Set("status", "archived")))
	require.Equal(t, 1, affected)

	// Verify updates
	sample := u.Must(sampleModel.
		Query(db.EQ("sample_no", f.sampleNo)).
		First(ctx))
	assert.Equal(t, "archived", sample.Get("status"))
	assert.EqualValues(t, f.sampleNo, sample.ID())

	// Attempt to create duplicate lab with same custom PK value
	_, err := labModel.CreateFromJSON(
		ctx,
		fmt.Sprintf(`{"code":"%s","name":"Duplicate"}`, f.labCode),
	)
	require.Error(t, err)
}

func testSystemRelationQueries(t *testing.T, client h.DBClient) {
	f := seedSystemIDGraph(t, client)
	ctx := h.Ctx()
	labModel := u.Must(client.C.Model("system_lab"))
	scientistModel := u.Must(client.C.Model("system_scientist"))
	experimentModel := u.Must(client.C.Model("system_experiment"))
	sampleModel := u.Must(client.C.Model("system_sample"))

	// Create additional scientist to test fk and m2m relations
	secondHandle := fmt.Sprintf("sci-%s", strings.ToLower(uuid.NewString())[:8])
	u.Must(scientistModel.CreateFromJSON(
		ctx,
		fmt.Sprintf(
			`{"handle":"%s","name":"Dr. Flo","discipline":"astro"}`,
			secondHandle,
		),
	))

	// Link second scientist to existing lab
	u.Must(labModel.
		Mutation().
		Where(db.EQ("code", f.labCode)).
		Update(ctx, entity.New().Set("scientists", []*entity.Entity{
			refEntity("handle", f.scientistHandle),
			refEntity("handle", secondHandle),
		})))

	// Create additional experiment to test fk relations
	secondExperimentID := uuid.New()
	u.Must(experimentModel.CreateFromJSON(
		ctx,
		fmt.Sprintf(`{"experiment_id":"%s","title":"Comet","lab":{"code":"%s"},"scientist":{"handle":"%s"}}`, secondExperimentID, f.labCode, secondHandle),
	))

	// Create additional sample to test fk relations
	secondSampleNo := uint64(9102)
	u.Must(sampleModel.CreateFromJSON(
		ctx,
		fmt.Sprintf(`{"sample_no":%d,"label":"Trace","experiment":{"experiment_id":"%s"}}`, secondSampleNo, secondExperimentID),
	))

	// Query labs by scientist using custom PK names and type=string
	labsByScientist := u.Must(labModel.
		Query(db.EQ("handle", secondHandle, "scientists")).
		Get(ctx))
	require.Len(t, labsByScientist, 1)
	assert.Equal(t, f.labCode, labsByScientist[0].ID())

	// Query scientists by lab using custom PK names and type=uuid
	scientistsByLab := u.Must(scientistModel.
		Query(db.EQ("code", f.labCode, "labs")).
		Get(ctx))
	require.Len(t, scientistsByLab, 2)
	handles := u.Map(scientistsByLab, func(e *entity.Entity) string {
		return e.Get("handle").(string)
	})
	assert.ElementsMatch(t, []string{f.scientistHandle, secondHandle}, handles)

	// Query experiments by lab and scientist using custom PK names and type=string
	experimentsByScientist := u.Must(experimentModel.
		Query(db.EQ("handle", secondHandle, "scientist")).
		Get(ctx))
	require.Len(t, experimentsByScientist, 1)
	assert.Equal(t, secondExperimentID, experimentsByScientist[0].ID())

	// Query samples by experiment using custom PK names and type=uuid
	samplesByExperiment := u.Must(sampleModel.
		Query(db.EQ("experiment_id", secondExperimentID, "experiment")).
		Get(ctx))
	require.Len(t, samplesByExperiment, 1)
	assert.EqualValues(t, secondSampleNo, samplesByExperiment[0].ID())

	// Select relations from lab with custom PK names and type=uuid
	lab := u.Must(labModel.
		Query(db.EQ("code", f.labCode)).
		Select("scientists", "experiments").
		First(ctx))

	// Verify selected scientists
	selScientists, ok := lab.Get("scientists").([]*entity.Entity)
	require.True(t, ok)
	require.Len(t, selScientists, 2)
	selHandles := u.Map(selScientists, func(e *entity.Entity) string {
		return e.ID().(string)
	})
	assert.ElementsMatch(t, []string{f.scientistHandle, secondHandle}, selHandles)

	// Verify selected experiments
	selExperiments, ok := lab.Get("experiments").([]*entity.Entity)
	require.True(t, ok)
	require.Len(t, selExperiments, 2)
	experimentIDs := u.Map(selExperiments, func(e *entity.Entity) uuid.UUID {
		return e.ID().(uuid.UUID)
	})
	assert.ElementsMatch(t, []uuid.UUID{f.experimentID, secondExperimentID}, experimentIDs)
}
