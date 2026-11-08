typefn F<T> { match (T) { `${a}aaab${b}` => { return [a,b]; } _ => { return false; } } }
type Result = F<"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaac">;
