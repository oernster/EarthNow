package region

import "testing"

func TestFRGLB014_ABareCodeIsTwoLetters(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
		ok       bool
	}{
		{"GB", "GB", true},
		{" gb ", "GB", true},
		{"001", "", false},
		{"G", "", false},
		{"GBR", "", false},
		{"", "", false},
		{"G1", "", false},
	}
	for _, c := range cases {
		if got, ok := Code(c.in); got != c.want || ok != c.ok {
			t.Errorf("Code(%q) = %q, %v; want %q, %v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestFRGLB014_TheTerritoryComesOutOfALocale(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
		ok       bool
	}{
		{"en_GB.UTF-8", "GB", true},
		{"en-GB", "GB", true},
		{"sr_RS@latin", "RS", true},
		{"zh-Hant-TW", "TW", true},
		{"en_US@rg=gbzzzz", "GB", true},
		{"en_US@calendar=gregorian;rg=nozzzz", "NO", true},
		{"en_US@rg=nozzzz;sd=no03", "NO", true},
		{"en_US@rg=z", "US", true},
		{"en_US@rg=1bzzzz", "US", true},
		{"de_DE", "DE", true},
		{"C.UTF-8", "", false},
		{"POSIX", "", false},
		{"C", "", false},
		{"en", "", false},
		{"", "", false},
		{"es-419", "", false},
	}
	for _, c := range cases {
		if got, ok := FromLocale(c.in); got != c.want || ok != c.ok {
			t.Errorf("FromLocale(%q) = %q, %v; want %q, %v", c.in, got, ok, c.want, c.ok)
		}
	}
}
