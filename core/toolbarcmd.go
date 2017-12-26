package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmigpin/editor/core/cmdutil"
	"github.com/jmigpin/editor/core/toolbardata"
	"github.com/jmigpin/editor/ui"
	"github.com/jmigpin/editor/ui/tautil"
)

func ToolbarCmdFromLayout(ed *Editor, ta *ui.TextArea) {
	td := toolbardata.NewToolbarData(ta.Str(), ed.HomeVars())
	part, ok := td.GetPartAtIndex(ta.CursorIndex())
	if !ok {
		ed.Errorf("missing part at index")
		return
	}
	runCommand(ed, part, nil)
}

func ToolbarCmdFromRow(ed *Editor, erow *ERow) {
	td := erow.ToolbarData()
	ta := erow.Row().Toolbar
	part, ok := td.GetPartAtIndex(ta.CursorIndex())
	if !ok {
		ed.Errorf("missing part at index")
		return
	}

	if part == td.Parts[0] {
		if !firstPartCmd(ed, part, erow) {
			ed.Errorf("no cmd was run")
		}
		return
	}

	runCommand(ed, part, erow)
}

func firstPartCmd(ed *Editor, part *toolbardata.Part, erow cmdutil.ERower) bool {
	if len(part.Args) == 0 {
		return false
	}

	a0 := part.Args[0]
	tb := erow.Row().Toolbar
	ci := tb.CursorIndex()

	// cursor index beyond arg0
	if ci > a0.E {
		return false
	}

	// get path up to cursor index
	str := a0.Str
	i := strings.Index(str[ci:], string(filepath.Separator))
	if i >= 0 {
		str = str[:ci+i]
	}

	// decode str
	td := toolbardata.NewToolbarData(str, erow.Ed().HomeVars())
	str = td.DecodePart0Arg0()

	// file info
	_, err := os.Stat(str)
	if err == nil {
		// open row next to erow and load content
		col := erow.Row().Col
		u, ok := erow.Row().NextRow()
		if !ok {
			u = nil
		}
		erow2 := ed.NewERowerBeforeRow(str, col, u)
		err := erow2.LoadContentClear()
		if err != nil {
			ed.Error(err)
		}
	}

	return true
}

func runCommand(ed *Editor, part *toolbardata.Part, erow cmdutil.ERower) {
	if len(part.Args) < 1 {
		return
	}

	p0 := part.Args[0].Str

	layoutOnly := func(fn func()) {
		if erow != nil {
			ed.Errorf("%s: layout only command", p0)
			return
		}
		fn()
	}

	erowCmd := func(fn func(cmdutil.ERower)) {
		e := erow
		if e == nil {
			aerow, ok := ed.ActiveERower()
			if !ok {
				ed.Errorf("%s: no active row", p0)
				return
			}
			e = aerow
		}
		fn(e)
	}

	switch p0 {
	case "Exit":
		layoutOnly(func() { ed.Close() })
	case "SaveSession":
		layoutOnly(func() { cmdutil.SaveSession(ed, part) })
	case "OpenSession":
		layoutOnly(func() { cmdutil.OpenSession(ed, part) })
	case "DeleteSession":
		layoutOnly(func() { cmdutil.DeleteSession(ed, part) })
	case "ListSessions":
		layoutOnly(func() { cmdutil.ListSessions(ed) })
	case "NewColumn":
		layoutOnly(func() { _ = ed.ui.Layout.Cols.NewColumn() })
	case "SaveAllFiles":
		layoutOnly(func() { cmdutil.SaveRowsFiles(ed) })
	case "ReloadAll":
		layoutOnly(func() { cmdutil.ReloadRows(ed) })
	case "ReloadAllFiles":
		layoutOnly(func() { cmdutil.ReloadRowsFiles(ed) })
	case "NewRow":
		layoutOnly(func() { cmdutil.NewRow(ed) })
	case "ReopenRow":
		layoutOnly(func() { ed.reopenRow.Reopen() })

	case "ColorTheme":
		_colorThemeCmd(ed)
	case "FontTheme":
		_fontThemeCmd(ed)
	case "FontRunes":
		_fontRunesCmd(ed)
	case "FWStatus":
		_fwStatusCmd(ed)

	case "XdgOpenDir":
		erowCmd(func(e cmdutil.ERower) { cmdutil.XdgOpenDirShortcut(ed, e) })
	case "RowDirectory":
		erowCmd(func(e cmdutil.ERower) { cmdutil.OpenRowDirectory(ed, e) })
	case "DuplicateRow":
		erowCmd(func(e cmdutil.ERower) { cmdutil.DuplicateRow(ed, e) })
	case "MaximizeRow":
		erowCmd(func(e cmdutil.ERower) { cmdutil.MaximizeRow(ed, e) })
	case "Save":
		erowCmd(func(e cmdutil.ERower) { cmdutil.SaveRowFile(e) })
	case "Reload":
		erowCmd(func(e cmdutil.ERower) { cmdutil.ReloadRow(e) })
	case "CloseRow":
		erowCmd(func(e cmdutil.ERower) { e.Row().Close() })
	case "CloseColumn":
		erowCmd(func(e cmdutil.ERower) { e.Row().Col.Close() })
	case "Find":
		erowCmd(func(e cmdutil.ERower) { _findCmd(e, part) })
	case "GotoLine":
		erowCmd(func(e cmdutil.ERower) { cmdutil.GotoLine(e, part) })
	case "Replace":
		erowCmd(func(e cmdutil.ERower) { cmdutil.Replace(e, part) })
	case "Stop":
		erowCmd(func(e cmdutil.ERower) { cmdutil.RowCtxCancel(e.Row()) })
	case "ListDir":
		erowCmd(func(e cmdutil.ERower) { cmdutil.ListDirEd(e, false, false) })
	case "ListDirSub":
		erowCmd(func(e cmdutil.ERower) { cmdutil.ListDirEd(e, true, false) })
	case "ListDirHidden":
		erowCmd(func(e cmdutil.ERower) { cmdutil.ListDirEd(e, false, true) })
	case "CopyFilePosition":
		erowCmd(func(e cmdutil.ERower) { cmdutil.CopyFilePosition(ed, e) })

	default:
		erowCmd(func(e cmdutil.ERower) { cmdutil.ExternalCmd(e, part) })
	}
}

func _findCmd(erow cmdutil.ERower, part *toolbardata.Part) {
	a := part.Args[1:]
	if len(a) != 1 {
		erow.Ed().Error(fmt.Errorf("find: expecting 1 argument"))
		return
	}
	tautil.Find(erow.Row().TextArea, a[0].Str)
}
func _colorThemeCmd(ed *Editor) {
	ui.CycleColorTheme()
	ed.ui.Layout.MarkNeedsPaint()
}
func _fontThemeCmd(ed *Editor) {
	ui.CycleFontTheme()
	ed.ui.Layout.CalcChildsBounds()
	ed.ui.Layout.MarkNeedsPaint()
}
func _fontRunesCmd(ed *Editor) {
	var u string
	for i := 0; i < 15000; {
		start := i
		var w string
		for j := 0; j < 25; j++ {
			w += string(rune(i))
			i++
		}
		u += fmt.Sprintf("%d: %s\n", start, w)
	}
	ed.Messagef("%s", u)
}
func _fwStatusCmd(ed *Editor) {
	ed.Messagef("%s", ed.fwatcher.Status())
}
