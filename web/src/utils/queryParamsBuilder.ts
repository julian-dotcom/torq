export const queryParamsBuilder = <T extends { [k: string]: unknown }>(params: T, isPaginable = false): string => {
  let paramObject: { [k: string]: string } = {};
  for (const key of Object.keys(params)) {
    if (params[key] === undefined) continue;
    if (typeof params[key] === "object") {
      paramObject[key] = JSON.stringify(params[key]);
      continue;
    }
    paramObject[key] = String(params[key]);
  }

  if (isPaginable) {
    paramObject = { limit: "100", offset: "0", ...paramObject };
  }

  return "?" + new URLSearchParams(paramObject).toString();
};
