typefn F<T> { match (T) { `unmatched${=""}${tail}` => { return tail; } _ => { return false; } } }
type Result = F<"other">;
