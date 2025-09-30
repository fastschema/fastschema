import { SchemaRawData as _SchemaRawData, FsEntity as _FsEntity } from './schema';
import {
  _FsPreDBQuery,
  _FsPostDBQuery,
  _FsPreDBExec,
  _FsPostDBExec,
  _FsPreDBCreate,
  _FsPostDBCreate,
  _FsPreDBUpdate,
  _FsPostDBUpdate,
  _FsPreDBDelete,
  _FsPostDBDelete,
} from './hooks';

export interface FsLoggerConfig {
  development: boolean;
  log_file: string;
  caller_skip: number;
  disable_console: boolean;
}

export interface FsDbConfig {
  driver: string;
  name: string;
  host: string;
  port: string;
  user: string;
  pass: string;
  log_queries: boolean;
  migration_dir: string;
  ignore_migration: boolean;
  disable_foreign_keys: boolean;
}

export interface FsDiskConfig {
  name: string;
  driver: string;
  root: string;
  base_url: string;
  public_path: string;
  provider: string;
  endpoint: string;
  region: string;
  bucket: string;
  access_key_id: string;
  secret_access_key: string;
  acl: string;
}

export interface FsStorageConfig {
  default_disk: string;
  disks: (DiskConfig | undefined)[];
}

export interface FsAuthConfig {
  enabled_providers: string[];
  providers: { [key: string]: { [key: string]: string } };
}

export interface FsAppConfig {
  dir: string;
  app_key: string;
  readonly port: string;
  base_url: string;
  dash_url: string;
  api_base_name: string;
  dash_base_name: string;
  logger_config?: FsLoggerConfig;
  db_config?: FsDbConfig;
  storage_config?: FsStorageConfig;
  hide_resources_info: boolean;
  auth_config?: FsAuthConfig;

  Set: (config: { [key: string]: any }) => void;

  AddSchemas: (...schemas: SchemaRawData[]) => void;

  OnPreResolve(...middlewares: FsMiddleware[]): void;
  OnPostResolve(...middlewares: FsMiddleware[]): void;

  OnPreDBQuery(...hooks: _FsPreDBQuery[]): void;
  OnPostDBQuery(...hooks: _FsPostDBQuery[]): void;

  OnPreDBExec(...hooks: _FsPreDBExec[]): void;
  OnPostDBExec(...hooks: _FsPostDBExec[]): void;

  OnPreDBCreate(...hooks: _FsPreDBCreate[]): void;
  OnPostDBCreate(...hooks: _FsPostDBCreate[]): void;

  OnPreDBUpdate(...hooks: _FsPreDBUpdate[]): void;
  OnPostDBUpdate(...hooks: _FsPostDBUpdate[]): void;

  OnPreDBDelete(...hooks: _FsPreDBDelete[]): void;
  OnPostDBDelete(...hooks: _FsPostDBDelete[]): void;
}
