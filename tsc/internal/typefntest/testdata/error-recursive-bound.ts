export {};
typefn Bad<> {typefn Build<Self extends string> {return {next:Self};} return types.recursive<Build>;}
type Result=Bad<>;
