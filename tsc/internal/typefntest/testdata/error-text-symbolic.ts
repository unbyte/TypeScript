typefn Len<S> { return text.length<S>; }
type Discard<T> = Len<T> | any;
type Bad = Discard<string>;
