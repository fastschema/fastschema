package id_test

import (
	"fmt"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	u "github.com/fastschema/fastschema/pkg/utils"
	h "github.com/fastschema/fastschema/tests/integration/helpers"
	"github.com/google/uuid"
)

type idFixture struct {
	projectCode    uuid.UUID
	engineerHandle string
	teamSlug       string
	deploymentID   uuid.UUID
	artifactNo     int
	taskID         uint64
}

type systemIDFixture struct {
	labCode         uuid.UUID
	scientistHandle string
	experimentID    uuid.UUID
	sampleNo        uint64
}

type systemLab struct {
	_           any                 `json:"-" fs:"name=system_lab;namespace=id_system_labs;label_field=name;primary_field=code"`
	Code        uuid.UUID           `json:"code" fs:"type=uuid;filterable;sortable"`
	Name        string              `json:"name"`
	Focus       string              `json:"focus" fs:"optional"`
	Scientists  []*systemScientist  `json:"scientists,omitempty" fs.relation:"{'type':'m2m','schema':'system_scientist','field':'labs','owner':true}"`
	Experiments []*systemExperiment `json:"experiments,omitempty" fs.relation:"{'type':'o2m','schema':'system_experiment','field':'lab','owner':true}"`
}

type systemScientist struct {
	_           any                 `json:"-" fs:"name=system_scientist;namespace=id_system_scientists;label_field=name;primary_field=handle"`
	Handle      string              `json:"handle" fs:"type=string;filterable;sortable"`
	Name        string              `json:"name"`
	Discipline  string              `json:"discipline" fs:"optional"`
	Labs        []*systemLab        `json:"labs,omitempty" fs.relation:"{'type':'m2m','schema':'system_lab','field':'scientists'}"`
	Experiments []*systemExperiment `json:"experiments,omitempty" fs.relation:"{'type':'o2m','schema':'system_experiment','field':'scientist','owner':true}"`
}

type systemExperiment struct {
	_            any              `json:"-" fs:"name=system_experiment;namespace=id_system_experiments;label_field=title;primary_field=experiment_id"`
	ExperimentID string           `json:"experiment_id" fs:"type=uuid;filterable;sortable"`
	Title        string           `json:"title"`
	Stage        string           `json:"stage" fs:"optional"`
	Lab          *systemLab       `json:"lab" fs.relation:"{'type':'o2m','schema':'system_lab','field':'experiments'}"`
	Scientist    *systemScientist `json:"scientist" fs.relation:"{'type':'o2m','schema':'system_scientist','field':'experiments'}"`
	Samples      []*systemSample  `json:"samples,omitempty" fs.relation:"{'type':'o2m','schema':'system_sample','field':'experiment','owner':true}"`
}

type systemSample struct {
	_          any               `json:"-" fs:"name=system_sample;namespace=id_system_samples;label_field=label;primary_field=sample_no"`
	SampleNo   uint64            `json:"sample_no" fs:"type=uint64;filterable;sortable"`
	Label      string            `json:"label"`
	Status     string            `json:"status" fs:"optional"`
	Experiment *systemExperiment `json:"experiment" fs.relation:"{'type':'o2m','schema':'system_experiment','field':'samples'}"`
}

var idTables = []string{
	"projects_teams",
	"members_teams",
	"labs_scientists",
	"id_artifacts",
	"id_deployments",
	"id_tasks",
	"id_projects",
	"id_engineers",
	"id_teams",
	"id_system_samples",
	"id_system_experiments",
	"id_system_scientists",
	"id_system_labs",
}

func seedIDGraph(t *testing.T, client h.DBClient) *idFixture {
	t.Helper()
	h.ClearDBData(client.C, idTables...)

	projectModel := u.Must(client.C.Model("project"))
	engineerModel := u.Must(client.C.Model("engineer"))
	teamModel := u.Must(client.C.Model("team"))
	taskModel := u.Must(client.C.Model("task"))
	deploymentModel := u.Must(client.C.Model("deployment"))
	artifactModel := u.Must(client.C.Model("artifact"))

	projectCode := u.Must(uuid.NewV7())
	u.Must(projectModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"code":"%s","name":"Alpha","status":"active"}`,
			projectCode,
		),
	))

	engineerHandle := u.Must(uuid.NewV7()).String()
	u.Must(engineerModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"handle":"%s","name":"Ada","level":"senior"}`,
			engineerHandle,
		),
	))

	teamSlug := u.Must(uuid.NewV7()).String()
	u.Must(teamModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"slug":"%s","name":"Backend"}`,
			teamSlug,
		),
	))

	u.Must(projectModel.
		Mutation().
		Where(db.EQ("code", projectCode)).
		Update(h.Ctx(), entity.New().Set("teams", []*entity.Entity{
			refEntity("slug", teamSlug),
		})))

	u.Must(teamModel.
		Mutation().
		Where(db.EQ("slug", teamSlug)).
		Update(h.Ctx(), entity.New().Set("members", []*entity.Entity{
			refEntity("handle", engineerHandle),
		})))

	artifactNo := int(7101)
	u.Must(artifactModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"artifact_no":%d,"description":"bundle","project":{"code":"%s"}}`,
			artifactNo, projectCode,
		),
	))

	deploymentID := u.Must(uuid.NewV7())
	u.Must(deploymentModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"deployment_id":"%s","environment":"staging","project":{"code":"%s"}}`,
			deploymentID, projectCode,
		),
	))

	taskID := h.IDUint64(t, u.Must(taskModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"title":"bootstrap","status":"todo","project":{"code":"%s"},"assignee":{"handle":"%s"}}`,
			projectCode, engineerHandle,
		),
	)))

	return &idFixture{
		projectCode:    projectCode,
		engineerHandle: engineerHandle,
		teamSlug:       teamSlug,
		deploymentID:   deploymentID,
		artifactNo:     artifactNo,
		taskID:         taskID,
	}
}

func refEntity(pkField string, value any) *entity.Entity {
	return entity.New(value).Set(pkField, value)
}

func seedSystemIDGraph(t *testing.T, client h.DBClient) *systemIDFixture {
	t.Helper()
	h.ClearDBData(client.C, idTables...)

	labModel := u.Must(client.C.Model("system_lab"))
	scientistModel := u.Must(client.C.Model("system_scientist"))
	experimentModel := u.Must(client.C.Model("system_experiment"))
	sampleModel := u.Must(client.C.Model("system_sample"))

	labCode := u.Must(uuid.NewV7())
	u.Must(labModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"code":"%s","name":"Andromeda","focus":"energy"}`,
			labCode,
		),
	))

	scientistHandle := u.Must(uuid.NewV7()).String()
	u.Must(scientistModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"handle":"%s","name":"Dr. Zee","discipline":"quantum"}`,
			scientistHandle,
		),
	))

	u.Must(labModel.
		Mutation().
		Where(db.EQ("code", labCode)).
		Update(h.Ctx(), entity.New().Set("scientists", []*entity.Entity{
			refEntity("handle", scientistHandle),
		})))

	experimentID := u.Must(uuid.NewV7())
	u.Must(experimentModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"experiment_id":"%s","title":"Zero-G","stage":"draft","lab":{"code":"%s"},"scientist":{"handle":"%s"}}`,
			experimentID, labCode, scientistHandle,
		),
	))

	sampleNo := uint64(8801)
	u.Must(sampleModel.CreateFromJSON(
		h.Ctx(),
		fmt.Sprintf(
			`{"sample_no":%d,"label":"Specimen","status":"fresh","experiment":{"experiment_id":"%s"}}`,
			sampleNo, experimentID,
		),
	))

	return &systemIDFixture{
		labCode:         labCode,
		scientistHandle: scientistHandle,
		experimentID:    experimentID,
		sampleNo:        sampleNo,
	}
}
