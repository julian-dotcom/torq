//for a given enum value it attempts to find the key and attempts remove camel casing to display to user
export function GetFormattedEnumLabelByValue(enumType: any, enumValue: number) {
  const label = Object.keys(enumType).find((key) => enumType[key] === enumValue);

  return label ? label.replace(/([a-z])([A-Z])/g, "$1 $2") : "";
}
