typefn Bad<> { return fail<"bad">; }
typefn Arity<> { if (false) { failAt<"bad">; } return 1; }
