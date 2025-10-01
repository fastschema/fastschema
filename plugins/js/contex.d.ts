import { FsFile, FsUser } from './content';
import { FsLogger } from './logger';
import { FsArg, FsResource } from './resource';

export interface FsResult {
  error?: {
    Error: () => string;
  };
  data?: any;
}

export interface FsContext {
  TraceID: () => string;
  User: () => FsUser;
  Local: <T = any>(name: string, defaultValue?: T) => T;
  Value: <T = any>(name: string) => T;
  Logger: () => FsLogger;
  Args: () => Record<string, FsArg>;
  Arg: (name: string, defaultValue?: string) => string;
  ArgInt: (name: string, defaultValue?: int) => int;
  Payload: () => FsEntity;
  Resource: () => FsResource;
  AuthToken: () => string;
  Next: () => error;
  Result: (result?: FsResult) => FsResult;
  Files: () => FsFile[];
  Redirect(url: string): void;
}
