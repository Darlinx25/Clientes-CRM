package controllers

import (
	"strings"
	"testing"

	"meerkat/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func companyTypeByName(name string) models.CompanyType {
	return models.CompanyType{Model: gorm.Model{ID: uint(1 + len(name))}, Name: name}
}

func foldTypeLookup(types ...models.CompanyType) map[string]models.CompanyType {
	m := make(map[string]models.CompanyType, len(types))
	for _, t := range types {
		m[foldTerm(strings.TrimSpace(t.Name))] = t
	}
	return m
}

func TestMatchCompanyTypeByFoldedName(t *testing.T) {
	lookup := foldTypeLookup(
		companyTypeByName("Administración"),
		companyTypeByName("Serv Dom"),
	)

	// Accent- and case-insensitive exact name matches.
	for _, raw := range []string{"Administración", "Administracion", "administracion", "ADMINISTRACION"} {
		got, ok := matchCompanyType(lookup, raw)
		require.True(t, ok, "expected %q to match Administración", raw)
		assert.Equal(t, "Administración", got.Name)
	}

	_, ok := matchCompanyType(lookup, "Tecnología")
	assert.False(t, ok)
}

func TestMatchCompanyTypeByAlias(t *testing.T) {
	lookup := foldTypeLookup(
		companyTypeByName("Administración"),
		companyTypeByName("Serv Dom"),
		companyTypeByName("Imp Pat"),
		companyTypeByName("PJExt"),
		companyTypeByName("RtasFinExter"),
	)

	cases := map[string]string{
		"ADMIN":             "Administración",
		"Admin":             "Administración",
		"SD":                "Serv Dom",
		"sd":                "Serv Dom",
		"IMPAT":             "Imp Pat",
		"IMP PAT":           "Imp Pat", // matches the app name "Imp Pat" folded
		"PJEXT":             "PJExt",
		"PJ EXT":            "PJExt", // alias after punctuation collapse
		"RtasFinExter":      "RtasFinExter",
		"RTASFINEXTER":      "RtasFinExter",
		"Rentas Exterior":   "RtasFinExter", // legacy long form
		" RENTAS EXTERIOR ": "RtasFinExter",
	}
	for raw, want := range cases {
		got, ok := matchCompanyType(lookup, raw)
		require.True(t, ok, "expected %q to be recognized", raw)
		assert.Equal(t, want, got.Name, "raw=%q", raw)
	}
}

func TestMatchCompanyTypeUnknown(t *testing.T) {
	lookup := foldTypeLookup(companyTypeByName("Serv Dom"))

	for _, raw := range []string{"", "   ", "Monotributo", "TRANSPORTE", "SDX"} {
		_, ok := matchCompanyType(lookup, raw)
		assert.False(t, ok, "raw=%q must not match", raw)
	}
}

func TestResolveTypeStringsDedupsByID(t *testing.T) {
	servDom := companyTypeByName("Serv Dom")
	impPat := companyTypeByName("Imp Pat")
	lookup := foldTypeLookup(servDom, impPat)

	// "SD" and "SERV.DOM" both resolve to Serv Dom; the result must not repeat it.
	list := resolveTypeStrings(lookup, []string{"SD", "SERV.DOM", "IMPAT"})
	require.Len(t, list, 2)
	names := []string{list[0].Name, list[1].Name}
	assert.ElementsMatch(t, []string{"Serv Dom", "Imp Pat"}, names)
}

func TestCompactTerm(t *testing.T) {
	for raw, want := range map[string]string{
		"IMP PAT":      "imppat",
		"imp.pat":      "imppat",
		"IMPAT":        "impat",
		"PJ EXT":       "pjext",
		"RtasFinExter": "rtasfinexter",
		"Serv.Dom":     "servdom",
	} {
		assert.Equal(t, want, compactTerm(raw), "raw=%q", raw)
	}
}
