export {};
typefn Box<T extends string> {return [T];}
type F=Box;
type Bad=F<1>;
