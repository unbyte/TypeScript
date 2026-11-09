/** @internal */
typefn Hidden<T> { return T; }
export interface Surface { value: Hidden<1> }
export declare function read(): Surface;
