typefn F<T> { return T; } typefn G<> { if (false) { return F<1, 2>; } return 0; }
