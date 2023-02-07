export type FieldName = string;
export type ErrorDescription = string;

export type ErrorCodeOrDescription = {
  code?: string;
  description?: string;
};

export type FormErrors = {
  server: Array<ErrorCodeOrDescription>;
  fields: Map<FieldName, Array<ErrorCodeOrDescription>>;
};

export type ServerErrorType = {
  errors: FormErrors;
};

// .map((error: string) => error.split(":")[0])
export function mergeServerError(serverErrors: ServerErrorType, existingErrors: FormErrors): FormErrors {
  if (existingErrors.server === undefined) {
    existingErrors.server = [];
  }
  existingErrors.server.push(...serverErrors.errors.server);
  for (const [key, errorCodeOrDescription] of serverErrors.errors.fields) {
    if (!existingErrors.fields.get(key)) {
      existingErrors.fields.set(key, []);
    }
    existingErrors.fields.get(key)?.push(...errorCodeOrDescription);
  }
  return existingErrors;
}
