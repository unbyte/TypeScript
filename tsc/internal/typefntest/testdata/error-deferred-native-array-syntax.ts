typefn Twice<N> { return N*2; }
type Erase<X>=string;
type Hidden<T>=Erase<Twice<T>[]>;
type Invalid=Hidden<"bad">;
