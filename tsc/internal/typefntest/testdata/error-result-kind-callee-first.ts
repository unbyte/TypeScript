export {};
typefn Cycle<T> {return Cycle<T>;}
typefn Client<T> {return Cycle<T>+1;}
