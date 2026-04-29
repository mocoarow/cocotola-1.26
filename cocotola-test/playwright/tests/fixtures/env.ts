type EnvKey =
  | "OWNER_LOGIN_ID"
  | "OWNER_PASSWORD"
  | "ORGANIZATION_NAME"
  | "NEW_USER_LOGIN_ID"
  | "NEW_USER_PASSWORD";

function read(name: EnvKey): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Environment variable ${name} is required (see env.example)`);
  }
  return value;
}

export const testEnv = {
  get ownerLoginId() {
    return read("OWNER_LOGIN_ID");
  },
  get ownerPassword() {
    return read("OWNER_PASSWORD");
  },
  get organizationName() {
    return read("ORGANIZATION_NAME");
  },
  get newUserLoginId() {
    return read("NEW_USER_LOGIN_ID");
  },
  get newUserPassword() {
    return read("NEW_USER_PASSWORD");
  },
} as const;
