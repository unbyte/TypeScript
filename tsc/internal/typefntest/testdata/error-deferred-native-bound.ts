typefn Twice<N> {return N*2;}
type Capital<T>=Uppercase<Twice<T>>;
type Invalid=Capital<21>;
