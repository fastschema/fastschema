import productSchema from './schemas/product.json';
import {
  preDBQuery,
  postDBQuery,
  //
  preDBExec,
  postDBExec,
  //
  preDBCreate,
  postDBCreate,
  //
  preDBUpdate,
  postDBUpdate,
  //
  preDBDelete,
  postDBDelete,
  //
  preResolve,
  postResolve,
} from './hooks';

import { hello, ping, submit, world } from './resources';

/** @param {FsAppConfig} config */
const Config = (config) => {
  config.Set({ port: '9000' });

  config.AddSchemas(productSchema);

  config.OnPreDBQuery(preDBQuery);
  config.OnPostDBQuery(postDBQuery);

  config.OnPreDBExec(preDBExec);
  config.OnPostDBExec(postDBExec);

  config.OnPreDBCreate(preDBCreate);
  config.OnPostDBCreate(postDBCreate);

  config.OnPreDBUpdate(preDBUpdate);
  config.OnPostDBUpdate(postDBUpdate);

  config.OnPreDBDelete(preDBDelete);
  config.OnPostDBDelete(postDBDelete);

  config.OnPreResolve(preResolve);
  config.OnPostResolve(postResolve);
};

/** @param {FsPlugin} plugin */
const Init = (plugin) => {
  plugin.resources
    .Group('plugin')
    .Add(ping, { public: true })
    .Add(submit, { public: true, post: '/submit' })
    .Add(hello, { public: true })
    .Add(world, { public: true });
  $logger().Info('Hello plugin initialized');
};

export default {
  Config,
  Init,
  //
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
  preResolve,
  postResolve,
  //
  ping,
  submit,
  hello,
  world,
};
