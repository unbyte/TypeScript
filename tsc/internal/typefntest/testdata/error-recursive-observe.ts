export {};
typefn Bad<> {return types.recursive<typefn<Self> {return Self==Self;}>;}
type Result=Bad<>;
