import type {Signature,Value} from './out/library';
type Assert<T extends true>=T;
type Check=Assert<ReturnType<Signature<>> extends 1 ? true : false>;
type NativeValue=Assert<number extends Value<> ? true : false>;
