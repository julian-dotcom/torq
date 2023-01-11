// https://javascript.plainenglish.io/deep-clone-an-object-and-preserve-its-type-with-typescript-d488c35e5574
// With a modification to make readonly properties writable so we can clone a clone
const clone = <T>(source: T): T => {
  return Array.isArray(source)
    ? source.map((item) => clone(item))
    : source instanceof Date
    ? new Date(source.getTime())
    : source && typeof source === "object"
    ? Object.getOwnPropertyNames(source).reduce((o, prop) => {
        Object.defineProperty(o, prop, { ...Object.getOwnPropertyDescriptor(source, prop)!, writable: true });
        o[prop] = clone((source as { [key: string]: any })[prop]);
        return o;
      }, Object.create(Object.getPrototypeOf(source)))
    : (source as T);
};

export default clone;
