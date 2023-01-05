// copy pasta from https://stackoverflow.com/questions/28150967/typescript-cloning-object/28152032#28152032
const clone = <T>(source: T): T => {
  if (source === null) return source;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  if (source instanceof Date) return new Date(source.getTime()) as any;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  if (source instanceof Array) return source.map((item: any) => clone<any>(item)) as any;

  if (typeof source === "object" && source !== undefined) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const clonnedObj = { ...(source as { [key: string]: any }) } as { [key: string]: any };
    Object.keys(clonnedObj).forEach((prop) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      clonnedObj[prop] = clone<any>(clonnedObj[prop]);
    });
    return clonnedObj as T;
  }
  return source;
};

export default clone;
