export {};
typefn Bad<> {return types.recursive<typefn<Outer> {
 type unused=types.recursive<typefn<Inner> {return Inner;}>;
 return 42;
}>;}
type Result=Bad<>;
