import type { FrozenBox, Reverse, Twice } from "./types";

// Hover on Twice or enable typefn result hints to see 42.
export type Answer = Twice<21>;
export const answer: Answer = 42;

// N stays symbolic here. Native specialization replays the calculation with 21.
export type Deferred<N> = Twice<N>;
export type DeferredAnswer = Deferred<21>;
export const deferredAnswer: DeferredAnswer = 42;

// Native tuples, unions, object Types and generic aliases flow through the VM.
export type Reversed = Reverse<["start", 1 | 2, { id: number }]>;
export const reversed: Reversed = [{ id: 7 }, 1, "start"];

export type Wrapped = FrozenBox<Answer>;
export const wrapped: Wrapped = { value: answer };

// Go to definition and rename work through the import. Changing the arithmetic
// in Twice changes both hover results and native assignment diagnostics here.
