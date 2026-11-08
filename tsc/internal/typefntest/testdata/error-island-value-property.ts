export {};
const input={present:1};
typefn F<> {return type(typeof input.absent);}
type Invalid=F<>;
