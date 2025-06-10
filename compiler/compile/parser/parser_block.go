package parser

// ----------------------------------------------------------------------------
// Blocks

// 与函数 parseBlockStmt 完全等价！
//func (p *parser) parseBody() *ast.BlockStmt {
//	lbrace := p.expect(LBRACE) // {
//	list := p.parseStmtList()
//	rbrace := p.expect(RBRACE) // }
//
//	return &ast.BlockStmt{Lbrace: lbrace, List: list, Rbrace: rbrace}
//}
