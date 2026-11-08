typefn Outer<> {type bound = string;typefn A<T extends bound> {return T;} return A<123>;} type Result = Outer<>;
