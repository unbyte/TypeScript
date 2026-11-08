export {};
typefn Bad<> {return types.recursive<typefn<Self> {return Self | unknown;}>;}
type Result=Bad<>;
