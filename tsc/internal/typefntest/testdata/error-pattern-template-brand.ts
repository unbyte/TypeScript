typefn F<T> { match (T) { `${tail}` => { return tail; } _ => { return false; } } }
type Result = F<string & {readonly brand: unique symbol}>;
