type OnlyString<T extends string> = {value:T};
typefn Use<T> { return type(OnlyString<T>); }
type Result = Use<number>;
