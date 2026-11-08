export {};
typefn Box<T> {return [T];}
typefn Unused<> {
 type F=Box;
 if(false) {
  type a=[F];
  type b={value:F};
  type c=F|number;
  type d=`${F}`;
  type e=F+1;
  type f=text.length<F>;
  type g=Box<F>;
  type h=Array<F>;
  type i=[...F];
  type j=F[0];
  type k={[F]:1};
  if(F) {return 1;}
  while(F) {break;}
  for(type item of F) {break;}
  match(F) {_=>{}}
  match(1) {captured if(F)=>{} _=>{}}
  match("x") {`${=F}`=>{} _=>{}}
 }
 return 0;
}
