package json

import "testing"

func TestFoo(t *testing.T) {
	t.Parallel()

	type tc struct {
		want string
	}

	cases := map[string]tc{
		"ReturnsBar": {
			want: "bar",
		},
	}

	for name, testCase := range cases {
		tc := testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Foo()

			if got != tc.want {
				t.Errorf("want %s, got %s", tc.want, got)
			}
		})
	}
}
