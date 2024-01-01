package ddlctlgo

import (
	"bytes"
	goast "go/ast"
	"go/token"
	"io"

	"github.com/kunitsucom/ddlctl/internal/logs"
)

func dumpDDLSource(fset *token.FileSet, ddlSrc []*ddlSource) {
	for _, r := range ddlSrc {
		logs.Trace.Print("== ddlSource ================================================================================================================================")
		_, _ = io.WriteString(logs.Trace.LineWriter("r.CommentGroup.Text: "), r.CommentGroup.Text())
		logs.Trace.Print("-- CommentGroup --------------------------------------------------------------------------------------------------------------------------------")
		{
			commentGroupAST := bytes.NewBuffer(nil)
			goast.Fprint(commentGroupAST, fset, r.CommentGroup, goast.NotNilFilter)
			_, _ = logs.Trace.LineWriter("").Write(commentGroupAST.Bytes())
		}
		logs.Trace.Print("-- TypeSpec --------------------------------------------------------------------------------------------------------------------------------")
		{
			typeSpecAST := bytes.NewBuffer(nil)
			goast.Fprint(typeSpecAST, fset, r.TypeSpec, goast.NotNilFilter)
			_, _ = logs.Trace.LineWriter("").Write(typeSpecAST.Bytes())
		}
		logs.Trace.Print("-- StructType --------------------------------------------------------------------------------------------------------------------------------")
		{
			structTypeAST := bytes.NewBuffer(nil)
			goast.Fprint(structTypeAST, fset, r.StructType, goast.NotNilFilter)
			_, _ = logs.Trace.LineWriter("").Write(structTypeAST.Bytes())
		}
	}
}
