type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
typefn Twice<N> {return N*2;}
type NumberOnly<N extends number>=string;
type Checked<T>=NumberOnly<Twice<T>>;
type Result=Assert<Equal<Checked<21>,string>>;
interface NumericBox<N extends number>{value:N;}
type Boxed<T>=NumericBox<Twice<T>>;
type Field=Assert<Equal<Boxed<21>["value"],42>>;
type Hidden<T>=Array<Twice<T>>;
type ElementResult=Assert<Equal<Hidden<21>[0],42>>;
type Recursive<T> = Recursive<T>[];
type RecursiveObject<T> = {next:RecursiveObject<T>;value:T};
type RecursiveUse = Recursive<string>;
type RecursiveField=Assert<Equal<RecursiveObject<42>["next"]["value"],42>>;

type WithDefault<T,U=Twice<T>>="done";
type DefaultNative<T>=WithDefault<T>;
type DefaultResult=Assert<Equal<DefaultNative<21>,"done">>;
typefn UseDefault<T>{return WithDefault<T>;}
type DefaultDSL<T>=UseDefault<T>;
type DefaultDSLResult=Assert<Equal<DefaultDSL<21>,"done">>;
interface DefaultContainer<T,U=Twice<T>> {tag:"container";}
type Container<T>=DefaultContainer<T>;
type ContainerResult=Assert<Equal<Container<21>["tag"],"container">>;
