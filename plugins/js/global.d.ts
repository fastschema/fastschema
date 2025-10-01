import { FsArg, FsMeta, FsResource, FsResourceHandler } from './resource';
import {
  FsDb,
  FsQueryOption as _FsQueryOption,
  FsDbExecResult as _FsDbExecResult,
  FsDbPredicate as _FsDbPredicate,
} from './db';
import { FsFile, FsUser } from './content';
import { FsContext as _FsContext } from './contex';
import { FsAppConfig as _FsAppConfig } from './config';
import { FsLogger as _FsLogger } from './logger';
import {
  SchemaRawData as _SchemaRawData,
  FsEntity as _FsEntity,
} from './schema';

export interface _FsPlugin {
  name: string;
  resources: FsResource;
}

declare global {
  interface FsPlugin extends _FsPlugin {}
  interface FsContext extends _FsContext {}
  interface FsAppConfig extends _FsAppConfig {}
  interface FsQueryOption extends _FsQueryOption {}
  interface FsSchema extends SchemaRawData {}
  interface FsDbExecResult extends _FsDbExecResult {}
  interface SchemaRawData extends _SchemaRawData {}
  interface FsDbPredicate extends _FsDbPredicate {}
  interface FsEntity extends _FsEntity {}
  interface FsLogger extends _FsLogger {}

  const $db: () => FsDb;
  const $context: () => FsContext;
  const $logger: () => FsLogger;
}
