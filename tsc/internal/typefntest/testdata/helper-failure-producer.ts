export typefn Check<T> {
 type saved = T;
 typefn Validate<> {if(saved is string) {return saved;} fail<"helper requires string">;}
 return Validate<>;
}
export type Pending<T> = Check<T>;
