package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/smyrman/error-handling-proposal/xerrors"
)

func main() {
	defer xerrors.Catch(nil)

	ab := xerrors.Must2(ParseAB("a", 0))(logAndExit)
	fmt.Println("AB:", ab)
}

func logAndExit(err error) error {
	slog.Error(err.Error())
	os.Exit(1)
	return nil
}

type AB struct {
	A A
	B B
}

func ParseAB(a string, b int) (_ AB, _err error) {
	defer xerrors.Catch(&_err)

	return AB{
		A: xerrors.Must2(ParseA(a))(xerrors.Wrap("a: %w")),
		B: xerrors.Must2(ParseB(b))(xerrors.Wrap("b: %w")),
	}, nil
}
