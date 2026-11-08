typefn F<T> { match (T) { [item] => { return item; } _ => { return 0; } } }
type Result = F<[1] | [2]>;
