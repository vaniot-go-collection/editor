/*
Build with:
$ go build -buildmode=plugin gotodefinition_godef.go
*/

package main

import (
	"bytes"
	"context"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jmigpin/editor/core"
	"github.com/jmigpin/editor/core/parseutil"
)

func OnLoad(ed *core.Editor) {
	// default contentcmds at: github.com/jmigpin/editor/core/contentcmds/init.go
	core.ContentCmds.Remove("gotodefinition") // remove default
	core.ContentCmds.Prepend("gotodefinition_godef", goToDefinition)
}

func goToDefinition(erow *core.ERow, index int) (handled bool, err error) {
	if erow.Info.IsDir() {
		return false, nil
	}
	if path.Ext(erow.Info.Name()) != ".go" {
		return false, nil
	}

	// timeout for the cmd to run
	timeout := 8000 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// it's a go file, return true from here

	// godef args
	args := []string{"godef", "-i", "-f", erow.Info.Name(), "-o", strconv.Itoa(index)}

	// godef can read from stdin: use textarea bytes
	bin, err := erow.Row.TextArea.Bytes()
	if err != nil {
		return true, err
	}
	in := bytes.NewBuffer(bin)

	// execute external cmd
	dir := filepath.Dir(erow.Info.Name())
	out, err := core.ExecCmdStdin(ctx, dir, in, args...)
	if err != nil {
		return true, err
	}

	// parse external cmd output
	filePos, err := parseutil.ParseFilePos(string(out))
	if err != nil {
		return true, err
	}

	// place under the calling row
	rowPos := erow.Row.PosBelow()

	conf := &core.OpenFileERowConfig{
		FilePos:               filePos,
		RowPos:                rowPos,
		FlashVisibleOffsets:   true,
		NewIfNotExistent:      true,
		NewIfOffsetNotVisible: true,
	}
	core.OpenFileERow(erow.Ed, conf)

	return true, nil
}
