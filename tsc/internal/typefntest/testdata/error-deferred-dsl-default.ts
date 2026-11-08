typefn Twice<N>{return N*2;}
type WithDefault<T,U=Twice<T>>="done";
typefn UseDefault<T>{return WithDefault<T>;}
type Deferred<T>=UseDefault<T>;
type Invalid=Deferred<"bad">;
