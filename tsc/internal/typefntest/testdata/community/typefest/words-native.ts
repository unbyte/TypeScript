// Original type-fest scanner, with its finite-character dependencies specialized.
// No DSL calls occur in this comparison implementation. See ORIGIN.md and LICENSE.
import type { WordSeparators, AsciiPunctuation } from "./characters";
type WordsOptions = { splitOnNumbers?: boolean; splitOnPunctuation?: boolean };
type IsNumeric<T extends string> = T extends `${number}` ? true : false;
type IsUppercase<T extends string> = T extends Uppercase<T> ? true : false;
type IsLowercase<T extends string> = T extends Lowercase<T> ? true : false;
export type WordsNative<S extends string, N extends boolean, P extends boolean> =
  _Words<S, { splitOnNumbers: N; splitOnPunctuation: P }>;

type SkipEmptyWord<Word extends string> = Word extends '' ? [] : [Word];

type RemoveLastCharacter<
	Sentence extends string,
	Character extends string,
> = Sentence extends `${infer LeftSide}${Character}`
	? SkipEmptyWord<LeftSide>
	: never;

export type _Words<
	Sentence extends string,
	Options extends Required<WordsOptions>,
	LastCharacter extends string = '',
	CurrentWord extends string = '',
	// The words found so far are kept in `Result`, so the recursion is in tail position and a long sentence does not hit the recursion limit.
	Result extends string[] = [],
> = Sentence extends `${infer FirstCharacter}${infer RemainingCharacters}`
	? FirstCharacter extends WordSeparators | (Options['splitOnPunctuation'] extends true ? AsciiPunctuation : never)
		// Skip word separator
		? _Words<RemainingCharacters, Options, '', '', [...Result, ...SkipEmptyWord<CurrentWord>]>
		: LastCharacter extends ''
			// First character of word
			? _Words<RemainingCharacters, Options, FirstCharacter, FirstCharacter, Result>
			// Case change between numeric and non-numeric characters.
			: [IsNumeric<LastCharacter>, IsNumeric<FirstCharacter>] extends [false, true] | [true, false]
				? Options['splitOnNumbers'] extends true
					// Split on number: push word
					? _Words<RemainingCharacters, Options, FirstCharacter, FirstCharacter, [...Result, ...SkipEmptyWord<CurrentWord>]>
					// No split on number: concat word
					: _Words<RemainingCharacters, Options, FirstCharacter, `${CurrentWord}${FirstCharacter}`, Result>
				// No case change: concat word
				: [true, true] extends [IsNumeric<LastCharacter>, IsNumeric<FirstCharacter>]
					? _Words<RemainingCharacters, Options, FirstCharacter, `${CurrentWord}${FirstCharacter}`, Result>
				// Case change: lower to upper, push word
					: [true, true] extends [IsLowercase<LastCharacter>, IsUppercase<FirstCharacter>]
						// Only a capital letter starts a new word. A character with no case at all, such as a Han character or an emoji, satisfies both case predicates but is its own lowercase form, so it is not one.
						? IsCapitalLetter<FirstCharacter> extends true
							? _Words<RemainingCharacters, Options, FirstCharacter, FirstCharacter, [...Result, ...SkipEmptyWord<CurrentWord>]>
							: _Words<RemainingCharacters, Options, FirstCharacter, `${CurrentWord}${FirstCharacter}`, Result>
					// Case change: upper to lower, brings back the last character, push word
						: [true, true] extends [IsUppercase<LastCharacter>, IsLowercase<FirstCharacter>]
							// Only a capital followed by a lowercase letter starts a new word. Punctuation or a character with no case after an acronym, as in `userID?`, satisfies both case predicates, so it would otherwise split off the last capital.
							? [IsCapitalLetter<LastCharacter>, IsLowercaseLetter<FirstCharacter>] extends [true, true]
								? _Words<
									RemainingCharacters, Options, FirstCharacter, `${LastCharacter}${FirstCharacter}`,
									[...Result, ...RemoveLastCharacter<CurrentWord, LastCharacter>]
								>
								: _Words<RemainingCharacters, Options, FirstCharacter, `${CurrentWord}${FirstCharacter}`, Result>
						// No case change: concat word
							: _Words<RemainingCharacters, Options, FirstCharacter, `${CurrentWord}${FirstCharacter}`, Result>
	: [...Result, ...SkipEmptyWord<CurrentWord>];

/**
Whether the given character is a capital letter, which is what decides whether it begins a new word.

A character with no case at all, such as a Han character or an emoji, is its own lowercase form and so is not a capital. Every other cased character is, in any script, so this is deliberately not restricted to ASCII: `É` and `Ä` are capitals just as `A` is.
*/
type IsCapitalLetter<Character extends string> = Lowercase<Character> extends Character ? false : true;

/**
Whether the given character is a lowercase letter, which is what decides whether a preceding capital begins a new word.

Like {@link IsCapitalLetter}, a character with no case at all, such as punctuation, a Han character, or an emoji, is its own uppercase form and so is not a lowercase letter.
*/
type IsLowercaseLetter<Character extends string> = Uppercase<Character> extends Character ? false : true;

export {};

