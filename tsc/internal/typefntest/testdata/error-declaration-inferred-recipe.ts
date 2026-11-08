typefn Validate<T> { type checked = Uppercase<T>; return string; }
export function capture<T>(value: Validate<T>) { return [value] as const; }
