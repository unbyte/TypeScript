import type {Outer} from "./library.js";
type First=Outer<123>;
type Pending<T>=Outer<T>;
type Second=Pending<false>;
