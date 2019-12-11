package godebug

// DO NOT EDIT: code generated by debugpack/debugpack.go
	
type FilePack struct {
	Name string
	Data string
}

func DebugFilePacks() []*FilePack {
	return []*FilePack{{"debug.go", "package debug\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\t\"sync\"\n)\n\nvar server *Server\nvar startServerMu sync.Mutex\n\n// Called by the generated config.\nfunc StartServer() {\n\thotStartServer()\n}\n\nfunc hotStartServer() {\n\tif server == nil {\n\t\tstartServerMu.Lock()\n\t\tif server == nil {\n\t\t\tstartServer()\n\t\t}\n\t\tstartServerMu.Unlock()\n\t}\n}\n\nfunc startServer() {\n\tsrv, err := NewServer()\n\tif err != nil {\n\t\tfmt.Printf(\"godebug/debug: start server error: %v\\n\", err)\n\t\tos.Exit(1)\n\t}\n\tserver = srv\n}\n\n//----------\n\n// Auto-inserted at main for a clean exit. Not to be used.\nfunc ExitServer() {\n\tif server != nil {\n\t\tserver.Close()\n\t}\n}\n\n//----------\n\n// Auto-inserted at annotations. Not to be used.\nfunc Line(fileIndex, debugIndex, offset int, item Item) {\n\thotStartServer()\n\tlmsg := &LineMsg{FileIndex: fileIndex, DebugIndex: debugIndex, Offset: offset, Item: item}\n\tserver.Send(lmsg)\n}\n\n//----------\n\n// DEPRECATED: use the \"//godebug:annotate*\" comments\n\n// no-op operation used for source detection by the annotator\n//func NoAnnotations()   {}\n//func AnnotateBlock()   {}\n//func AnnotateFile()    {}\n//func AnnotatePackage() {}\n"},
		{"encode.go", "package debug\n\nimport (\n\t\"bytes\"\n\t\"encoding/binary\"\n\t\"encoding/gob\"\n\t\"fmt\"\n\t\"io\"\n\t\"log\"\n)\n\nfunc RegisterStructure(v interface{}) {\n\tgob.Register(v)\n}\n\n//----------\n\nfunc EncodeMessage(msg interface{}) ([]byte, error) {\n\t// message buffer\n\tvar bbuf bytes.Buffer\n\n\t// reserve space to encode v size\n\tsizeBuf := make([]byte, 4)\n\tif _, err := bbuf.Write(sizeBuf[:]); err != nil {\n\t\treturn nil, err\n\t}\n\n\t// encode v\n\tenc := gob.NewEncoder(&bbuf)\n\tif err := enc.Encode(&msg); err != nil { // decoder uses &interface{}\n\t\treturn nil, err\n\t}\n\n\t// get bytes\n\tbuf := bbuf.Bytes()\n\n\t// encode v size at buffer start\n\tl := uint32(len(buf) - len(sizeBuf))\n\tbinary.BigEndian.PutUint32(buf, l)\n\n\treturn buf, nil\n}\n\nfunc DecodeMessage(reader io.Reader) (interface{}, error) {\n\treadN := func(b []byte) error {\n\t\tm := len(b)\n\t\ti := 0\n\t\tfor i < m {\n\t\t\tn, err := reader.Read(b[i:])\n\t\t\tif err != nil && n == 0 {\n\t\t\t\treturn err\n\t\t\t}\n\t\t\ti += n\n\t\t}\n\t\tif i != m {\n\t\t\terr := fmt.Errorf(\"expected to read %v but got %v\", m, i)\n\t\t\tlog.Printf(\"error: %v\", err)\n\t\t}\n\t\treturn nil\n\t}\n\n\t// read size\n\tsizeBuf := make([]byte, 4)\n\tif err := readN(sizeBuf); err != nil {\n\t\treturn nil, err\n\t}\n\tl := int(binary.BigEndian.Uint32(sizeBuf))\n\n\t// read msg\n\tmsgBuf := make([]byte, l)\n\tif err := readN(msgBuf); err != nil {\n\t\treturn nil, err\n\t}\n\n\t// decode msg\n\tbuf := bytes.NewBuffer(msgBuf)\n\tdec := gob.NewDecoder(buf)\n\tvar msg interface{}\n\tif err := dec.Decode(&msg); err != nil {\n\t\treturn nil, err\n\t}\n\n\treturn msg, nil\n}\n\n//----------\n\n// TODO: document why this simplified version doesn't work (hangs)\n\n//func EncodeMessage(msg interface{}) ([]byte, error) {\n//\tvar buf bytes.Buffer\n//\tenc := gob.NewEncoder(&buf)\n//\tif err := enc.Encode(&msg); err != nil {\n//\t\treturn nil, err\n//\t}\n//\treturn buf.Bytes(), nil\n//}\n\n//func DecodeMessage(reader io.Reader) (interface{}, error) {\n//\tdec := gob.NewDecoder(reader)\n//\tvar msg interface{}\n//\tif err := dec.Decode(&msg); err != nil {\n//\t\treturn nil, err\n//\t}\n//\treturn msg, nil\n//}\n\n//----------\n"},
		{"limitedwriter.go", "package debug\n\nimport (\n\t\"bytes\"\n\t\"fmt\"\n)\n\ntype LimitedWriter struct {\n\tsize int\n\tbuf  bytes.Buffer\n}\n\nfunc NewLimitedWriter(size int) *LimitedWriter {\n\treturn &LimitedWriter{size: size}\n}\n\nfunc (w *LimitedWriter) Write(p []byte) (n int, err error) {\n\tif w.size < len(p) {\n\t\tp = p[:w.size]\n\t\terr = LimitReachedErr\n\t}\n\tn, err2 := w.buf.Write(p)\n\tif err2 != nil {\n\t\treturn n, err2\n\t}\n\tw.size -= n\n\treturn n, err\n}\n\nfunc (w *LimitedWriter) Bytes() []byte {\n\treturn w.buf.Bytes()\n}\n\nvar LimitReachedErr = fmt.Errorf(\"limit reached\")\n"},
		{"server.go", "package debug\n\nimport (\n\t\"io\"\n\t\"io/ioutil\"\n\t\"log\"\n\t\"net\"\n\t\"sync\"\n\t\"time\"\n)\n\n// Vars populated at init by godebugconfig pkg (generated at compile).\nvar AnnotatorFilesData []*AnnotatorFileData // all debug data\nvar ServerNetwork string\nvar ServerAddress string\n\n//----------\n\n//var logger = log.New(os.Stdout, \"debug: \", 0)\nvar logger = log.New(ioutil.Discard, \"debug: \", 0)\n\nconst chunkSendRate = 15  // per second\nconst sendNowNMsgs = 2048 // don't wait for send rate, send now (memory)\nconst sendQSize = 512     // msgs queueing to be sent\n\n//----------\n\ntype Server struct {\n\tln     net.Listener\n\tlnwait sync.WaitGroup\n\tclient struct {\n\t\tsync.RWMutex\n\t\tcconn *CConn\n\t}\n\tsendReady sync.RWMutex\n}\n\nfunc NewServer() (*Server, error) {\n\t// start listening\n\tlogger.Print(\"listen\")\n\tln, err := net.Listen(ServerNetwork, ServerAddress)\n\tif err != nil {\n\t\treturn nil, err\n\t}\n\n\tsrv := &Server{ln: ln}\n\tsrv.sendReady.Lock() // not ready to send (no client yet)\n\n\t// accept connections\n\tsrv.lnwait.Add(1)\n\tgo func() {\n\t\tdefer srv.lnwait.Done()\n\t\tsrv.acceptClientsLoop()\n\t}()\n\n\treturn srv, nil\n}\n\n//----------\n\nfunc (srv *Server) Close() {\n\t// close listener\n\tlogger.Println(\"closing server\")\n\t_ = srv.ln.Close()\n\tsrv.lnwait.Wait()\n\n\t// close client\n\tlogger.Println(\"closing client\")\n\tsrv.client.Lock()\n\tif srv.client.cconn != nil {\n\t\tsrv.client.cconn.Close()\n\t\tsrv.client.cconn = nil\n\t}\n\tsrv.client.Unlock()\n\n\tlogger.Println(\"server closed\")\n}\n\n//----------\n\nfunc (srv *Server) acceptClientsLoop() {\n\tfor {\n\t\t// accept client\n\t\tlogger.Println(\"waiting for client\")\n\t\tconn, err := srv.ln.Accept()\n\t\tif err != nil {\n\t\t\tlogger.Printf(\"accept error: (%T) %v \", err, err)\n\n\t\t\t// unable to accept (ex: server was closed)\n\t\t\tif operr, ok := err.(*net.OpError); ok {\n\t\t\t\tif operr.Op == \"accept\" {\n\t\t\t\t\tlogger.Println(\"end accept client loop\")\n\t\t\t\t\treturn\n\t\t\t\t}\n\t\t\t}\n\n\t\t\tcontinue\n\t\t}\n\t\tlogger.Println(\"got client\")\n\n\t\t// start client\n\t\tsrv.client.Lock()\n\t\tif srv.client.cconn != nil {\n\t\t\tsrv.client.cconn.Close() // close previous connection\n\t\t}\n\t\tsrv.client.cconn = NewCCon(srv, conn)\n\t\tsrv.client.Unlock()\n\t}\n}\n\n//----------\n\nfunc (srv *Server) Send(v *LineMsg) {\n\t// locks if client is not ready to send\n\tsrv.sendReady.RLock()\n\tdefer srv.sendReady.RUnlock()\n\n\tsrv.client.cconn.Send(v)\n}\n\n//----------\n\n// Client connection.\ntype CConn struct {\n\tsrv          *Server\n\tconn         net.Conn\n\trwait, swait sync.WaitGroup\n\tsendch       chan *LineMsg // sending loop channel\n\treqStart     struct {\n\t\tsync.Mutex\n\t\tstart   chan struct{}\n\t\tstarted bool\n\t\tclosed  bool\n\t}\n}\n\nfunc NewCCon(srv *Server, conn net.Conn) *CConn {\n\tcconn := &CConn{srv: srv, conn: conn}\n\tcconn.sendch = make(chan *LineMsg, sendQSize)\n\tcconn.reqStart.start = make(chan struct{})\n\n\t// receive messages\n\tcconn.rwait.Add(1)\n\tgo func() {\n\t\tdefer cconn.rwait.Done()\n\t\tcconn.receiveMsgsLoop()\n\t}()\n\n\t// send msgs\n\tcconn.swait.Add(1)\n\tgo func() {\n\t\tdefer cconn.swait.Done()\n\t\tcconn.sendMsgsLoop()\n\t}()\n\n\treturn cconn\n}\n\nfunc (cconn *CConn) Close() {\n\tcconn.reqStart.Lock()\n\tif cconn.reqStart.started {\n\t\t// not sendready anymore\n\t\tcconn.srv.sendReady.Lock()\n\t}\n\tcconn.reqStart.closed = true\n\tcconn.reqStart.Unlock()\n\n\t// close send msgs: can't close receive msgs first (closes client)\n\tclose(cconn.reqStart.start) // ok even if it didn't start\n\tclose(cconn.sendch)\n\tcconn.swait.Wait()\n\n\t// close receive msgs\n\t_ = cconn.conn.Close()\n\tcconn.rwait.Wait()\n}\n\n//----------\n\nfunc (cconn *CConn) receiveMsgsLoop() {\n\tfor {\n\t\tmsg, err := DecodeMessage(cconn.conn)\n\t\tif err != nil {\n\t\t\t// unable to read (server was probably closed)\n\t\t\tif operr, ok := err.(*net.OpError); ok {\n\t\t\t\tif operr.Op == \"read\" {\n\t\t\t\t\tbreak\n\t\t\t\t}\n\t\t\t}\n\t\t\t// connection ended gracefully by the client\n\t\t\tif err == io.EOF {\n\t\t\t\tbreak\n\t\t\t}\n\n\t\t\t// always print if the error reaches here\n\t\t\tlog.Print(err)\n\t\t\treturn\n\t\t}\n\n\t\t// handle msg\n\t\tswitch t := msg.(type) {\n\t\tcase *ReqFilesDataMsg:\n\t\t\tlogger.Print(\"sending files data\")\n\t\t\tmsg := &FilesDataMsg{Data: AnnotatorFilesData}\n\t\t\tif err := cconn.send2(msg); err != nil {\n\t\t\t\tlog.Println(err)\n\t\t\t}\n\t\tcase *ReqStartMsg:\n\t\t\tlogger.Print(\"reqstart\")\n\t\t\tcconn.reqStart.Lock()\n\t\t\tif !cconn.reqStart.started && !cconn.reqStart.closed {\n\t\t\t\tcconn.reqStart.start <- struct{}{}\n\t\t\t\tcconn.reqStart.started = true\n\t\t\t\tcconn.srv.sendReady.Unlock()\n\t\t\t}\n\t\t\tcconn.reqStart.Unlock()\n\t\tdefault:\n\t\t\t// always print if there is a new msg type\n\t\t\tlog.Printf(\"todo: unexpected msg type: %T\", t)\n\t\t}\n\t}\n}\n\n//----------\n\nfunc (cconn *CConn) sendMsgsLoop() {\n\t// wait for reqstart, or the client won't have the index data\n\t_, ok := <-cconn.reqStart.start\n\tif !ok {\n\t\treturn\n\t}\n\n\t//// commented: simple iterative send (slow)\n\t//for {\n\t//\tv, ok := <-conn.sendch\n\t//\tif !ok {\n\t//\t\tbreak\n\t//\t}\n\t//\tif err := conn.send2(v); err != nil {\n\t//\t\tlog.Println(err)\n\t//\t}\n\t//}\n\t//return\n\n\t// send in chunks (better performance)\n\tscheduled := false\n\ttimeToSend := make(chan bool)\n\tmsgs := []*LineMsg{}\n\tsendMsgs := func() {\n\t\tif len(msgs) > 0 {\n\t\t\tif err := cconn.send2(msgs); err != nil {\n\t\t\t\tlog.Println(err)\n\t\t\t}\n\t\t\tmsgs = nil\n\t\t}\n\t}\n\tfor {\n\t\tselect {\n\t\tcase v, ok := <-cconn.sendch:\n\t\t\tif !ok {\n\t\t\t\tgoto loopEnd\n\t\t\t}\n\t\t\tmsgs = append(msgs, v)\n\t\t\tif len(msgs) >= sendNowNMsgs {\n\t\t\t\tsendMsgs()\n\t\t\t} else if !scheduled {\n\t\t\t\tscheduled = true\n\t\t\t\tgo func() {\n\t\t\t\t\td := time.Second / time.Duration(chunkSendRate)\n\t\t\t\t\ttime.Sleep(d)\n\t\t\t\t\ttimeToSend <- true\n\t\t\t\t}()\n\t\t\t}\n\t\tcase <-timeToSend:\n\t\t\tscheduled = false\n\t\t\tsendMsgs()\n\t\t}\n\t}\nloopEnd:\n\t// send last messages if any\n\tsendMsgs()\n}\n\nfunc (cconn *CConn) send2(v interface{}) error {\n\tencoded, err := EncodeMessage(v)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\tn, err := cconn.conn.Write(encoded)\n\tif err != nil {\n\t\treturn err\n\t}\n\tif n != len(encoded) {\n\t\tlogger.Printf(\"n!=len(encoded): %v %v\\n\", n, len(encoded))\n\t}\n\treturn nil\n}\n\n//----------\n\nfunc (cconn *CConn) Send(v *LineMsg) {\n\tcconn.sendch <- v\n}\n"},
		{"stringifyv.go", "package debug\n\nimport (\n\t\"fmt\"\n\t\"strconv\"\n)\n\nfunc stringifyV(v V) string {\n\t// Note: rune is an alias for int32, can't \"case rune:\"\n\tconst max = 150\n\tstr := \"\"\n\tswitch t := v.(type) {\n\tcase nil:\n\t\treturn \"nil\"\n\tcase string:\n\t\tstr = ReducedSprintf(max, \"%q\", t)\n\tcase []string:\n\t\tstr = quotedStrings(max, t)\n\tcase fmt.Stringer:\n\t\tstr = ReducedSprintf(max, \"%q\", t)\n\tcase error:\n\t\tstr = ReducedSprintf(max, \"%q\", t)\n\tcase float32:\n\t\tstr = strconv.FormatFloat(float64(t), 'f', -1, 32)\n\tcase float64:\n\t\tstr = strconv.FormatFloat(t, 'f', -1, 64)\n\tdefault:\n\t\tstr = ReducedSprintf(max, \"%v\", v) // ex: bool\n\t}\n\treturn str\n}\n\nfunc ReducedSprintf(max int, format string, a ...interface{}) string {\n\tw := NewLimitedWriter(max)\n\t_, err := fmt.Fprintf(w, format, a...)\n\ts := string(w.Bytes())\n\tif err == LimitReachedErr {\n\t\ts += \"...\"\n\t\t// close quote if present\n\t\tconst q = '\"'\n\t\tif rune(s[0]) == q {\n\t\t\ts += string(q)\n\t\t}\n\t}\n\treturn s\n}\n\nfunc quotedStrings(max int, a []string) string {\n\tw := NewLimitedWriter(max)\n\tsp := \"\"\n\tlimited := 0\n\tfor i, s := range a {\n\t\tif i > 0 {\n\t\t\tsp = \" \"\n\t\t}\n\t\tn, err := fmt.Fprintf(w, \"%s%q\", sp, s)\n\t\tif err != nil {\n\t\t\tif err == LimitReachedErr {\n\t\t\t\tlimited = n\n\t\t\t}\n\t\t\tbreak\n\t\t}\n\t}\n\ts := string(w.Bytes())\n\tif limited > 0 {\n\t\ts += \"...\"\n\t\tif limited >= 2 { // 1=space, 2=quote\n\t\t\ts += `\"` // close quote\n\t\t}\n\t}\n\treturn \"[\" + s + \"]\"\n}\n"},
		{"structs.go", "package debug\n\nimport (\n\t\"fmt\"\n)\n\nfunc init() {\n\t// register structs to be able to encode/decode from interface{}\n\n\treg := RegisterStructure\n\n\treg(&ReqFilesDataMsg{})\n\treg(&FilesDataMsg{})\n\treg(&ReqStartMsg{})\n\treg(&LineMsg{})\n\treg([]*LineMsg{})\n\n\treg(&ItemValue{})\n\treg(&ItemList{})\n\treg(&ItemList2{})\n\treg(&ItemAssign{})\n\treg(&ItemSend{})\n\treg(&ItemCall{})\n\treg(&ItemCallEnter{})\n\treg(&ItemIndex{})\n\treg(&ItemIndex2{})\n\treg(&ItemKeyValue{})\n\treg(&ItemSelector{})\n\treg(&ItemTypeAssert{})\n\treg(&ItemBinary{})\n\treg(&ItemUnary{})\n\treg(&ItemUnaryEnter{})\n\treg(&ItemParen{})\n\treg(&ItemLiteral{})\n\treg(&ItemBranch{})\n\treg(&ItemStep{})\n\treg(&ItemAnon{})\n}\n\n//----------\n\ntype ReqFilesDataMsg struct{}\ntype ReqStartMsg struct{}\n\n//----------\n\ntype LineMsg struct {\n\tFileIndex  int\n\tDebugIndex int\n\tOffset     int\n\tItem       Item\n}\n\ntype FilesDataMsg struct {\n\tData []*AnnotatorFileData\n}\n\ntype AnnotatorFileData struct {\n\tFileIndex int\n\tDebugLen  int\n\tFilename  string\n\tFileSize  int\n\tFileHash  []byte\n}\n\n//----------\n\ntype Item interface {\n}\ntype ItemValue struct {\n\tStr string\n}\ntype ItemList struct { // separated by \",\"\n\tList []Item\n}\ntype ItemList2 struct { // separated by \";\"\n\tList []Item\n}\ntype ItemAssign struct {\n\tLhs, Rhs *ItemList\n}\ntype ItemSend struct {\n\tChan, Value Item\n}\ntype ItemCall struct {\n\tName   string\n\tArgs   *ItemList\n\tResult Item\n}\ntype ItemCallEnter struct {\n\tName string\n\tArgs *ItemList\n}\ntype ItemIndex struct {\n\tResult Item\n\tExpr   Item\n\tIndex  Item\n}\ntype ItemIndex2 struct {\n\tResult         Item\n\tExpr           Item\n\tLow, High, Max Item\n\tSlice3         bool // 2 colons present\n}\ntype ItemKeyValue struct {\n\tKey   Item\n\tValue Item\n}\ntype ItemSelector struct {\n\tX   Item\n\tSel Item\n}\ntype ItemTypeAssert struct {\n\tX    Item\n\tType Item\n}\ntype ItemBinary struct {\n\tResult Item\n\tOp     int\n\tX, Y   Item\n}\ntype ItemUnary struct {\n\tResult Item\n\tOp     int\n\tX      Item\n}\ntype ItemUnaryEnter struct {\n\tOp int\n\tX  Item\n}\ntype ItemParen struct {\n\tX Item\n}\ntype ItemLiteral struct {\n\tFields *ItemList\n}\ntype ItemBranch struct{}\ntype ItemStep struct{}\ntype ItemAnon struct{}\n\n//----------\n\ntype V interface{}\n\n// ItemValue\nfunc IV(v V) Item {\n\treturn &ItemValue{Str: stringifyV(v)}\n}\n\n// ItemValue: raw string\nfunc IVs(s string) Item {\n\treturn &ItemValue{Str: s}\n}\n\n// ItemValue: typeof\nfunc IVt(v V) Item {\n\treturn &ItemValue{Str: fmt.Sprintf(\"%T\", v)}\n}\n\n// ItemValue: len\nfunc IVl(v V) Item {\n\treturn &ItemValue{Str: fmt.Sprintf(\"%v=len()\", v)}\n}\n\n// ItemList (\",\" and \";\")\nfunc IL(u ...Item) *ItemList {\n\treturn &ItemList{List: u}\n}\nfunc IL2(u ...Item) Item {\n\treturn &ItemList2{List: u}\n}\n\n// ItemAssign\nfunc IA(lhs, rhs *ItemList) Item {\n\treturn &ItemAssign{Lhs: lhs, Rhs: rhs}\n}\n\n// ItemSend\nfunc IS(ch, value Item) Item {\n\treturn &ItemSend{Chan: ch, Value: value}\n}\n\n// ItemCall\nfunc IC(name string, result Item, args ...Item) Item {\n\treturn &ItemCall{Name: name, Result: result, Args: IL(args...)}\n}\n\n// ItemCall: enter\nfunc ICe(name string, args ...Item) Item {\n\treturn &ItemCallEnter{Name: name, Args: IL(args...)}\n}\n\n// ItemIndex\nfunc II(result, expr, index Item) Item {\n\treturn &ItemIndex{Result: result, Expr: expr, Index: index}\n}\nfunc II2(result, expr, low, high, max Item, slice3 bool) Item {\n\treturn &ItemIndex2{Result: result, Expr: expr, Low: low, High: high, Max: max, Slice3: slice3}\n}\n\n// ItemKeyValue\nfunc IKV(key, value Item) Item {\n\treturn &ItemKeyValue{Key: key, Value: value}\n}\n\n// ItemSelector\nfunc ISel(x, sel Item) Item {\n\treturn &ItemSelector{X: x, Sel: sel}\n}\n\n// ItemTypeAssert\nfunc ITA(x, t Item) Item {\n\treturn &ItemTypeAssert{X: x, Type: t}\n}\n\n// ItemBinary\nfunc IB(result Item, op int, x, y Item) Item {\n\treturn &ItemBinary{Result: result, Op: op, X: x, Y: y}\n}\n\n// ItemUnary\nfunc IU(result Item, op int, x Item) Item {\n\treturn &ItemUnary{Result: result, Op: op, X: x}\n}\n\n// ItemUnary: enter\nfunc IUe(op int, x Item) Item {\n\treturn &ItemUnaryEnter{Op: op, X: x}\n}\n\n// ItemParen\nfunc IP(x Item) Item {\n\treturn &ItemParen{X: x}\n}\n\n// ItemLiteral\nfunc ILit(fields ...Item) Item {\n\treturn &ItemLiteral{Fields: IL(fields...)}\n}\n\n// ItemBranch\nfunc IBr() Item {\n\treturn &ItemBranch{}\n}\n\n// ItemStep\nfunc ISt() Item {\n\treturn &ItemStep{}\n}\n\n// ItemAnon\nfunc IAn() Item {\n\treturn &ItemAnon{}\n}\n"}}
}
