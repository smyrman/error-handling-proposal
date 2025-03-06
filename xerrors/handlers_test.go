package xerrors_test

import (
	"testing"

	"github.com/smyrman/error-handling-proposal/xerrors"
)

func TestWrap(t *testing.T) {
	const (
		fieldA = "a"
		fieldB = "b"
	)
	parse := func(in transportModel) (_ BusinessModel, _err error) {
		defer xerrors.Catch(&_err)
		const id = "test"

		return BusinessModel{
			A: xerrors.Must2(ParseA(in.A))(xerrors.Wrap("%s: %w", fieldA)),
			B: xerrors.Must2(ParseB(in.B))(xerrors.Wrap("%[1]s: %[3]w (id:%[2]s)", fieldB, id)),
		}, nil
	}

	t.Run("fail a", func(t *testing.T) {
		expect := "a: length must be in range 3-10"
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
		expect := "b: must be in range 5-100 (id:test)"
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
