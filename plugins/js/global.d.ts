import { FsArg, FsMeta, FsResource, FsResourceHandler } from './resource';
import { FsHooks } from './hook';
import {
  FsDb,
  FsQueryOption as _FsQueryOption,
  FsDbExecResult as _FsDbExecResult,
  FsDbPredicate as _FsDbPredicate,
} from './db';
import { FsFile, FsUser } from './content';
import { FsContext as _FsContext } from './contex';
import { FsAppConfig as _FsAppConfig } from './config';
import { SchemaRawData as _SchemaRawData, FsEntity as _FsEntity } from './schema';

import {
  _FsResolveHook,
  _FsPreDBQueryHook,
  _FsPostDBQueryHook,
  _FsPreDBExecHook,
  _FsPostDBExecHook,
  _FsPreDBCreateHook,
  _FsPostDBCreateHook,
  _FsPreDBUpdateHook,
  _FsPostDBUpdateHook,
  _FsPreDBDeleteHook,
  _FsPostDBDeleteHook,
} from './hooks';

export interface _FsPlugin {
  name: string;
  resources: FsResource;
  plugin: FsPlugin;
}

export interface _FsAppConfigActions extends _FsAppConfig {
  AddSchemas: (...schemas: SchemaRawData[]) => void;

  OnPreResolve(...hooks: _FsResolveHook[]): void;
  OnPostResolve(...hooks: _FsResolveHook[]): void;

  OnPreDBQuery(...hooks: _FsPreDBQueryHook[]): void;
  OnPostDBQuery(...hooks: _FsPostDBQueryHook[]): void;

  OnPreDBExec(...hooks: _FsPreDBExecHook[]): void;
  OnPostDBExec(...hooks: _FsPostDBExecHook[]): void;

  OnPreDBCreate(...hooks: _FsPreDBCreateHook[]): void;
  OnPostDBCreate(...hooks: _FsPostDBCreateHook[]): void;

  OnPreDBUpdate(...hooks: _FsPreDBUpdateHook[]): void;
  OnPostDBUpdate(...hooks: _FsPostDBUpdateHook[]): void;

  OnPreDBDelete(...hooks: _FsPreDBDeleteHook[]): void;
  OnPostDBDelete(...hooks: _FsPostDBDeleteHook[]): void;
}

declare global {
  interface FsPlugin extends _FsPlugin { }
  interface FsContext extends _FsContext { }
  interface FsAppConfig extends _FsAppConfig { }
  interface FsAppConfigActions extends _FsAppConfigActions { }
  interface FsQueryOption extends _FsQueryOption { }
  interface FsSchema extends SchemaRawData { }
  interface FsDbExecResult extends _FsDbExecResult { }
  interface SchemaRawData extends _SchemaRawData { }
  interface FsDbPredicate extends _FsDbPredicate { }
  interface FsEntity extends _FsEntity { }
  type FsResolveHook = _FsResolveHook

  const $db: () => FsDb;
  const $context: () => FsContext;
}
