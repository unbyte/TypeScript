export {};
typefn Box<T> {return [T];}
type F=Box;
type Bad=[F];
