interface Animal { animal:true; }
interface Dog extends Animal { dog:true; }
type Method<T> = {method(value:T):void}['method'];
typefn Rebuild<T> { return types.signature<reflect.signature<T>>; }
declare const dog: Rebuild<Method<Dog>>;
const animal: Rebuild<Method<Animal>> = dog;
