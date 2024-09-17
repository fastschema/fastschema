// export type FsDBOperator = '$eq' | '$neq' | '$gt' | '$gte' | '$lt' | '$lte' | '$like' | '$in' | '$nin' | '$null';

import { FsEntity, SchemaRawData } from './schema';

export interface FsDbPredicate {
  Field: string;
  Operator: FsDBOperator;
  Value: any;
  RelationFieldNames: string[];
  And?: FsDbPredicate[];
  Or?: FsDbPredicate[];
}

export interface FsQueryOption {
  schema: SchemaRawData;
  limit: number;
  offset: number;
  columns: string[];
  order: string[];
  predicates?: FsDbPredicate[];
}

export interface FsDbCountOption {
  Column: string;
  Unique: boolean;
}

export interface FsDbQueryBuilder {
  // Where: (...predicates: FsDbPredicate[]) => FsDbQueryBuilder;
  Where: (...predicates: Object[]) => FsDbQueryBuilder;

  // Mutation methods
  Create: (ctx: FsContext, dataCreate: any) => FsEntity;
  CreateFromJSON: (ctx: FsContext, json: string) => FsEntity;
  Update: (ctx: FsContext, updateData: any) => FsEntity[];
  Delete: (ctx: FsContext) => number;

  // Query methods
  Limit: (limit: number) => FsDbQueryBuilder;
  Offset: (offset: number) => FsDbQueryBuilder;
  Order: (...order: string[]) => FsDbQueryBuilder;
  Select: (...fields: string[]) => FsDbQueryBuilder;
  Count: (options: FsDbCountOption) => number;
  Get: (ctx: FsContext) => FsEntity[];
  First: (ctx: FsContext) => FsEntity;
  Only: (ctx: FsContext) => FsEntity;
}

export interface FsDbExecResult {
  LastInsertId: () => number;
	RowsAffected: () => number;
}

export interface FsDb {
  Exec: (c: FsContext, sql: string, ...args: any) => FsDbExecResult;
  Query: (c: FsContext, sql: string, ...args: any) => FsEntity[];
  Builder: (schema: string) => FsDbQueryBuilder;
  Tx: (c: FsContext) => FsDb;
  Commit: () => void;
  Rollback: () => void;
}

export type PreDBQuery = (ctx: FsContext, query: FsQueryOption) => void;
