export {};
function outer() {
 typefn Bad<>:type {type captured=type(typeof this);fail<"invalid native context executed">;}
 let result!:Bad<>;return result;
}
