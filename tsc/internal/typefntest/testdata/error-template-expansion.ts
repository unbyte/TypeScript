typefn Interpolate<T extends string> { return `abcdefghij${T}`; }
type Generic<T extends string> = Interpolate<T>;
type Bad = Generic<"a" | "b" | "c" | "d">;
