export type _FsPreDBQuery = (
  ctx: FsContext,
  option: FsQueryOption,
) => Promise<void> | void;

export type _FsPostDBQuery = (
  ctx: FsContext,
  option: FsQueryOption,
  entities: FsEntity[],
) => Promise<void> | void;

export type _FsPreDBExec = (
  ctx: FsContext,
  option: FsQueryOption,
) => Promise<void> | void;

export type _FsPostDBExec = (
  ctx: FsContext,
  option: FsQueryOption,
  result: FsDbExecResult,
) => Promise<void> | void;

export type _FsPreDBCreate = (
  ctx: FsContext,
  schema: SchemaRawData,
  entity: FsEntity,
) => Promise<void> | void;

export type _FsPostDBCreate = (
  ctx: FsContext,
  schema: SchemaRawData,
  entity: FsEntity,
  createdId: number,
) => Promise<void> | void;

export type _FsPreDBUpdate = (
  ctx: FsContext,
  schema: SchemaRawData,
  predicates: FsDbPredicate[],
  createData: FsEntity,
) => Promise<void> | void;

export type _FsPostDBUpdate = (
  ctx: FsContext,
  schema: SchemaRawData,
  predicates: FsDbPredicate[],
  createData: FsEntity,
  originalEntities: FsEntity[],
  affected: number,
) => Promise<void> | void;

export type _FsPreDBDelete = (
  ctx: FsContext,
  schema: SchemaRawData,
  predicates: FsDbPredicate[],
) => Promise<void> | void;

export type _FsPostDBDelete = (
  ctx: FsContext,
  schema: SchemaRawData,
  predicates: FsDbPredicate[],
  originalEntities: FsEntity[],
  affected: number,
) => Promise<void> | void;
