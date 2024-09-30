'use strict';

const product = require('./schemas/product.json');
const {
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
} = require('./hooks');

const {
  ping,
  world,
  message,
} = require('./resources');

/** @param {FsAppConfigActions} config */
const Config = config => {
  config.AddSchemas(product);
  config.port = '8000';

  // config.OnPreResolve(hookPreResolve);
  // config.OnPostResolve(hookPostResolve);

  // config.OnPreDBQuery(hookPreDBQuery);
  // config.OnPostDBQuery(hookPostDBQuery);

  // config.OnPreDBExec(hookPreDBExec);
  // config.OnPostDBExec(hookPostDBExec);

  // config.OnPreDBCreate(hookPreDBCreate);
  // config.OnPostDBCreate(hookPostDBCreate);

  // config.OnPreDBUpdate(hookPreDBUpdate);
  // config.OnPostDBUpdate(hookPostDBUpdate);

  // config.OnPreDBDelete(hookPreDBDelete);
  // config.OnPostDBDelete(hookPostDBDelete);
}

/** @param {FsPlugin} plugin */
const Init = plugin => {
  plugin.resources
    .Group('hello')
    .Add(ping, { public: true })
    .Add(message, { public: true, post: '/message' })
    .Add(world, { public: true });
}
