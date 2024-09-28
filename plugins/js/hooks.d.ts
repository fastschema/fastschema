export type _FsResolveHook = (ctx: FsContext) => Promise<void> | void;

export type _FsPreDBQueryHook = (
  ctx: FsContext,
  option: FsQueryOption,
) => Promise<void> | void;

export type _FsPostDBQueryHook = (
  ctx: FsContext,
  option: FsQueryOption,
  entities: FsEntity[],
) => Promise<void> | void;

export type _FsPreDBExecHook = (
  ctx: FsContext,
  option: FsQueryOption,
) => Promise<void> | void;

export type _FsPostDBExecHook = (
  ctx: FsContext,
  option: FsQueryOption,
  result: FsDbExecResult,
) => Promise<void> | void;

export type _FsPreDBCreateHook = (
  ctx: FsContext,
  schema: SchemaRawData,
  entity: FsEntity,
) => Promise<void> | void;

export type _FsPostDBCreateHook = (
  ctx: FsContext,
  schema: SchemaRawData,
  entity: FsEntity,
  createdId: number,
) => Promise<void> | void;

export type _FsPreDBUpdateHook = (
  ctx: FsContext,
  schema: SchemaRawData,
  predicates: FsDbPredicate[],
  createData: FsEntity,
) => Promise<void> | void;

export type _FsPostDBUpdateHook = (
  ctx: FsContext,
  schema: SchemaRawData,
  predicates: FsDbPredicate[],
  createData: FsEntity,
  originalEntities: FsEntity[],
  affected: number,
) => Promise<void> | void;

export type _FsPreDBDeleteHook = (
  ctx: FsContext,
  schema: SchemaRawData,
  predicates: FsDbPredicate[],
) => Promise<void> | void;

export type _FsPostDBDeleteHook = (
  ctx: FsContext,
  schema: SchemaRawData,
  predicates: FsDbPredicate[],
  originalEntities: FsEntity[],
  affected: number,
) => Promise<void> | void;
