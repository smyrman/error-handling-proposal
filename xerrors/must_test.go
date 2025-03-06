package xerrors_test

import (
	"fmt"
	"testing"

	"github.com/smyrman/error-handling-proposal/xerrors"
)

func TestMust(t *testing.T) {
	t.Run("fail at 3", func(t *testing.T) {
		expect := "number 3 not ok"
		expectLastNumber := 3

		var lastNumber int
		checkNumber := func(i int) error {
			lastNumber = i
			if i >= 3 {
				return fmt.Errorf("number %d not ok", i)
			}
			return nil
		}
		test := func() (_err error) {
			defer xerrors.Catch(&_err)

			xerrors.Must(checkNumber(0))()
			xerrors.Must(checkNumber(1))()
			xerrors.Must(checkNumber(2))()
			xerrors.Must(checkNumber(3))()
			xerrors.Must(checkNumber(4))()

			return nil
		}

		err := test()
		switch {
		case err == nil:
			t.Error("expected error, got nil")
		case err.Error() != expect:
			t.Errorf("unexpected error message;\ngot:  %q\nwant: %q", err.Error(), expect)
		}
		if lastNumber != expectLastNumber {
			t.Errorf("unexpected i value for last run of checkNumber; got %d, want %d", lastNumber, expectLastNumber)
		}
	})
}

func TestMust2(t *testing.T) {
	parse := func(in transportModel) (_ BusinessModel, _err error) {
		defer xerrors.Catch(&_err)

		return BusinessModel{
			A: xerrors.Must2(ParseA(in.A))(),
			B: xerrors.Must2(ParseB(in.B))(),
		}, nil
	}

	t.Run("fail a", func(t *testing.T) {
		expect := "length must be in range 3-10"
		v, err := parse(transportModel{A: "", B: 5})
		switch {
		case err == nil:
			t.Error("expected error, got nil")
		case err.Error() != expect:
			t.Errorf("unexpected error message;\ngot:  %q\nwant: %q", err.Error(), expect)
		}
		if v.A.a != "" {
			t.Errorf("a unexpectedly set to %q", v.A.a)
		}
		if v.B.b != 0 {
			t.Errorf("b unexpectedly set to %q", v.B.b)
		}
	})

	t.Run("fail b", func(t *testing.T) {
		expect := "must be in range 5-100"
		v, err := parse(transportModel{A: "good", B: 101})
		switch {
		case err == nil:
			t.Error("expected error, got nil")
		case err.Error() != expect:
			t.Errorf("unexpected error message;\ngot:  %q\nwant: %q", err.Error(), expect)
		}
		if v.A.a != "" {
			t.Errorf("a unexpectedly set to %q", v.A.a)
		}
		if v.B.b != 0 {
			t.Errorf("b unexpectedly set to %q", v.B.b)
		}
	})
}
