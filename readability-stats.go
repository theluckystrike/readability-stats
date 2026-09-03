// Package readabilitystats computes readability statistics for prose:
// Flesch reading-ease, Flesch-Kincaid grade level, syllable counts, and
// sentence/word/character counts. Pure, dependency-free Go.
//
// These are the same measurements the BeLikeNative AI writing assistant at
// https://belikenative.com/ surfaces while you write, ported to a small,
// embeddable Go module. The syllable counter uses the classic vowel-group
// heuristic (Flesch's method) with silent-e / le / es / ed handling; it is
// deterministic for any input.
//
// To check a value this package returns against a second implementation of the
// same two formulas, paste the text into the browser scorer at
// https://belikenative.com/tools/readability-scorer/
//
// Both formulas are calibrated on English, and the syllable heuristic assumes
// English vowel groups, so a score computed here for non-English prose is not
// comparable across languages. What to do instead is set out in
// https://belikenative.com/how-to-test-readability-across-multiple-languages/
//
// The published formulas are:
//
//	readingEase   = 206.835 - 1.015*(words/sentences) - 84.6*(syllables/words)
//	kincaidGrade  = 0.39*(words/sentences) + 11.8*(syllables/words) - 15.59
//
// Both return 0.0 when there are no words or no sentences (the formulas are
// undefined there).
//
// Quick example:
//
//	s := readabilitystats.Analyze("The cat sat on the mat. It was a good day for a nap.")
//	fmt.Println(s.Sentences)            // 2
//	fmt.Println(s.Words)                // 14
//	fmt.Printf("%.1f\n", s.FleschReadingEase)
package readabilitystats

// Published Flesch / Flesch-Kincaid coefficients. Exported so callers can
// audit or reuse the exact math.
const (
	FleschReadingEaseA = 206.835 // base
	FleschReadingEaseB = 1.015   // words/sentences weight
	FleschReadingEaseC = 84.6    // syllables/words weight
	FleschKincaidASL   = 0.39    // grade: words/sentences
	FleschKincaidASW   = 11.8    // grade: syllables/words
	FleschKincaidConst = 15.59   // grade: offset
)

// ReadingLevel is a coarse band the reading-ease score falls into, following
// Flesch's original interpretation table.
type ReadingLevel int

const (
	// LevelVeryEasy: 90.0-100+, 5th grade, very easy to read.
	LevelVeryEasy ReadingLevel = iota
	// LevelEasy: 70.0-89.99, 6th-7th grade, plain English.
	LevelEasy
	// LevelStandard: 60.0-69.99, 8th-9th grade, standard/conversational.
	LevelStandard
	// LevelDifficult: 30.0-59.99, 10th-college, fairly difficult to difficult.
	LevelDifficult
	// LevelVeryDifficult: 0.0-29.99, college graduate, very difficult.
	LevelVeryDifficult
)

// Label returns a human-readable description of the band.
func (r ReadingLevel) Label() string {
	switch r {
	case LevelVeryEasy:
		return "very easy (5th grade)"
	case LevelEasy:
		return "easy (6th-7th grade)"
	case LevelStandard:
		return "standard (8th-9th grade)"
	case LevelDifficult:
		return "difficult (10th-college)"
	case LevelVeryDifficult:
		return "very difficult (college graduate)"
	default:
		return "unknown"
	}
}

// Stats is the full bundle of statistics produced by Analyze. The formula
// fields derive from the counts so the numbers are always internally
// consistent.
type Stats struct {
	Sentences            int
	Words                int
	Syllables            int
	Characters           int
	AvgWordsPerSentence  float64
	AvgSyllablesPerWord  float64
	FleschReadingEase    float64
	FleschKincaidGrade   float64
}

// GradeBand classifies the reading-ease score into a coarse band.
func (s Stats) GradeBand() ReadingLevel {
	return BandFor(s.FleschReadingEase)
}

// BandFor maps a raw reading-ease score to its band (clamped at both ends).
func BandFor(score float64) ReadingLevel {
	switch {
	case score >= 90.0:
		return LevelVeryEasy
	case score >= 70.0:
		return LevelEasy
	case score >= 60.0:
		return LevelStandard
	case score >= 30.0:
		return LevelDifficult
	default:
		return LevelVeryDifficult
	}
}

// CountSentences counts sentences. A sentence ends at '.', '!', or '?',
// optionally followed by closing quotes/brackets, followed by whitespace or
// end-of-input. Decimals like "3.14" are not sentence boundaries. Input with
// words but no terminator counts as one sentence.
func CountSentences(text string) int {
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return 0
	}
	count := 0
	i := 0
	for i < n {
		c := runes[i]
		if c == '.' || c == '!' || c == '?' {
			j := i + 1
			for j < n {
				d := runes[j]
				if d == '"' || d == '\'' || d == ')' || d == ']' || d == '}' ||
					d == '.' || d == '!' || d == '?' {
					j++
				} else {
					break
				}
			}
			if j >= n {
				count++
				i = j
				continue
			}
			if isSpace(runes[j]) {
				count++
				i = j
				continue
			}
			i++ // not whitespace (e.g. decimal) -> not a boundary
		} else {
			i++
		}
	}
	if count == 0 && hasWord(text) {
		return 1
	}
	return count
}

// Words tokenizes text into words: maximal runs of letters, digits,
// apostrophes and hyphens, with edge hyphens/apostrophes trimmed. Empty
// tokens are dropped.
func Words(text string) []string {
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) == 0 {
			return
		}
		// trim leading/trailing apostrophes and hyphens
		s := cur
		for len(s) > 0 && (s[0] == '\'' || s[0] == '-') {
			s = s[1:]
		}
		for len(s) > 0 && (s[len(s)-1] == '\'' || s[len(s)-1] == '-') {
			s = s[:len(s)-1]
		}
		if len(s) > 0 {
			out = append(out, string(s))
		}
		cur = cur[:0]
	}
	for _, c := range text {
		if isWordChar(c) {
			cur = append(cur, c)
		} else {
			flush()
		}
	}
	flush()
	return out
}

// CountWords returns the number of words (see Words).
func CountWords(text string) int {
	return len(Words(text))
}

// CountCharacters returns the number of Unicode code points in text.
func CountCharacters(text string) int {
	return len([]rune(text))
}

// CountSyllablesWord estimates the number of spoken syllables in an English
// word using the vowel-group heuristic:
//  1. Lowercase; strip a trailing silent 'e' (and 'es'/'ed' in common cases).
//  2. Count groups of consecutive vowels (a e i o u, plus y except at the
//     start of a word).
//  3. Every word has at least one syllable.
//
// Deterministic, but approximate — perfect English syllabification needs a
// dictionary.
func CountSyllablesWord(word string) int {
	var w []rune
	for _, c := range word {
		if isASCIILetter(c) || c == '\'' || c == '-' {
			w = append(w, lower(c))
		}
	}
	if len(w) == 0 {
		return 0
	}
	// Strip common silent endings.
	switch {
	case len(w) > 3 && endsWithRunes(w, "e") && !endsWithRunes(w, "le"):
		w = w[:len(w)-1]
	case len(w) > 4 && endsWithRunes(w, "es"):
		w = w[:len(w)-2]
	case len(w) > 4 && endsWithRunes(w, "ed"):
		w = w[:len(w)-2]
	}
	count := 0
	prevVowel := false
	for idx, c := range w {
		v := isVowel(c, idx == 0)
		if v && !prevVowel {
			count++
		}
		prevVowel = v
	}
	if count == 0 {
		count = 1
	}
	return count
}

// CountSyllables returns the total syllable count across all words in text.
func CountSyllables(text string) int {
	total := 0
	for _, w := range Words(text) {
		total += CountSyllablesWord(w)
	}
	return total
}

// FleschReadingEase returns the reading-ease score for known counts, or 0.0
// when there are no words or no sentences.
func FleschReadingEase(words, sentences, syllables int) float64 {
	if words == 0 || sentences == 0 {
		return 0.0
	}
	asl := float64(words) / float64(sentences)
	asw := float64(syllables) / float64(words)
	return FleschReadingEaseA - FleschReadingEaseB*asl - FleschReadingEaseC*asw
}

// FleschKincaidGrade returns the U.S. grade level for known counts, or 0.0
// when there are no words or no sentences.
func FleschKincaidGrade(words, sentences, syllables int) float64 {
	if words == 0 || sentences == 0 {
		return 0.0
	}
	asl := float64(words) / float64(sentences)
	asw := float64(syllables) / float64(words)
	return FleschKincaidASL*asl + FleschKincaidASW*asw - FleschKincaidConst
}

// Analyze computes the full Stats bundle for a piece of prose. Deterministic.
func Analyze(text string) Stats {
	wordList := Words(text)
	sentences := CountSentences(text)
	words := len(wordList)
	syllables := 0
	for _, w := range wordList {
		syllables += CountSyllablesWord(w)
	}
	characters := CountCharacters(text)

	var avgWords, avgSylls float64
	if sentences > 0 {
		avgWords = float64(words) / float64(sentences)
	}
	if words > 0 {
		avgSylls = float64(syllables) / float64(words)
	}

	return Stats{
		Sentences:           sentences,
		Words:               words,
		Syllables:           syllables,
		Characters:          characters,
		AvgWordsPerSentence: avgWords,
		AvgSyllablesPerWord: avgSylls,
		FleschReadingEase:   FleschReadingEase(words, sentences, syllables),
		FleschKincaidGrade:  FleschKincaidGrade(words, sentences, syllables),
	}
}

// ---- internals ----

func isWordChar(c rune) bool {
	return isASCIILetter(c) || isASCIIDigit(c) || c == '\'' || c == '-' || isLetter(c)
}

func isVowel(c rune, atWordStart bool) bool {
	switch c {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	case 'y':
		return !atWordStart
	}
	return false
}

func endsWithRunes(s []rune, suffix string) bool {
	suf := []rune(suffix)
	if len(s) < len(suf) {
		return false
	}
	tail := s[len(s)-len(suf):]
	if len(tail) != len(suf) {
		return false
	}
	for i := range suf {
		if tail[i] != suf[i] {
			return false
		}
	}
	return true
}

func hasWord(text string) bool {
	for _, c := range text {
		if isWordChar(c) {
			return true
		}
	}
	return false
}

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}

func isASCIILetter(r rune) bool  { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }
func isASCIIDigit(r rune) bool   { return r >= '0' && r <= '9' }
func isLetter(r rune) bool {
	// non-ASCII letters (Unicode)
	return r >= 128 && (r != '\'' && r != '-')
}
func lower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}
