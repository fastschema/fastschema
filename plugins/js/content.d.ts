export interface FsRole {
  id?: number;
  name?: string;
  description?: string;
  root?: boolean;
  users?: (FsUser | undefined)[];
  permissions?: (FsPermission | undefined)[];
  created_at?: string;
  updated_at?: string;
  deleted_at?: string;
}

export interface FsPermission {
  id?: number;
  role_id?: number;
  resource?: string;
  value?: string;
  role?: FsRole;
  created_at?: string;
  updated_at?: string;
  deleted_at?: string;
}

export interface FsFile {
  id?: number;
  disk?: string;
  name?: string;
  path?: string;
  type?: string;
  size?: number;
  user_id?: number;
  user?: FsUser;
  url?: string;
  created_at?: string;
  updated_at?: string;
  deleted_at?: string;
}

export interface FsUser {
  id?: number;
  username?: string;
  email?: string;
  password?: string;
  active?: boolean;
  provider?: string;
  provider_id?: string;
  provider_username?: string;
  role_ids?: number[];
  roles?: (FsRole | undefined)[];
  files?: (FsFile | undefined)[];
  created_at?: string;
  updated_at?: string;
  deleted_at?: string;
}
