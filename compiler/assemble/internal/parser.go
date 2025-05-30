package internal

import (
	"io"
	"os"
)

type Parser struct {
	lex           lex.TokenReader
	lineNum       int   // Line number in source file.
	errorLine     int   // Line number of last error.
	errorCount    int   // Number of errors.
	sawCode       bool  // saw code in this file (as opposed to comments and blank lines)
	pc            int64 // virtual PC; count of Progs; doesn't advance for GLOBL or DATA.
	input         []lex.Token
	inputPos      int
	pendingLabels []string // Labels to attach to next instruction.
	labels        map[string]*obj.Prog
	toPatch       []Patch
	addr          []obj.Addr
	arch          *arch.Arch
	ctxt          *obj.Link
	firstProg     *obj.Prog
	lastProg      *obj.Prog
	dataAddr      map[string]int64 // Most recent address for DATA for this symbol.
	isJump        bool             // Instruction being assembled is a jump.
	allowABI      bool             // Whether ABI selectors are allowed.
	pkgPrefix     string           // Prefix to add to local symbols.
	errorWriter   io.Writer
}

func NewParser(ctxt *obj.Link, ar *arch.Arch, lexer lex.TokenReader) *Parser {
	pkgPrefix := obj.UnlinkablePkg
	if ctxt != nil {
		pkgPrefix = objabi.PathToPrefix(ctxt.Pkgpath)
	}
	return &Parser{
		ctxt:        ctxt,
		arch:        ar,
		lex:         lexer,
		labels:      make(map[string]*obj.Prog),
		dataAddr:    make(map[string]int64),
		errorWriter: os.Stderr,
		allowABI:    ctxt != nil && objabi.LookupPkgSpecial(ctxt.Pkgpath).AllowAsmABI,
		pkgPrefix:   pkgPrefix,
	}
}
