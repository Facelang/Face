package prog

import "github.com/facelang/face/internal/tokens"

type Parser interface {
	NextToken() tokens.Token
}
