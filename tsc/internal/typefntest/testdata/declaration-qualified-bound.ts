export interface Open { first: string }
typefn Shape<T extends object> { return type({ [K in keyof T]-?: unknown }); }
namespace Api { export declare function bounded<T extends Shape<Open>>(): string; }
export const result = Api.bounded<{first: string}>();
