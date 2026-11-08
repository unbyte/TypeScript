export {};
typefn Fail<T>: type {fail<"bound executed">;}
typefn Apply<T extends Fail<1>> {return T;}
type Invalid=Apply<1>;
