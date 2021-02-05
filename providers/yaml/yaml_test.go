package yaml

import "testing"

func TestBar(t *testing.T) {
	t.Parallel()

	type tc struct {
		want string
	}

	cases := map[string]tc{
		"ReturnsFoo": {
			want: "foo",
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Bar()

			if got != tc.want {
				t.Errorf("want %s, got %s", tc.want, got)
			}
		})
	}
}
