# readability-stats

[![Go Reference](https://pkg.go.dev/badge/github.com/theluckystrike/readability-stats.svg)](https://pkg.go.dev/github.com/theluckystrike/readability-stats)

Flesch reading-ease, Flesch-Kincaid grade level, syllable counts, and
sentence/word/character statistics for prose — pure, dependency-free Go.
The same readability math behind the
[BeLikeNative](https://belikenative.com/) AI writing assistant, ported to a
small, embeddable module.

The published formulas:

- **Reading ease** = `206.835 − 1.015 × (words/sentences) − 84.6 × (syllables/words)`
- **Grade level** = `0.39 × (words/sentences) + 11.8 × (syllables/words) − 15.59`

Both return `0.0` when there are no words or no sentences.

To check a value this package returns against a second implementation of the
same two formulas, paste the text into the browser
[readability scorer](https://belikenative.com/tools/readability-scorer/).

## Install

```sh
go get github.com/theluckystrike/readability-stats
```

## Quick example

```go
package main

import (
	"fmt"

	readabilitystats "github.com/theluckystrike/readability-stats"
)

func main() {
	s := readabilitystats.Analyze("The cat sat on the mat. It was a good day for a nap.")
	fmt.Println(s.Sentences)                          // 2
	fmt.Println(s.Words)                              // 14
	fmt.Printf("%.1f\n", s.FleschReadingEase)         // very high (plain prose)
	fmt.Println(s.GradeBand().Label())                // "very easy (5th grade)"

	// Or call the formulas directly with your own counts.
	re := readabilitystats.FleschReadingEase(119, 6, 210)
	fmt.Printf("%.2f\n", re)                          // 37.41
}
```

## API

| Function | Description |
| --- | --- |
| `Analyze(text string) Stats` | Full stats bundle: counts + both Flesch formulas + averages. |
| `CountSentences(text string) int` | Sentences (`.`, `!`, `?` terminators; decimals ignored). |
| `CountWords(text string) int` | Word count (hyphenated & contracted handled). |
| `CountCharacters(text string) int` | Unicode code-point count. |
| `Words(text string) []string` | Word tokenization. |
| `CountSyllables(text string) int` | Total syllables across all words. |
| `CountSyllablesWord(word string) int` | Syllables for one word (vowel-group heuristic). |
| `FleschReadingEase(words, sentences, syllables int) float64` | Reading-ease score (0–100+). |
| `FleschKincaidGrade(words, sentences, syllables int) float64` | U.S. grade level. |
| `BandFor(score float64) ReadingLevel` | Map a score to its band. |

`Stats` also has a `GradeBand() ReadingLevel` method. `ReadingLevel.Label()`
returns a human-readable string.

## Determinism & limitations

Syllable counting uses the classic vowel-group heuristic with silent-`e` /
`le` / `es` / `ed` handling. It is deterministic for any input but, like all
rule-based syllabifiers, is approximate — perfect English syllabification
needs a dictionary. Sentence splitting is conservative: decimals (`3.14`)
are not sentence boundaries, and trailing quotes are handled.

## License

MIT.

## Links

- **BeLikeNative** — AI writing assistant: <https://belikenative.com/>
- Package docs: <https://pkg.go.dev/github.com/theluckystrike/readability-stats>
