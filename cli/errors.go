package cli

import (
	"errors"

	"github.com/tamnd/xiaoyuzhou-cli/xiaoyuzhou"
)

func isNotFound(err error) bool {
	return errors.Is(err, xiaoyuzhou.ErrNotFound)
}
