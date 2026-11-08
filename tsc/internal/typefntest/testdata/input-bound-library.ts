export typefn Bound<T extends string> {return T;}
export typefn Outer<T> {return Bound<T>;}
