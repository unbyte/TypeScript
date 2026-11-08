export {};
type Array<T>=T|number;
typefn Bad<> {return types.recursive<Array>;}
type Result=Bad<>;
