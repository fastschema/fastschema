'use strict';

import { getRandomName } from './utils';

let count = 5;

/** @param {FsContext} ctx */
export const ping = (ctx) => {
  return 'pong';
};

/** @param {FsContext} ctx */
export const hello = async (ctx) => {
  return {
    message: 'Hello, ' + ctx.Arg('name', 'World'),
    payload: ctx.Payload(),
  };
};

/** @param {FsContext} ctx */
export const submit = async (ctx) => {
  return {
    payload: ctx.Payload().ToMap(),
  };
};

/** @param {FsContext} ctx */
export const world = async (ctx) => {
  const name = await getRandomName();
  const tx = $db().Tx(ctx);
  /** @type {FsEntity[]} */
  try {
    const roles = tx.Query(
      ctx,
      'SELECT * FROM roles WHERE id IN ($1, $2)',
      1,
      3
    );

    const studentRole = tx.Create(ctx, 'role', { name: 'Student' });
    const teacherRole = tx.Builder('role').Create(ctx, { name: 'Teacher' });
    tx.Builder('role')
      .Where({ id: 1 })
      .Update(ctx, { name: 'Student Updated' });

    tx.Builder('role')
      .Where({ id: studentRole.Get('id') })
      .Delete(ctx);

    return {
      data: 'Hello, World from ' + name,
      roles: roles.map((r) => r.ToMap()),
      studentRole: studentRole.ToMap(),
      teacherRole: teacherRole.ToMap(),
      count: ++count,
      trace_id: ctx.TraceID(),
    };
  } catch (err) {
    console.log('error:', err);
  } finally {
    // always rollback for demo
    tx.Rollback();
  }
};
