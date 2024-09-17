'use strict';

const { getRandomName } = require('./utils');

let count = 5;

/** @param {FsContext} ctx */
const ping = ctx => {
  return 'pong';
}

/** @param {FsContext} ctx */
const world = async ctx => {
  const name = await getRandomName();
  const tx = $db().Tx(ctx);
  /** @type {FsEntity[]} */
  let roles = [];

  try {
    roles = tx.Query(ctx, 'SELECT * FROM roles WHERE id IN ($1, $2)', 1, 3);
    const result = tx.Exec(
      ctx,
      'INSERT INTO roles (name) VALUES ($1) RETURNING *',
      'Teacher'
    )
    console.log('inserted teacher role:', result.LastInsertId());
    console.log('affected rows:', result.RowsAffected())

    const teacherRole = tx.Query(ctx, 'SELECT * FROM roles WHERE name = $1', 'Teacher');
    console.log('teacher role:', teacherRole);

    const builder = tx.Builder('role');
    const studentRole = builder.Create(ctx, { name: 'Student' });
    console.log('student role');
    console.log(studentRole.Get('id'));

    tx.Builder('role')
      .Where({ id: 1 })
      .Update(ctx, { name: 'Student Updated' });

    tx.Builder('role')
      .Where({ id: studentRole.Get('id') })
      .Delete(ctx);
  } catch (err) {
    console.log('error:', err);
  }

  tx.Rollback();

  return {
    data: 'Hello, World from ' + name,
    roles,
    count: ++count,
    trace_id: ctx.TraceID(),
  };
}

/** @param {FsContext} ctx */
const message = async ctx => {
  return {
    message: 'Hello, ' + ctx.Arg('name', 'World'),
    payload: ctx.Payload(),
  };
};

/**
 * @param {FsContext} ctx 
 * @param {FsQueryOption} option 
 */
const helloPreDBQuery = (ctx, option) => {
  console.log('[plugin] PreDBQuery this runs before the query is executed');
}

module.exports = {
  ping,
  world,
  message,
  helloPreDBQuery,
};
