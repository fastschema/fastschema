'use strict';

/** @param {FsContext} ctx */
const hookPreResolve = ctx => {
  console.log('[plugin] PreResolve this runs before the request is resolved');
}

/** @param {FsContext} ctx */
const hookPostResolve = ctx => {
  console.log('[plugin] PostResolve this runs after the request is resolved');
}

/**
 * @param {FsContext} ctx 
 * @param {FsQueryOption} option 
 */
const hookPreDBQuery = (ctx, option) => {
  console.log('[plugin] PreDBQuery this runs before the query is executed');
}

/**
 * @param {FsContext} ctx 
 * @param {FsQueryOption} option 
 * @param {FsEntity[]} entities 
 */
const hookPostDBQuery = (ctx, option, entities) => {
  console.log('[plugin] PostDBQuery this runs before the query is executed');
  console.log(JSON.stringify(entities, null, 2));
}

/**
 * @param {FsContext} ctx
 * @param {FsQueryOption} option
 */
const hookPreDBExec = (ctx, option) => {
  console.log('[plugin] PreDBExec this runs before the exec is executed');
}

/**
 * @param {FsContext} ctx
 * @param {FsQueryOption} option
 * @param {FsDbExecResult} result
 */
const hookPostDBExec = (ctx, option, result) => {
  console.log('[plugin] PostDBExec this runs after the exec is executed');
  console.log('insertId:', result.LastInsertId());
  console.log('affectedRows:', result.RowsAffected());
}


/**
 * @param {FsContext} ctx 
 * @param {SchemaRawData} schema 
 * @param {FsEntity} createData
 */
const hookPreDBCreate = (ctx, schema, createData) => {
  console.log('[plugin] PreDBCreate this runs before the create is executed');
  console.log('schema:', schema);
  console.log('createData:', createData);
}

/**
 * @param {FsContext} ctx 
 * @param {SchemaRawData} schema 
 * @param {FsEntity} createData
 * @param {number} id
 */
const hookPostDBCreate = (ctx, schema, createData, id) => {
  console.log('[plugin] PostDBCreate this runs after the create is executed');
  console.log('id:', id);
  console.log('createData:', createData);
}

/**
 * @param {FsContext} ctx 
 * @param {SchemaRawData} schema 
 * @param {FsDbPredicate[]} predicates
 * @param {FsEntity} updateData
 */
const hookPreDBUpdate = (ctx, schema, predicates, updateData) => {
  console.log('[plugin] PreDBUpdate this runs before the update is executed');
  console.log('schema:', schema);
  console.log('id:', predicates);
  console.log('createData:', updateData);
}

/**
 * @param {FsContext} ctx
 * @param {SchemaRawData} schema
 * @param {FsDbPredicate[]} predicates
 * @param {FsEntity} updateData
 * @param {FsEntity[]} originalEntities
 * @param {number} affected
 */
const hookPostDBUpdate = (ctx, schema, predicates, updateData, originalEntities, affected) => {
  console.log('[plugin] PostDBUpdate this runs after the update is executed');
  console.log('schema:', schema);
  console.log('predicates:', predicates);
  console.log('updateData:', updateData);
  console.log('originalEntities:', originalEntities);
  console.log('affected:', affected);
}

/**
 * @param {FsContext} ctx
 * @param {SchemaRawData} schema
 * @param {FsDbPredicate[]} predicates
 */
const hookPreDBDelete = (ctx, schema, predicates) => {
  console.log('[plugin] PreDBDelete this runs before the delete is executed');
  console.log('schema:', schema);
  console.log('predicates:', predicates);
}

/**
 * @param {FsContext} ctx
 * @param {SchemaRawData} schema
 * @param {FsDbPredicate[]} predicates
 * @param {FsEntity[]} originalEntities
 * @param {number} affected
 */
const hookPostDBDelete = (ctx, schema, predicates, originalEntities, affected) => {
  console.log('[plugin] PostDBDelete this runs after the delete is executed');
  console.log('schema:', schema);
  console.log('predicates:', predicates);
  console.log('originalEntities:', originalEntities);
  console.log('affected:', affected);
}

module.exports = {
  hookPreResolve,
  hookPostResolve,

  hookPreDBQuery,
  hookPostDBQuery,

  hookPreDBExec,
  hookPostDBExec,
  
  hookPreDBCreate,
  hookPostDBCreate,

  hookPreDBUpdate,
  hookPostDBUpdate,

  hookPreDBDelete,
  hookPostDBDelete,
}
