typefn F<T,D> { match (T) { `${a}${=D}${b}` => { return [a,b]; } _ => { return false; } } }
type Later<D> = F<"left::right",D>;
type Result = Later<"">;
