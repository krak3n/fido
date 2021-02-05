package fido

import "testing"

func TestFizz(t *testing.T) {
	t.Parallel()

	type tc struct {
		want string
	}

	cases := map[string]tc{
		"ReturnsBuzz": {
			want: "buzz",
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Fizz()

			if got != tc.want {
				t.Errorf("want %s, got %s", tc.want, got)
			}
		})
	}
}
