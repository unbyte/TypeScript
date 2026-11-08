typefn Twice<N>{return N*2;}
type WithDefault<T,U=Twice<T>>="done";
typefn UseDefault<T>{return WithDefault<T>;}
type Deferred<T>=WithDefault<T>;
type Invalid=Deferred<"bad">;
