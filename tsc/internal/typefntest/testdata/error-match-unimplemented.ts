typefn Match<T> { match (T) { _ => { return T; } } }
type Result = Match<1>;
