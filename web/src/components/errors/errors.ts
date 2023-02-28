export type FieldName = string;
export type ErrorDescription = string;

export type ErrorCodeOrDescription = {
  code?: string;
  description?: string;
  attributes: Record<string, string>;
};

export type FormErrors = {
  server?: Array<ErrorCodeOrDescription>;
  fields?: Record<FieldName, Array<ErrorCodeOrDescription>>;
};

export type ServerErrorType = {
  errors: FormErrors;
};

export function mergeServerError(serverErrors: ServerErrorType, existingErrors: FormErrors): FormErrors {
  if (existingErrors.server === undefined) {
    existingErrors.server = [];
  }
  existingErrors.server = serverErrors.errors.server?.map((se) => {
    if (se.description) {
      // remove unessary fluff from GRPC errors
      se.description = se.description.replace("rpc error: code = Unknown desc =", "");
      se.description = se.description.trim();
      se.description = se.description.charAt(0).toUpperCase() + se.description.slice(1);
    }
    return se;
  });
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

// template messages should use :mergetag: format and attributes should just use mergetag (without colons) as key
export function replaceMessageMergeTags(message: string, attributes: Record<string, string>): string {
  const words = message.split(" ");
  for (const word of words) {
    if (word.startsWith(":") && word.endsWith(":")) {
      message = message.replace(word, attributes[word.slice(1, -1) ?? ""]);
    }
  }
  return message;
}
