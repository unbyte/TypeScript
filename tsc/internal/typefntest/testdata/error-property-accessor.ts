class Accessor {
 get value():number {return 1;}
 set value(input:string|number) {}
}
typefn Fields<T> {return reflect.properties<T>;}
type Result=Fields<Accessor>;
