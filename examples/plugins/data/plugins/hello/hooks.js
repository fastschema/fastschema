'use strict';

/**
 * @param {FsContext} ctx
 * @param {FsQueryOption} option
 */
export const preDBQuery = (ctx, option) => {
  console.log('[plugin] PreDBQuery this runs before the query is executed');
};

/**
 * @param {FsContext} ctx
 * @param {FsQueryOption} option
 * @param {FsEntity[]} entities
 */
export const postDBQuery = (ctx, option, entities) => {
  console.log(
    '[plugin] PostDBQuery this runs after the query is executed, result: ',
    JSON.stringify(entities.map(e => e.ToMap()), null, 2)
  );
};

/**
 * @param {FsContext} ctx
 * @param {FsQueryOption} option
 */
export const preDBExec = (ctx, option) => {
  console.log('[plugin] PreDBExec this runs before the exec is executed');
};

/**
 * @param {FsContext} ctx
 * @param {FsQueryOption} option
 * @param {FsDbExecResult} result
 */
export const postDBExec = (ctx, option, result) => {
  console.log('[plugin] PostDBExec this runs after the exec is executed');
  console.log('insertId:', result.LastInsertId());
  console.log('affectedRows:', result.RowsAffected());
};

/**
 * @param {FsContext} ctx
 * @param {SchemaRawData} schema
 * @param {FsEntity} createData
 */
export const preDBCreate = (ctx, schema, createData) => {
  console.log('[plugin] PreDBCreate this runs before the create is executed');
  console.log('schema:', schema);
  console.log('createData:', createData.ToMap());
};

/**
 * @param {FsContext} ctx
 * @param {SchemaRawData} schema
 * @param {FsEntity} createData
 * @param {number} id
 */
export const postDBCreate = (ctx, schema, createData, id) => {
  console.log('[plugin] PostDBCreate this runs after the create is executed');
  console.log('id:', id);
  console.log('createData:', createData.ToMap());
};

/**
 * @param {FsContext} ctx
 * @param {SchemaRawData} schema
 * @param {FsDbPredicate[]} predicates
 * @param {FsEntity} updateData
 */
export const preDBUpdate = (ctx, schema, predicates, updateData) => {
  console.log('[plugin] PreDBUpdate this runs before the update is executed');
  console.log('schema:', schema);
  console.log('id:', predicates);
  console.log('createData:', updateData.ToMap());
};

/**
 * @param {FsContext} ctx
 * @param {SchemaRawData} schema
 * @param {FsDbPredicate[]} predicates
 * @param {FsEntity} updateData
 * @param {FsEntity[]} originalEntities
 * @param {number} affected
 */
export const postDBUpdate = (
  ctx,
  schema,
  predicates,
  updateData,
  originalEntities,
  affected
) => {
  console.log('[plugin] PostDBUpdate this runs after the update is executed');
  console.log('schema:', schema);
  console.log('predicates:', predicates);
  console.log('updateData:', updateData);
  console.log('originalEntities:', originalEntities.map(e => e.ToMap()));
  console.log('affected:', affected);
};

/**
 * @param {FsContext} ctx
 * @param {SchemaRawData} schema
 * @param {FsDbPredicate[]} predicates
 */
export const preDBDelete = (ctx, schema, predicates) => {
  console.log('[plugin] PreDBDelete this runs before the delete is executed');
  console.log('schema:', schema);
  console.log('predicates:', predicates);
};

/**
 * @param {FsContext} ctx
 * @param {SchemaRawData} schema
 * @param {FsDbPredicate[]} predicates
 * @param {FsEntity[]} originalEntities
 * @param {number} affected
 */
export const postDBDelete = (
  ctx,
  schema,
  predicates,
  originalEntities,
  affected
) => {
  console.log('[plugin] PostDBDelete this runs after the delete is executed');
  console.log('schema:', schema);
  console.log('predicates:', predicates);
  console.log('originalEntities:', originalEntities);
  console.log('affected:', affected);
};

/** @param {FsContext} ctx */
export const preResolve = (ctx) => {
  console.log('[plugin] PreResolve this runs before the request is resolved');
};

/** @param {FsContext} ctx */
export const postResolve = (ctx) => {
  console.log('[plugin] PostResolve this runs after the request is resolved');
};

export default {
  preResolve,
  postResolve,

  preDBQuery,
  postDBQuery,

  preDBExec,
  postDBExec,

  preDBCreate,
  postDBCreate,

  preDBUpdate,
  postDBUpdate,

  preDBDelete,
  postDBDelete,
};
