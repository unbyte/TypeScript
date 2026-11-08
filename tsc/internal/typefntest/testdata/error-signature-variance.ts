interface Animal { animal: true; }
interface Dog extends Animal { dog: true; }
typefn Forward<T> { return types.overload<reflect.calls<T>>; }
declare const dog: Forward<(x: Dog) => void>;
const animal: Forward<(x: Animal) => void> = dog;
