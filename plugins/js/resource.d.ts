export type FsArgType =
  | 'bool'
  | 'time'
  | 'json'
  | 'uuid'
  | 'bytes'
  | 'enum'
  | 'string'
  | 'text'
  | 'int'
  | 'int8'
  | 'int16'
  | 'int32'
  | 'int64'
  | 'uint'
  | 'uint8'
  | 'uint16'
  | 'uint32'
  | 'uint64'
  | 'float32'
  | 'float64';

export interface FsArg {
  Type: FsArgType;
  Required?: boolean;
  Description?: string;
  Example?: any;
}

export interface FsMeta {
  prefix?: string;
  get?: string;
  head?: string;
  post?: string;
  put?: string;
  delete?: string;
  connect?: string;
  options?: string;
  trace?: string;
  patch?: string;
  ws?: string;
  public?: boolean;
  args?: Record<string, FsArg>;
  signatures?: any[2];
}

export type FsMiddleware = (ctx: FsContext) => Promise<void> | void;

// Remove the input parameter from the handler function
//  The input should be passed through the context object: ctx.Entity()
export type FsResourceHandler<O = any> = (ctx: FsContext) => Promise<O> | O;

export interface FsResource {
  Group: (name: string, meta?: FsMeta) => FsResource;
  // Replace the implementation of the Add method
  // with the implementation of the AddResource method
  // for the convenience usage.
  // Add: (name: string, handler: FsResourceHandler, meta?: FsMeta) => void;
  Add: (handler: FsResourceHandler, meta?: FsMeta) => FsResource;
  // Add: (handler: string, meta?: FsMeta) => FsResource;
  // Add: (...resources: FsResource) => ThisParameterType;
  // AddResource: (
  //   name: string,
  //   handler: FsResourceHandler,
  //   meta?: FsMeta
  // ) => ThisParameterType;
  // Remove: (resource: FsResource) => ThisParameterType;
  // Find: (resourceID: string) => FsResource;
  // ID: () => string;
  // Name: () => string;
  // Handler: () => Handler;
  // Meta: () => FsMeta | null;
  // Signature: () => any[2];
  // Resources: () => FsResource[];
  // IsGroup: () => bool;
  // IsGroup: () => bool;
  // IsPublic: () => bool;
  // Group: (name: string, meta?: FsMeta) => FsResource;
  // String: () => string;
  // Init: () => error;
  // MarshalJSON: () => BinaryType;
}
