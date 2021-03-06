package sqlite3_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mutablelogic/go-sqlite/sys/sqlite3"
)

func Test_Func_001(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)
	db, err := sqlite3.OpenPathEx(filepath.Join(tmpdir, "test.sqlite"), sqlite3.SQLITE_OPEN_CREATE, "")
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	db.SetBusyTimeout(time.Second)

	// Create a function which sleeps
	if err := db.CreateScalarFunction("sleepy", 0, true, func(ctx *sqlite3.Context, args []*sqlite3.Value) {
		sqlite3.Sleep(time.Second * 5)
	}); err != nil {
		t.Error(err)
	}

	// Execute sleepy function
	st, err := db.Prepare(fmt.Sprint("SELECT SLEEPY()"))
	if err != nil {
		t.Error(err)
	}
	defer st.Close()
	r, err := st.Exec(0)
	if err != nil {
		t.Error(err)
	}
	t.Log(r)
	for {
		row := r.Next()
		if row == nil {
			break
		}
		t.Log(row)
	}
}
