typefn Twice<N>{return N*2;}
interface DefaultContainer<T,U=Twice<T>> {tag:"container";}
type Erase<X>=string;
type Hidden<T>=Erase<DefaultContainer<T>>;
type Invalid=Hidden<"bad">;
