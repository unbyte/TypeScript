typefn Count<T> {type i=0;while(i<T){i=i+1;}return i;}
type Alias<T>=Count<T>;
type TooMuch=Alias<100>;
