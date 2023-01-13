export function IsStringOption(result: unknown): result is { value: "string"; label: string } {
  return result !== null && typeof result === "object" && "value" in result && "label" in result;
}
