export type FieldName = string;
export type ErrorDescription = string;

export type ValidationErrors = {
  serverErrors: Array<ErrorDescription>;
  fieldErrors: Map<FieldName, Array<ErrorDescription>>;
};

export type FormErrors = {
  server: Array<ErrorDescription>;
  fields: Map<FieldName, Array<ErrorDescription>>;
};

export type ServerErrorType = {
  errors: FormErrors;
};

// Example JSON from server as specified in server_errors.go
// {"errors":{"fields":null,"server":["Obtaining publicKey/chain/network from gRPC"]}}

// export function parseServerError(errorJSON: {errors: {server: Array<string>}}): ValidationErrors {
//   return errorJSON as ValidationErrors;
// }
