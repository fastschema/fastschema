package plugins_test

import (
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

const pluginContentEnv = `
const Init = (plugin) => {
	const invalidResource = plugin.resources.Find('invalid')
	if (invalidResource !== null) {
		throw new Error('Expected invalidResource to be null');
	}

	plugin.resources
		.Find('api')
    .Group('plugin')
    .Add(syncResource, { get: '/sync', public: true })
    .Add(syncResourceError, { get: '/sync-error', public: true })
    .Add(asyncResource, { get: '/async', public: true })
    .Add(asyncResourceError, { get: '/async-error', public: true })
    .Add(asyncResourceErrorThrow, { get: '/async-error-throw', public: true })
    .Add(dbquery, { get: '/db', public: true });
};

const syncResource = () => {
	return { message: 'Hello from syncResource' };
}

const syncResourceError = () => {
	throw new Error('This is a test error from syncResourceError');
}

const asyncResource = async () => {
	return new Promise((resolve) => {
		setTimeout(() => {
			resolve({ message: 'Hello from asyncResource' });
		}, 100);
	});
};

const asyncResourceError = async () => {
	return new Promise((_, reject) => {
		setTimeout(() => {
			reject(new Error('This is a test error from asyncResourceError'));
		}, 100);
	});
};

const asyncResourceErrorThrow = async () => {
	throw new Error('This is a test error from asyncResourceErrorThrow');
};

const dbquery = (ctx) => {
	const tx = $db().Tx(ctx);

	// Create role with executing raw SQL
	const createdRole1 = tx.Exec(
		ctx,
		'INSERT INTO roles (name) VALUES (?)',
		['testrole1'],
	);
	const createdRole1Id = createdRole1.LastInsertId();

	// Query role with executing raw SQL
	const queriedRoles = tx.Query(
		ctx,
		'SELECT * FROM roles WHERE id = ?',
		[createdRole1Id],
	);

	// Create builder with invalid schema
	try {
		tx.Builder('invalid_schema').Where({ id: 1 }).Get(ctx);
	} catch (error) {
		console.log('Expected error for invalid schema:', error.message);
	}

	// Invalid where clause
	try {
		tx.Builder('role').Where({ invalid_field: 1 }).Get(ctx);
	} catch (error) {
		console.log('Expected error for invalid where clause:', error.message);
	}	

	// Create role
	const createdRole2 = tx.Create(ctx, 'role', { name: 'testrole2' });

	// Create role with builder
	const createdRole3 = tx.Builder('role').Create(ctx, { name: 'testrole3' });

	// Builder methods
	const builder = tx.Builder('role')
		.Where({
			id: {
				$gt: 1, // id > 1
			}
		})
		.Limit(1)
		.Offset(1)
		.Select(['id', 'name', 'created_at']);
	const filteredRoles = builder.Get(ctx);
	const filteredRoleCount = builder.Count(ctx);
	const firstFilteredRole = builder.First(ctx);

	// Query only one role
	const onlyRole = tx.Builder('role').Where({ id: createdRole2.Get('id') }).Only(ctx);

	// Update via builder
	tx.Builder('role').Where({ id: createdRole3.Get('id') }).Update(ctx, { name: 'updatedrole3' });

	// Delete via builder
	tx.Builder('role').Where({ id: createdRole1Id }).Delete(ctx);

	// Verify update and delete
	const updatedRole3 = tx.Builder('role').Where({ id: createdRole3.Get('id') }).Only(ctx);
	try {
		tx.Builder('role').Where({ id: createdRole1Id }).Only(ctx);
	} catch (error) {
		console.log('Expected error for deleted role:', error.message);
	}

	// Query role
	const roles = tx.Builder('role').Where({ id: 2 }).Get(ctx);

	tx.Commit();

	// Another transaction to test rollback
	const tx2 = $db().Tx(ctx);
	const roleToDelete = tx2.Builder('role').Where({ id: createdRole2.Get('id') }).Only(ctx);
	tx2.Builder('role').Where({ id: roleToDelete.Get('id') }).Delete(ctx);
	tx2.Rollback();

	// Verify rollback
	const rolledBackRole = $db().Builder('role').Where({ id: createdRole2.Get('id') }).Only(ctx);
	if (rolledBackRole.Get('id') !== createdRole2.Get('id')) {
		throw new Error('Rollback failed: role was deleted');
	}

	// Create transaction error
	try {
		$db().Close();
		const tx3 = $db().Tx(ctx);
	} catch (error) {
		console.log('Expected error for transaction on closed DB:', error.message);
	}

	return {
		queriedRoles: queriedRoles.map(r => r.ToMap()),
		createdRole2: createdRole2.ToMap(),
		createdRole3: createdRole3.ToMap(),
		filteredRoles: filteredRoles.map(r => r.ToMap()),
		filteredRoleCount,
		firstFilteredRole: firstFilteredRole.ToMap(),
		onlyRole: onlyRole.ToMap(),
		updatedRole3: updatedRole3.ToMap(),
		roles: roles.map(r => r.ToMap()),
	};
}

export default {
	Init,
	syncResource,
	syncResourceError,
	asyncResource,
	asyncResourceError,
	asyncResourceErrorThrow,
	dbquery,
};
`

func TestPluginEnv(t *testing.T) {
	app, plugin, _ := createPlugin(t, pluginContentEnv, nil)
	assert.NoError(t, plugin.Init())

	resources := app.Resources()
	server := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	}).Server()

	type testResource struct {
		name         string
		path         string
		expectStatus int
		expectBody   string
	}

	tests := []testResource{
		{"syncResource", "/api/plugin/sync", 200, "syncResource"},
		{"syncResourceError", "/api/plugin/sync-error", 500, "syncResourceError"},
		{"asyncResource", "/api/plugin/async", 200, "asyncResource"},
		{"asyncResourceError", "/api/plugin/async-error", 500, "asyncResourceError"},
		{"asyncResourceErrorThrow", "/api/plugin/async-error-throw", 500, "asyncResourceErrorThrow"},
	}

	for _, tr := range tests {
		t.Run(tr.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tr.path, nil)
			resp := utils.Must(server.Test(req))
			defer func() { assert.NoError(t, resp.Body.Close()) }()
			assert.Equal(t, tr.expectStatus, resp.StatusCode)
			assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), tr.expectBody)
		})
	}

	t.Run("dbquery", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/plugin/db", nil)
		resp := utils.Must(server.Test(req))
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), `"name":"testrole2"`)
	})
}
