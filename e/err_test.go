package e

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

var errTests = []struct {
	err    error
	tester func(error) bool
	want   bool
}{
	{
		err:    ErrFull(fmt.Errorf("kaboom!")),
		tester: IsFull,
		want:   true,
	},
	{
		err:    ErrFull(errors.New("kaboom!")),
		tester: IsFull,
		want:   true,
	},
	{
		err:    errors.Wrap(ErrFull(errors.New("kaboom!")), "full!"),
		tester: IsFull,
		want:   true,
	},
	{
		err:    errors.New("kaboom!"),
		tester: IsFull,
		want:   false,
	},
	{
		err:    ErrNotFound(fmt.Errorf("kaboom!")),
		tester: IsNotFound,
		want:   true,
	},
	{
		err:    ErrNotFound(errors.New("kaboom!")),
		tester: IsNotFound,
		want:   true,
	},
	{
		err:    errors.Wrap(ErrNotFound(errors.New("kaboom!")), "not found!"),
		tester: IsNotFound,
		want:   true,
	},
	{
		err:    errors.New("kaboom!"),
		tester: IsNotFound,
		want:   false,
	},
	{
		err:    ErrInvalid(fmt.Errorf("kaboom!")),
		tester: IsInvalid,
		want:   true,
	},
	{
		err:    ErrInvalid(errors.New("kaboom!")),
		tester: IsInvalid,
		want:   true,
	},
	{
		err:    errors.Wrap(ErrInvalid(errors.New("kaboom!")), "not found!"),
		tester: IsInvalid,
		want:   true,
	},
	{
		err:    errors.New("kaboom!"),
		tester: IsInvalid,
		want:   false,
	},
	{
		err:    ErrInvalid(errors.New("kaboom!")),
		tester: IsNotFound,
		want:   false,
	},
}

func TestErr(t *testing.T) {
	for _, tt := range errTests {
		got := tt.tester(tt.err)
		if got != tt.want {
			t.Errorf("%v %v: got %v, want %v", tt.tester, tt.err, got, tt.want)
		}
	}
}
