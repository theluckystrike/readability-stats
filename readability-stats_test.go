package readabilitystats

import (
	"math"
	"testing"
)

func approx(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// ---- syllables ----------------------------------------------------------

func TestSyllableSimpleWords(t *testing.T) {
	cases := map[string]int{"cat": 1, "dog": 1, "the": 1, "mat": 1}
	for w, want := range cases {
		if got := CountSyllablesWord(w); got != want {
			t.Errorf("CountSyllablesWord(%q) = %d, want %d", w, got, want)
		}
	}
}

func TestSyllableTwoSyllable(t *testing.T) {
	cases := map[string]int{"happy": 2, "reading": 2, "table": 2, "apple": 2}
	for w, want := range cases {
		if got := CountSyllablesWord(w); got != want {
			t.Errorf("CountSyllablesWord(%q) = %d, want %d", w, got, want)
		}
	}
}

func TestSyllableThreePlus(t *testing.T) {
	cases := map[string]int{"beautiful": 3, "calculate": 3, "readability": 5}
	for w, want := range cases {
		if got := CountSyllablesWord(w); got != want {
			t.Errorf("CountSyllablesWord(%q) = %d, want %d", w, got, want)
		}
	}
}

func TestSyllableSilentE(t *testing.T) {
	cases := map[string]int{"name": 1, "time": 1, "home": 1}
	for w, want := range cases {
		if got := CountSyllablesWord(w); got != want {
			t.Errorf("CountSyllablesWord(%q) = %d, want %d", w, got, want)
		}
	}
}

func TestSyllableLeEnding(t *testing.T) {
	cases := map[string]int{"apple": 2, "table": 2, "little": 2}
	for w, want := range cases {
		if got := CountSyllablesWord(w); got != want {
			t.Errorf("CountSyllablesWord(%q) = %d, want %d", w, got, want)
		}
	}
}

func TestSyllableYAsVowel(t *testing.T) {
	// 'y' at start is treated as consonant; elsewhere vowel.
	if got := CountSyllablesWord("yellow"); got != 2 {
		t.Errorf("yellow = %d, want 2", got)
	}
	if got := CountSyllablesWord("happy"); got != 2 {
		t.Errorf("happy = %d, want 2", got)
	}
	// "rhythm": leading y treated as consonant -> 0 vowel groups -> floor 1.
	if got := CountSyllablesWord("rhythm"); got != 1 {
		t.Errorf("rhythm = %d, want 1", got)
	}
}

func TestSyllableEmptyAndNonalpha(t *testing.T) {
	if got := CountSyllablesWord(""); got != 0 {
		t.Errorf("empty = %d, want 0", got)
	}
	if got := CountSyllablesWord("123"); got != 0 {
		t.Errorf("123 = %d, want 0", got)
	}
}

func TestSyllablePunctOnlyFloorsAtOne(t *testing.T) {
	if got := CountSyllablesWord("---"); got != 1 {
		t.Errorf("--- = %d, want 1", got)
	}
}

// ---- words & sentences --------------------------------------------------

func TestWordsBasic(t *testing.T) {
	cases := map[string]int{
		"The cat sat on the mat.": 6,
		"Hello, world!":           2,
		"don't stop":              2,
		"well-known author":       2,
	}
	for text, want := range cases {
		if got := CountWords(text); got != want {
			t.Errorf("CountWords(%q) = %d, want %d", text, got, want)
		}
	}
}

func TestWordsEmptyAndPunct(t *testing.T) {
	for _, text := range []string{"", "...!!!", "   "} {
		if got := CountWords(text); got != 0 {
			t.Errorf("CountWords(%q) = %d, want 0", text, got)
		}
	}
}

func TestSentencesBasic(t *testing.T) {
	cases := map[string]int{
		"Hello world.":         1,
		"Hi. Hello! Howdy?":    3,
		"One sentence only":    1,
		"":                     0,
	}
	for text, want := range cases {
		if got := CountSentences(text); got != want {
			t.Errorf("CountSentences(%q) = %d, want %d", text, got, want)
		}
	}
}

func TestSentencesIgnoresDecimal(t *testing.T) {
	if got := CountSentences("Pi is 3.14 and that is fine."); got != 1 {
		t.Errorf("decimal split = %d, want 1", got)
	}
}

func TestSentencesTrailingQuotes(t *testing.T) {
	if got := CountSentences(`He said "hi." Then he left.`); got != 2 {
		t.Errorf("trailing quotes = %d, want 2", got)
	}
}

// ---- formulas -----------------------------------------------------------

func TestFleschReadingEasePlainProse(t *testing.T) {
	// 6 words, 1 sentence, 6 syllables -> 116.145
	re := FleschReadingEase(6, 1, 6)
	if !approx(re, 116.145, 1e-3) {
		t.Fatalf("reading ease = %v, want 116.145", re)
	}
}

func TestFleschReadingEaseTextbookValue(t *testing.T) {
	// 119 words, 6 sentences, 210 syllables -> 37.410049...
	re := FleschReadingEase(119, 6, 210)
	if !approx(re, 37.41004901960787, 1e-3) {
		t.Fatalf("reading ease = %v, want 37.410049", re)
	}
}

func TestFleschKincaidTextbookValue(t *testing.T) {
	// 119 words, 6 sentences, 210 syllables -> 12.968529...
	fk := FleschKincaidGrade(119, 6, 210)
	if !approx(fk, 12.968529411764706, 1e-3) {
		t.Fatalf("kincaid = %v, want 12.968529", fk)
	}
}

func TestFormulasZeroOnDegenerateInput(t *testing.T) {
	if FleschReadingEase(0, 1, 0) != 0.0 {
		t.Fatal("expected 0")
	}
	if FleschReadingEase(5, 0, 5) != 0.0 {
		t.Fatal("expected 0")
	}
	if FleschKincaidGrade(0, 1, 0) != 0.0 {
		t.Fatal("expected 0")
	}
	if FleschKincaidGrade(5, 0, 5) != 0.0 {
		t.Fatal("expected 0")
	}
}

// ---- analyze end-to-end -------------------------------------------------

func TestAnalyzeSimpleSentence(t *testing.T) {
	s := Analyze("The cat sat on the mat.")
	if s.Sentences != 1 || s.Words != 6 || s.Syllables != 6 {
		t.Fatalf("counts = %+v", s)
	}
	if s.Characters != 23 {
		t.Errorf("characters = %d, want 23", s.Characters)
	}
	if !approx(s.AvgWordsPerSentence, 6.0, 1e-9) {
		t.Errorf("ASL = %v", s.AvgWordsPerSentence)
	}
	if s.FleschReadingEase <= 100.0 {
		t.Errorf("expected >100, got %v", s.FleschReadingEase)
	}
	if s.GradeBand() != LevelVeryEasy {
		t.Errorf("band = %v, want VeryEasy", s.GradeBand())
	}
}

func TestAnalyzeTwoSentences(t *testing.T) {
	// 6 + 8 = 14 words over 2 sentences.
	s := Analyze("The cat sat on the mat. It was a good day for a nap.")
	if s.Sentences != 2 || s.Words != 14 {
		t.Fatalf("counts = %+v", s)
	}
	if !approx(s.AvgWordsPerSentence, 7.0, 1e-9) {
		t.Errorf("ASL = %v", s.AvgWordsPerSentence)
	}
	if s.GradeBand() != LevelVeryEasy {
		t.Errorf("band = %v, want VeryEasy", s.GradeBand())
	}
}

func TestAnalyzeEmpty(t *testing.T) {
	s := Analyze("")
	if s.Sentences != 0 || s.Words != 0 || s.Syllables != 0 {
		t.Fatalf("counts = %+v", s)
	}
	if s.FleschReadingEase != 0.0 || s.FleschKincaidGrade != 0.0 {
		t.Errorf("formulas not zero: %+v", s)
	}
}

func TestAnalyzeDeterministic(t *testing.T) {
	text := "Readability is the ease with which a reader can understand a written text. " +
		"The readability of a text depends on its vocabulary, sentence length, and complexity."
	a := Analyze(text)
	b := Analyze(text)
	if a != b {
		t.Fatalf("non-deterministic:\n%+v\n%+v", a, b)
	}
	if a.Sentences != 2 {
		t.Errorf("sentences = %d, want 2", a.Sentences)
	}
	if a.Syllables <= a.Words {
		t.Errorf("expected some multi-syllable words")
	}
}

// ---- band mapping -------------------------------------------------------

func TestBandBoundaries(t *testing.T) {
	cases := []struct {
		score float64
		want  ReadingLevel
	}{
		{95.0, LevelVeryEasy},
		{90.0, LevelVeryEasy},
		{89.99, LevelEasy},
		{70.0, LevelEasy},
		{65.0, LevelStandard},
		{60.0, LevelStandard},
		{45.0, LevelDifficult},
		{30.0, LevelDifficult},
		{20.0, LevelVeryDifficult},
		{0.0, LevelVeryDifficult},
	}
	for _, c := range cases {
		if got := BandFor(c.score); got != c.want {
			t.Errorf("BandFor(%v) = %v, want %v", c.score, got, c.want)
		}
	}
}

func TestReadingLevelLabels(t *testing.T) {
	if LevelVeryEasy.Label() == "" {
		t.Error("empty label")
	}
	if LevelStandard.Label() == "" {
		t.Error("empty label")
	}
}
