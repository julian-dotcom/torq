//for a given enum value it attempts to find the key and attempts remove camel casing to display to user had to use any
//here because there isn't a good way to pass an enum as an argument!
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function GetFormattedEnumLabelByValue(enumType: any, enumValue: number) {
  const label = Object.keys(enumType).find((key) => enumType[key] === enumValue);

  return label ? label.replace(/([a-z])([A-Z])/g, "$1 $2") : "";
}
