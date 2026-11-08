typefn Fields<T> {return reflect.properties<T>;}
type Result=Fields<{a:1}|{b:2}>;
