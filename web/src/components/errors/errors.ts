export type FieldName = string;
export type ErrorDescription = string;

export type ErrorCodeOrDescription = {
  code?: string;
  description?: string;
};

export type FormErrors = {
  server?: Array<ErrorCodeOrDescription>;
  fields?: Record<FieldName, Array<ErrorCodeOrDescription>>;
};

export type ServerErrorType = {
  errors: FormErrors;
};

// .map((error: string) => error.split(":")[0])
export function mergeServerError(serverErrors: ServerErrorType, existingErrors: FormErrors): FormErrors {
  if (existingErrors.server === undefined) {
    existingErrors.server = [];
  }
  existingErrors.server = serverErrors.errors.server;
  if (!serverErrors.errors.fields) {
    return existingErrors;
  }
  for (const [key, errorCodeOrDescription] of Object.entries(serverErrors.errors.fields)) {
    if (existingErrors.fields === undefined) {
      existingErrors.fields = {} as Record<FieldName, Array<ErrorCodeOrDescription>>;
    }
    if (!existingErrors.fields[key]) {
      existingErrors.fields[key] = [];
    }
    existingErrors.fields[key].push(...errorCodeOrDescription);
  }
  return existingErrors;
}
