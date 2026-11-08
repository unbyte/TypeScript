typefn Reject<T>: type { failAt<T, 3>; }
type Generic<T> = Reject<T>;
type Bad = Generic<"Expected digit">;
