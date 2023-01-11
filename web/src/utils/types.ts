// to cast a Class to this do this: myObject as unknown as Record<string, unknown>
// When trying to cast an interface to this it will fail, change your interface to a type
export type AnyObject = Record<string, unknown>;
