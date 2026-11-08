export {};
typefn Bad<> {return types.recursive<typefn<Self> {return {[Self]:1};}>;}
type Result=Bad<>;
