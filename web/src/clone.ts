// https://javascript.plainenglish.io/deep-clone-an-object-and-preserve-its-type-with-typescript-d488c35e5574
// MODIFICATION: Made readonly properties writable so we can clone a clone
// MODIFICATION: Replaced any with unknown
// MODIFICATION: Explicit cast to PropertyDescriptor
const clone = <T>(source: T): T => {
  return Array.isArray(source)
    ? source.map((item) => clone(item))
    : source instanceof Date
      ? new Date(source.getTime())
      : source && typeof source === "object"
        ? Object.getOwnPropertyNames(source).reduce((o, prop) => {
          Object.defineProperty(o, prop, {
            ...(Object.getOwnPropertyDescriptor(source, prop) as PropertyDescriptor),
            writable: true,
          });
          o[prop] = clone((source as { [key: string]: unknown })[prop]);
          return o;
        }, Object.create(Object.getPrototypeOf(source)))
        : (source as T);
};

export default clone;
