typefn Check<T> {type result=T+1; return 1;}
typefn Factory<T> {return typefn<U> {return U+1;};}
export typefn Waiting<T> {type checked=Check<T>; type F=Factory<checked>; return F<"invalid after creation">;}
