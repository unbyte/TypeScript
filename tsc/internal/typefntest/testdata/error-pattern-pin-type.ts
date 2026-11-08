typefn F<T> { match (T) { `unmatched${=1}${tail}` => { return tail; } _ => { return false; } } }
type Result = F<"other">;
