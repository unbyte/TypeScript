typefn Reflect<T>{return reflect.tuple<T>;}
type Result=Reflect<Partial<[string]>>;
