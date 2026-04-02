package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestMatrixBuildsASVSCoverageRows(t *testing.T) {
	rows := Matrix(MatrixInfos(rule.Builtins(model.Config{})))
	if len(rows) == 0 {
		t.Fatal("expected ASVS matrix rows")
	}
	found := false
	for _, row := range rows {
		if row.Control == "OWASP-ASVS-V4.1" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected authorization control row")
	}
}
