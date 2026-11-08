export {};
typefn Bad<> {return types.recursive<typefn<Outer> {
 type unused=types.recursive<typefn<Inner> {return Inner;}>;
 return {next:Outer};
}>;}
type Result=Bad<>;
