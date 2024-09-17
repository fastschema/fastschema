export type FsLogContext = Record<string, any>;
export interface FsLogger {
  Info: (...msgs: any) => void;
  Infof: (format: string, ...params: any) => void;
  Error: (...msgs: any) => void;
  Errorf: (format: string, ...params: any) => void;
  Debug: (...msgs: any) => void;
  Fatal: (...msgs: any) => void;
  Warn: (...msgs: any) => void;
  Panic: (...msgs: any) => void;
  DPanic: (...msgs: any) => void;
  WithContext: (context: FsLogContext, callerSkips?: int) => FsLogger;
}
