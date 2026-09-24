package controllers

import "strings"

var accentFoldPairs = []struct{ from, to string }{
	{"Á", "a"}, {"á", "a"}, {"À", "a"}, {"à", "a"}, {"Â", "a"}, {"â", "a"}, {"Ä", "a"}, {"ä", "a"}, {"Ã", "a"}, {"ã", "a"}, {"Å", "a"}, {"å", "a"},
	{"É", "e"}, {"é", "e"}, {"È", "e"}, {"è", "e"}, {"Ê", "e"}, {"ê", "e"}, {"Ë", "e"}, {"ë", "e"},
	{"Í", "i"}, {"í", "i"}, {"Ì", "i"}, {"ì", "i"}, {"Î", "i"}, {"î", "i"}, {"Ï", "i"}, {"ï", "i"},
	{"Ó", "o"}, {"ó", "o"}, {"Ò", "o"}, {"ò", "o"}, {"Ô", "o"}, {"ô", "o"}, {"Ö", "o"}, {"ö", "o"}, {"Õ", "o"}, {"õ", "o"},
	{"Ú", "u"}, {"ú", "u"}, {"Ù", "u"}, {"ù", "u"}, {"Û", "u"}, {"û", "u"}, {"Ü", "u"}, {"ü", "u"},
	{"Ñ", "n"}, {"ñ", "n"},
	{"Ç", "c"}, {"ç", "c"},
}

var accentFoldMap = func() map[rune]rune {
	m := make(map[rune]rune, len(accentFoldPairs))
	for _, p := range accentFoldPairs {
		m[[]rune(p.from)[0]] = []rune(p.to)[0]
	}
	return m
}()

// accentFoldExpr folds accented characters in a SQL expression to their ASCII
// base and lowercases, so LIKE matches regardless of tildes.
func accentFoldExpr(col string) string {
	expr := col
	for _, p := range accentFoldPairs {
		expr = "REPLACE(" + expr + ", '" + p.from + "', '" + p.to + "')"
	}
	return "LOWER(" + expr + ")"
}

// foldTerm folds a search term (lowercase, no accents) to match the columns
// produced by accentFoldExpr.
func foldTerm(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 128 {
			if r >= 'A' && r <= 'Z' {
				return r + 32
			}
			return r
		}
		if folded, ok := accentFoldMap[r]; ok {
			return folded
		}
		return r
	}, s)
}