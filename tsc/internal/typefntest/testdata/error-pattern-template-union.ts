typefn F<T> { match (T) { `${tail}` => { return tail; } _ => { return false; } } }
type Result = F<"a"|"b">;
