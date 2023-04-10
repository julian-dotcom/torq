import { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import { SerializedError } from "@reduxjs/toolkit";
import clone from "clone";

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

// .map((error: string) => error.split(":")[0])
export function mergeServerError(serverErrors: ServerErrorType, existingErrors: FormErrors): FormErrors {
  serverErrors = clone(serverErrors);
  existingErrors = clone(existingErrors);

  if (existingErrors.server === undefined) {
    existingErrors.server = [];
  }
  existingErrors.server = serverErrors?.errors?.server?.map((se) => {
    if (se.description) {
      // remove unnecessary fluff from GRPC errors
      se.description = se.description.replace("rpc error: code = Unknown desc =", "");
      se.description = se.description.trim();
      se.description = se.description.charAt(0).toUpperCase() + se.description.slice(1);
    }
    return se;
  });
  if (!serverErrors?.errors?.fields) {
    return existingErrors;
  }
  for (const [key, errorCodeOrDescription] of Object.entries(serverErrors?.errors?.fields)) {
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

export function RtqToServerError(rtqErrors: FetchBaseQueryError | SerializedError | undefined): ServerErrorType {
  if (rtqErrors === undefined) {
    return { errors: {} };
  }
  // check if the rtqErrors is a FetchBaseQueryError
  if (!("status" in rtqErrors) || !("data" in rtqErrors)) {
    return { errors: {} };
  }

  const rtqError = rtqErrors as FetchBaseQueryError;
  return rtqError.data as ServerErrorType;
}
