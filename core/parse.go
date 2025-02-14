package core

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unsafe"
)

type (
	Expr interface {
		Eval(env *LocalEnv) Object
		InferType() *Type
		Pos() Position
		Dump(includePosition bool) Map
		Pack(p []byte, env *PackEnv) []byte
	}
	LiteralExpr struct {
		Position
		obj         Object
		isSurrogate bool
	}
	VectorExpr struct {
		Position
		v []Expr
	}
	MapExpr struct {
		Position
		keys   []Expr
		values []Expr
	}
	SetExpr struct {
		Position
		elements []Expr
	}
	IfExpr struct {
		Position
		cond     Expr
		positive Expr
		negative Expr
	}
	DefExpr struct {
		Position
		vr               *Var
		name             Symbol
		value            Expr
		meta             Expr
		isCreatedByMacro bool
	}
	CallExpr struct {
		Position
		callable Expr
		args     []Expr
	}
	MacroCallExpr struct {
		Position
		macro Callable
		args  []Object
		name  string
	}
	RecurExpr struct {
		Position
		args []Expr
	}
	VarRefExpr struct {
		Position
		vr *Var
	}
	BindingExpr struct {
		Position
		binding *Binding
	}
	MetaExpr struct {
		Position
		meta *MapExpr
		expr Expr
	}
	DoExpr struct {
		Position
		body             []Expr
		isCreatedByMacro bool
	}
	FnArityExpr struct {
		Position
		args       []Symbol
		body       []Expr
		taggedType *Type
	}
	FnExpr struct {
		Position
		arities  []FnArityExpr
		variadic *FnArityExpr
		self     Symbol
	}
	LetExpr struct {
		Position
		names  []Symbol
		values []Expr
		body   []Expr
	}
	LoopExpr  LetExpr
	ThrowExpr struct {
		Position
		e Expr
	}
	CatchExpr struct {
		Position
		excType   *Type
		excSymbol Symbol
		body      []Expr
	}
	TryExpr struct {
		Position
		body        []Expr
		catches     []*CatchExpr
		finallyExpr []Expr
	}
	SetMacroExpr struct {
		Position
		vr *Var
	}
	ParseError struct {
		obj Object
		msg string
	}
	Callable interface {
		Call(args []Object) Object
	}
	Binding struct {
		name         Symbol
		index        int
		frame        int
		isUsed       bool
		inferredType *Type
	}
	Bindings struct {
		bindings map[*string]*Binding
		parent   *Bindings
		frame    int
	}
	LocalEnv struct {
		bindings []Object
		parent   *LocalEnv
		frame    int
	}
	ParseContext struct {
		GlobalEnv              *Env
		localBindings          *Bindings
		loopBindings           [][]Symbol
		linterBindings         *Bindings
		recur                  bool
		noRecurAllowed         bool
		isUnknownCallableScope bool
	}
	Warnings struct {
		ifWithoutElse           bool
		unusedFnParameters      bool
		fnWithEmptyBody         bool
		ignoredUnusedNamespaces Set
		IgnoredFileRegexes      []*regexp.Regexp
		entryPoints             Set
	}
	Keywords struct {
		tag                Keyword
		skipUnused         Keyword
		private            Keyword
		line               Keyword
		column             Keyword
		file               Keyword
		ns                 Keyword
		macro              Keyword
		message            Keyword
		form               Keyword
		data               Keyword
		cause              Keyword
		arglist            Keyword
		doc                Keyword
		added              Keyword
		meta               Keyword
		knownMacros        Keyword
		rules              Keyword
		ifWithoutElse      Keyword
		unusedFnParameters Keyword
		fnWithEmptyBody    Keyword
		_prefix            Keyword
		pos                Keyword
		startLine          Keyword
		endLine            Keyword
		startColumn        Keyword
		endColumn          Keyword
		filename           Keyword
		object             Keyword
		type_              Keyword
		var_               Keyword
		value              Keyword
		vector             Keyword
		name               Keyword
		dynamic            Keyword
		require            Keyword
		_import            Keyword
		else_              Keyword
		none               Keyword
		validIdent         Keyword
		characterSet       Keyword
		encodingRange      Keyword
		core               Keyword
		symbol             Keyword
		visible            Keyword
		ascii              Keyword
		unicode            Keyword
		any                Keyword
	}
	Symbols struct {
		joker_core         Symbol
		underscore         Symbol
		catch              Symbol
		finally            Symbol
		amp                Symbol
		_if                Symbol
		quote              Symbol
		fn_                Symbol
		fn                 Symbol
		let_               Symbol
		let                Symbol
		letfn_             Symbol
		letfn              Symbol
		loop_              Symbol
		loop               Symbol
		recur              Symbol
		setMacro_          Symbol
		def                Symbol
		defLinter          Symbol
		_var               Symbol
		do                 Symbol
		throw              Symbol
		try                Symbol
		unquoteSplicing    Symbol
		list               Symbol
		concat             Symbol
		seq                Symbol
		apply              Symbol
		emptySymbol        Symbol
		unquote            Symbol
		vector             Symbol
		hashMap            Symbol
		hashSet            Symbol
		defaultDataReaders Symbol
		backslash          Symbol
		deref              Symbol
		ns                 Symbol
		defrecord          Symbol
		defprotocol        Symbol
		extendProtocol     Symbol
		extendType         Symbol
		deftype            Symbol
		proxy              Symbol
		reify              Symbol
	}
	Str struct {
		_if          *string
		quote        *string
		fn_          *string
		let_         *string
		letfn_       *string
		loop_        *string
		recur        *string
		setMacro_    *string
		def          *string
		defLinter    *string
		_var         *string
		do           *string
		throw        *string
		try          *string
		coreFilename *string
	}
)

var (
	LOCAL_BINDINGS *Bindings = nil
	KNOWN_MACROS   *Var
	REQUIRE_VAR    *Var
	ALIAS_VAR      *Var
	REFER_VAR      *Var
	CREATE_NS_VAR  *Var
	IN_NS_VAR      *Var
	WARNINGS       = Warnings{
		fnWithEmptyBody: true,
		entryPoints:     EmptySet(),
	}
)

func (b *Bindings) ToMap() Map {
	var res Map = EmptyArrayMap()
	for b != nil {
		for _, v := range b.bindings {
			res = res.Assoc(v.name, NIL).(Map)
		}
		b = b.parent
	}
	return res
}

func (localEnv *LocalEnv) addEmptyFrame(capacity int) *LocalEnv {
	res := LocalEnv{
		bindings: make([]Object, 0, capacity),
		parent:   localEnv,
	}
	if localEnv != nil {
		res.frame = localEnv.frame + 1
	}
	return &res
}

func (localEnv *LocalEnv) addBinding(obj Object) {
	localEnv.bindings = append(localEnv.bindings, obj)
}

func (localEnv *LocalEnv) addFrame(values []Object) *LocalEnv {
	res := LocalEnv{
		bindings: values,
		parent:   localEnv,
	}
	if localEnv != nil {
		res.frame = localEnv.frame + 1
	}
	return &res
}

func (localEnv *LocalEnv) replaceFrame(values []Object) *LocalEnv {
	res := LocalEnv{
		bindings: values,
		parent:   localEnv.parent,
		frame:    localEnv.frame,
	}
	return &res
}

func (ctx *ParseContext) PushLoopBindings(bindings []Symbol) {
	ctx.loopBindings = append(ctx.loopBindings, bindings)
}

func (ctx *ParseContext) PopLoopBindings() {
	ctx.loopBindings = ctx.loopBindings[:len(ctx.loopBindings)-1]
}

func (ctx *ParseContext) GetLoopBindings() []Symbol {
	n := len(ctx.loopBindings)
	if n == 0 {
		return nil
	}
	return ctx.loopBindings[n-1]
}

func (b *Bindings) PushFrame() *Bindings {
	frame := 0
	if b != nil {
		frame = b.frame + 1
	}
	return &Bindings{
		bindings: make(map[*string]*Binding),
		parent:   b,
		frame:    frame,
	}
}

func (b *Bindings) PopFrame() *Bindings {
	return b.parent
}

func (b *Bindings) AddBinding(sym Symbol, index int, skipUnused bool, inferredType *Type) {
	if LINTER_MODE && !skipUnused {
		old := b.bindings[sym.name]
		if old != nil && needsUnusedWarning(old) {
			printParseWarning(GetPosition(old.name), "Unused binding: "+old.name.ToString(false))
		}
	}
	b.bindings[sym.name] = &Binding{
		name:         sym,
		frame:        b.frame,
		index:        index,
		inferredType: inferredType,
	}
}

func (ctx *ParseContext) PushEmptyLocalFrame() {
	ctx.localBindings = ctx.localBindings.PushFrame()
}

func (ctx *ParseContext) PushLocalFrame(names []Symbol) {
	ctx.PushEmptyLocalFrame()
	for i, sym := range names {
		ctx.localBindings.AddBinding(sym, i, true, nil)
	}
}

func (ctx *ParseContext) PopLocalFrame() {
	ctx.localBindings = ctx.localBindings.PopFrame()
}

func (b *Bindings) GetBinding(sym Symbol) *Binding {
	env := b
	for env != nil {
		if b, ok := env.bindings[sym.name]; ok {
			return b
		}
		env = env.parent
	}
	return nil
}

func (ctx *ParseContext) GetLocalBinding(sym Symbol) *Binding {
	if sym.ns != nil {
		return nil
	}
	return ctx.localBindings.GetBinding(sym)
}

func (pos Position) Pos() Position {
	return pos
}

func printError(pos Position, msg string) {
	PROBLEM_COUNT++
	fmt.Fprintf(Stderr, "%s:%d:%d: %s\n", pos.Filename(), pos.startLine, pos.startColumn, msg)
}

func printParseWarning(pos Position, msg string) {
	printError(pos, "Parse warning: "+msg)
}

func printParseError(pos Position, msg string) {
	printError(pos, "Parse error: "+msg)
}

func printReadWarning(reader *Reader, msg string) {
	pos := Position{
		filename:    reader.filename,
		startColumn: reader.column,
		startLine:   reader.line,
	}
	printError(pos, "Read warning: "+msg)
}

func printReadError(reader *Reader, msg string) {
	pos := Position{
		filename:    reader.filename,
		startColumn: reader.column,
		startLine:   reader.line,
	}
	printError(pos, "Read error: "+msg)
}

func isIgnoredUnusedNamespace(ns *Namespace) bool {
	if WARNINGS.ignoredUnusedNamespaces == nil {
		return false
	}
	ok, _ := WARNINGS.ignoredUnusedNamespaces.Get(ns.Name)
	return ok
}

func ResetUsage() {
	for _, ns := range GLOBAL_ENV.Namespaces {
		if ns == GLOBAL_ENV.CoreNamespace {
			continue
		}
		ns.isUsed = true
		for _, vr := range ns.mappings {
			vr.isUsed = true
		}
	}
}

func isEntryPointNs(ns *Namespace) bool {
	ok, _ := WARNINGS.entryPoints.Get(ns.Name)
	return ok
}

func WarnOnGloballyUnusedNamespaces() {
	var names []string
	positions := make(map[string]Position)

	for _, ns := range GLOBAL_ENV.Namespaces {
		if !ns.isGloballyUsed && !isIgnoredUnusedNamespace(ns) && !isEntryPointNs(ns) {
			pos := ns.Name.GetInfo()
			if pos != nil && pos.Filename() != "<joker.core>" && pos.Filename() != "<user>" {
				name := ns.Name.ToString(false)
				names = append(names, name)
				positions[name] = pos.Position
			}
		}
	}

	sort.Strings(names)
	for _, name := range names {
		printParseWarning(positions[name], "globally unused namespace "+name)
	}
}

func WarnOnUnusedNamespaces() {
	var names []string
	positions := make(map[string]Position)

	for _, ns := range GLOBAL_ENV.Namespaces {
		if ns != GLOBAL_ENV.CurrentNamespace() && !ns.isUsed && !isIgnoredUnusedNamespace(ns) {
			pos := ns.Name.GetInfo()
			if pos != nil && pos.Filename() != "<joker.core>" && pos.Filename() != "<user>" {
				name := ns.Name.ToString(false)
				names = append(names, name)
				positions[name] = pos.Position
			}
		}
	}

	sort.Strings(names)
	for _, name := range names {
		printParseWarning(positions[name], "unused namespace "+name)
	}
}

func isEntryPointVar(vr *Var) bool {
	if isEntryPointNs(vr.ns) {
		return true
	}
	sym := Symbol{
		ns:   vr.ns.Name.name,
		name: vr.name.name,
	}
	ok, _ := WARNINGS.entryPoints.Get(sym)
	return ok
}

func WarnOnGloballyUnusedVars() {
	var names []string
	positions := make(map[string]Position)

	for _, ns := range GLOBAL_ENV.Namespaces {
		if ns == GLOBAL_ENV.CoreNamespace {
			continue
		}
		for _, vr := range ns.mappings {
			if vr.ns == ns && !vr.isGloballyUsed && !vr.isPrivate && !isRecordConstructor(vr.name) && !isEntryPointVar(vr) {
				pos := vr.GetInfo()
				if pos != nil {
					varName := vr.Name()
					names = append(names, varName)
					positions[varName] = pos.Position
				}
			}
		}
	}

	sort.Strings(names)
	for _, name := range names {
		printParseWarning(positions[name], "globally unused var "+name)
	}
}

func WarnOnUnusedVars() {
	var names []string
	positions := make(map[string]Position)

	for _, ns := range GLOBAL_ENV.Namespaces {
		if ns == GLOBAL_ENV.CoreNamespace {
			continue
		}
		for _, vr := range ns.mappings {
			if vr.ns == ns && !vr.isUsed && vr.isPrivate {
				pos := vr.GetInfo()
				if pos != nil {
					names = append(names, *vr.name.name)
					positions[*vr.name.name] = pos.Position
				}
			}
		}
	}

	sort.Strings(names)
	for _, name := range names {
		printParseWarning(positions[name], "unused var "+name)
	}
}

func NewLiteralExpr(obj Object) *LiteralExpr {
	res := LiteralExpr{obj: obj}
	info := obj.GetInfo()
	if info != nil {
		res.Position = info.Position
	}
	return &res
}

func NewSurrogateExpr(obj Object) *LiteralExpr {
	res := NewLiteralExpr(obj)
	res.isSurrogate = true
	return res
}

func (err *ParseError) ToString(escape bool) string {
	return err.Error()
}

func (err *ParseError) Equals(other interface{}) bool {
	return err == other
}

func (err *ParseError) GetInfo() *ObjectInfo {
	return nil
}

func (err *ParseError) GetType() *Type {
	return TYPE.ParseError
}

func (err *ParseError) Hash() uint32 {
	return HashPtr(uintptr(unsafe.Pointer(err)))
}

func (err *ParseError) WithInfo(info *ObjectInfo) Object {
	return err
}

func (err *ParseError) Message() Object {
	return MakeString(err.msg)
}

func (err ParseError) Error() string {
	line, column, filename := 0, 0, "<file>"
	info := err.obj.GetInfo()
	if info != nil {
		line, column, filename = info.startLine, info.startColumn, info.Filename()
	}
	return fmt.Sprintf("%s:%d:%d: Parse error: %s", filename, line, column, err.msg)
}

func parseSeq(seq Seq, ctx *ParseContext) []Expr {
	res := make([]Expr, 0)
	for !seq.IsEmpty() {
		res = append(res, Parse(seq.First(), ctx))
		seq = seq.Rest()
	}
	return res
}

func parseVector(v *Vector, pos Position, ctx *ParseContext) Expr {
	r := make([]Expr, v.count)
	for i := 0; i < v.count; i++ {
		r[i] = Parse(v.at(i), ctx)
	}
	return &VectorExpr{
		v:        r,
		Position: pos,
	}
}

func parseMap(m Map, pos Position, ctx *ParseContext) *MapExpr {
	res := &MapExpr{
		keys:     make([]Expr, m.Count()),
		values:   make([]Expr, m.Count()),
		Position: pos,
	}
	for iter, i := m.Iter(), 0; iter.HasNext(); i++ {
		p := iter.Next()
		res.keys[i] = Parse(p.Key, ctx)
		res.values[i] = Parse(p.Value, ctx)
	}
	return res
}

func parseSet(s *MapSet, pos Position, ctx *ParseContext) Expr {
	res := &SetExpr{
		elements: make([]Expr, s.m.Count()),
		Position: pos,
	}
	for iter, i := iter(s.Seq()), 0; iter.HasNext(); i++ {
		res.elements[i] = Parse(iter.Next(), ctx)
	}
	return res
}

func checkForm(obj Object, min int, max int) int {
	seq := obj.(Seq)
	c := SeqCount(seq)
	if c < min {
		panic(&ParseError{obj: obj, msg: "Too few arguments to " + seq.First().ToString(false)})
	}
	if c > max {
		panic(&ParseError{obj: obj, msg: "Too many arguments to " + seq.First().ToString(false)})
	}
	return c
}

func GetPosition(obj Object) Position {
	info := obj.GetInfo()
	if info != nil {
		return info.Position
	}
	return Position{}
}

func updateVar(vr *Var, info *ObjectInfo, valueExpr Expr, sym Symbol) {
	vr.WithInfo(info)
	vr.expr = valueExpr
	meta := sym.GetMeta()
	if meta != nil {
		if ok, p := meta.Get(KEYWORDS.private); ok {
			vr.isPrivate = ToBool(p)
		}
		if ok, p := meta.Get(KEYWORDS.dynamic); ok {
			vr.isDynamic = ToBool(p)
		}
		vr.taggedType = getTaggedType(sym)
	}
}

func isCreatedByMacro(formSeq Seq) bool {
	return formSeq.First().GetInfo().Pos().filename == STR.coreFilename
}

func parseDef(obj Object, ctx *ParseContext, isForLinter bool) *DefExpr {
	count := checkForm(obj, 2, 4)
	seq := obj.(Seq)
	s := Second(seq)
	var meta Map
	switch sym := s.(type) {
	case Symbol:
		if sym.ns != nil && (Symbol{name: sym.ns} != ctx.GlobalEnv.CurrentNamespace().Name) {
			panic(&ParseError{
				msg: "Can't create defs outside of current ns",
				obj: obj,
			})
		}
		symWithoutNs := sym
		symWithoutNs.ns = nil
		vr := ctx.GlobalEnv.CurrentNamespace().Intern(symWithoutNs)
		if isForLinter {
			vr.isGloballyUsed = true
		}
		res := &DefExpr{
			vr:               vr,
			name:             sym,
			value:            nil,
			Position:         GetPosition(obj),
			isCreatedByMacro: isCreatedByMacro(seq),
		}
		meta = sym.GetMeta()
		if count == 3 {
			res.value = Parse(Third(seq), ctx)
		} else if count == 4 {
			res.value = Parse(Fourth(seq), ctx)
			docstring := Third(seq)
			switch docstring.(type) {
			case String:
				if meta != nil {
					meta = meta.Assoc(KEYWORDS.doc, docstring).(Map)
				} else {
					meta = EmptyArrayMap().Assoc(KEYWORDS.doc, docstring).(Map)
				}
			default:
				panic(&ParseError{obj: docstring, msg: "Docstring must be a string"})
			}
		}
		updateVar(vr, obj.GetInfo(), res.value, sym)
		if meta != nil {
			res.meta = Parse(DeriveReadObject(obj, meta), ctx)
		}
		return res
	default:
		panic(&ParseError{obj: s, msg: "First argument to def must be a Symbol"})
	}
}

func skipRedundantDo(obj Object) bool {
	if meta, ok := obj.(Meta); ok {
		if m := meta.GetMeta(); m != nil {
			if ok, res := m.Get(MakeKeyword("skip-redundant-do")); ok {
				return res.Equals(Boolean{B: true})
			}
		}
	}
	return false
}

func parseBody(seq Seq, ctx *ParseContext) []Expr {
	recur := ctx.recur
	ctx.recur = false
	defer func() { ctx.recur = recur }()
	res := make([]Expr, 0)
	for !seq.IsEmpty() {
		ro := seq.First()
		expr := Parse(ro, ctx)
		seq = seq.Rest()
		if ctx.recur && !seq.IsEmpty() && !LINTER_MODE {
			panic(&ParseError{obj: ro, msg: "Can only recur from tail position"})
		}
		res = append(res, expr)
		if LINTER_MODE {
			if defExpr, ok := expr.(*DefExpr); ok && !defExpr.isCreatedByMacro {
				printParseWarning(defExpr.Pos(), "inline def")
			} else if doExpr, ok := expr.(*DoExpr); ok && !doExpr.isCreatedByMacro && !skipRedundantDo(ro) {
				printParseWarning(doExpr.Pos(), "redundant do form")
			}
		}
	}
	return res
}

func parseParams(params Object) (bindings []Symbol, isVariadic bool) {
	res := make([]Symbol, 0)
	v := params.(*Vector)
	for i := 0; i < v.count; i++ {
		ro := v.at(i)
		sym := ro
		if !IsSymbol(sym) {
			if LINTER_MODE {
				sym = generateSymbol("linter")
			} else {
				panic(&ParseError{obj: ro, msg: "Unsupported binding form: " + sym.ToString(false)})
			}
		}
		if SYMBOLS.amp.Equals(sym) {
			if v.count > i+2 {
				ro := v.at(i + 2)
				panic(&ParseError{obj: ro, msg: "Unexpected parameter: " + ro.ToString(false)})
			}
			if v.count == i+2 {
				variadic := v.at(i + 1)
				if !IsSymbol(variadic) {
					if LINTER_MODE {
						variadic = generateSymbol("linter")
					} else {
						panic(&ParseError{obj: variadic, msg: "Unsupported binding form: " + variadic.ToString(false)})
					}
				}
				res = append(res, variadic.(Symbol))
				return res, true
			} else {
				return res, false
			}
		}
		res = append(res, sym.(Symbol))
	}
	return res, false
}

func needsUnusedWarning(b *Binding) bool {
	return !b.isUsed &&
		!strings.HasPrefix(*b.name.name, "_") &&
		!strings.HasPrefix(*b.name.name, "&form") &&
		!strings.HasPrefix(*b.name.name, "&env") &&
		!isSkipUnused(b.name)
}

func addArity(fn *FnExpr, sig Seq, ctx *ParseContext) {
	params := sig.First()
	body := sig.Rest()
	args, isVariadic := parseParams(params)
	ctx.PushLocalFrame(args)
	defer ctx.PopLocalFrame()
	ctx.PushLoopBindings(args)
	defer ctx.PopLoopBindings()

	noRecurAllowed := ctx.noRecurAllowed
	ctx.noRecurAllowed = false
	defer func() { ctx.noRecurAllowed = noRecurAllowed }()

	arity := FnArityExpr{
		Position:   GetPosition(sig),
		args:       args,
		body:       parseBody(body, ctx),
		taggedType: getTaggedType(params.(Meta)),
	}
	if isVariadic {
		if fn.variadic != nil {
			panic(&ParseError{obj: params, msg: "Can't have more than 1 variadic overload"})
		}
		for _, arity := range fn.arities {
			if len(arity.args) >= len(args) {
				panic(&ParseError{obj: params, msg: "Can't have fixed arity function with more params than variadic function"})
			}
		}
		fn.variadic = &arity
	} else {
		for _, arity := range fn.arities {
			if len(arity.args) == len(args) {
				panic(&ParseError{obj: params, msg: "Can't have 2 overloads with same arity"})
			}
		}
		if fn.variadic != nil && len(args) >= len(fn.variadic.args) {
			panic(&ParseError{obj: params, msg: "Can't have fixed arity function with more params than variadic function"})
		}
		fn.arities = append(fn.arities, arity)
	}

	if LINTER_MODE {
		if WARNINGS.fnWithEmptyBody {
			if len(arity.body) == 0 {
				printParseWarning(arity.Position, "fn form with empty body")
			}
		}

		if WARNINGS.unusedFnParameters {
			var unused []Symbol
			for _, b := range ctx.localBindings.bindings {
				if needsUnusedWarning(b) {
					unused = append(unused, b.name)
				}
			}
			sort.Sort(BySymbolName(unused))
			for _, u := range unused {
				printParseWarning(GetPosition(u), "unused parameter: "+u.ToString(false))
			}
		}
	}
}

func wrapWithMeta(fnExpr *FnExpr, obj Object, ctx *ParseContext) Expr {
	meta := obj.(Meta).GetMeta()
	if meta != nil {
		return &MetaExpr{
			meta:     parseMap(meta, fnExpr.Pos(), ctx),
			expr:     fnExpr,
			Position: fnExpr.Pos(),
		}
	}
	return fnExpr
}

// Examples:
// (fn f [] 1 2)
// (fn f ([] 1 2)
//
//	([a] a 3)
//	([a & b] a b))
func parseFn(obj Object, ctx *ParseContext) Expr {
	res := &FnExpr{Position: GetPosition(obj)}
	bodies := obj.(Seq).Rest()
	p := bodies.First()
	if IsSymbol(p) { // self reference
		res.self = p.(Symbol)
		bodies = bodies.Rest()
		p = bodies.First()
		ctx.PushLocalFrame([]Symbol{res.self})
		defer ctx.PopLocalFrame()
	}
	if IsVector(p) { // single arity
		addArity(res, bodies, ctx)
		return wrapWithMeta(res, obj, ctx)
	}
	// multiple arities
	if bodies.IsEmpty() {
		panic(&ParseError{obj: p, msg: "Parameter declaration missing"})
	}
	for !bodies.IsEmpty() {
		body := bodies.First()
		switch s := body.(type) {
		case Seq:
			params := s.First()
			if !IsVector(params) {
				panic(&ParseError{obj: params, msg: "Parameter declaration must be a vector. Got: " + params.ToString(false)})
			}
			addArity(res, s, ctx)
		default:
			panic(&ParseError{obj: body, msg: "Function body must be a list. Got: " + s.ToString(false)})
		}
		bodies = bodies.Rest()
	}
	return wrapWithMeta(res, obj, ctx)
}

func isCatch(obj Object) bool {
	return IsSeq(obj) && obj.(Seq).First().Equals(SYMBOLS.catch)
}

func isFinally(obj Object) bool {
	return IsSeq(obj) && obj.(Seq).First().Equals(SYMBOLS.finally)
}

func resolveType(obj Object, ctx *ParseContext) *Type {
	excType := Parse(obj, ctx)
	switch excType := excType.(type) {
	case *LiteralExpr:
		switch t := excType.obj.(type) {
		case *Type:
			return t
		}
	}
	if LINTER_MODE {
		return TYPE.Error
	}
	panic(&ParseError{obj: obj, msg: "Unable to resolve type: " + obj.ToString(false)})
}

func parseCatch(obj Object, ctx *ParseContext) *CatchExpr {
	seq := obj.(Seq).Rest()
	if seq.IsEmpty() || seq.Rest().IsEmpty() {
		panic(&ParseError{obj: obj, msg: "catch requires at least two arguments: type symbol and binding symbol"})
	}
	excSymbol := Second(seq)
	excType := resolveType(seq.First(), ctx)
	if !IsSymbol(excSymbol) {
		panic(&ParseError{obj: excSymbol, msg: "Bad binding form, expected symbol, got: " + excSymbol.ToString(false)})
	}
	ctx.PushLocalFrame([]Symbol{excSymbol.(Symbol)})
	defer ctx.PopLocalFrame()
	return &CatchExpr{
		Position:  GetPosition(obj),
		excType:   excType,
		excSymbol: excSymbol.(Symbol),
		body:      parseBody(seq.Rest().Rest(), ctx),
	}
}

func parseFinally(body Seq, ctx *ParseContext) []Expr {
	return parseBody(body, ctx)
}

func parseTry(obj Object, ctx *ParseContext) *TryExpr {
	const (
		Regular = iota
		Catch   = iota
		Finally = iota
	)
	res := &TryExpr{Position: GetPosition(obj)}
	lastType := Regular
	seq := obj.(Seq).Rest()

	noRecurAllowed := ctx.noRecurAllowed
	ctx.noRecurAllowed = true
	defer func() { ctx.noRecurAllowed = noRecurAllowed }()

	for !seq.IsEmpty() {
		obj = seq.First()
		if lastType == Finally {
			panic(&ParseError{obj: obj, msg: "finally clause must be last in try expression"})
		}
		if isCatch(obj) {
			res.catches = append(res.catches, parseCatch(obj, ctx))
			lastType = Catch
		} else if isFinally(obj) {
			res.finallyExpr = parseFinally(obj.(Seq).Rest(), ctx)
			lastType = Finally
		} else {
			if lastType == Catch {
				panic(&ParseError{obj: obj, msg: "Only catch or finally clause can follow catch in try expression"})
			}
			res.body = append(res.body, Parse(obj, ctx))
		}
		seq = seq.Rest()
	}
	if LINTER_MODE {
		if res.body == nil {
			printParseWarning(res.Pos(), "try form with empty body")
		}
		if res.catches == nil && res.finallyExpr == nil {
			printParseWarning(res.Pos(), "try form without catch or finally")
		}
		if res.finallyExpr != nil && len(res.finallyExpr) == 0 {
			printParseWarning(GetPosition(obj), "finally form with empty body")
		}
	}
	return res
}

func parseLet(obj Object, ctx *ParseContext) *LetExpr {
	return parseLetLoop(obj, "let", ctx)
}

func parseLoop(obj Object, ctx *ParseContext) *LoopExpr {
	return (*LoopExpr)(parseLetLoop(obj, "loop", ctx))
}

func parseLetfn(obj Object, ctx *ParseContext) *LoopExpr {
	return (*LoopExpr)(parseLetLoop(obj, "letfn", ctx))
}

func isSkipUnused(obj Meta) bool {
	if m := obj.GetMeta(); m != nil {
		if ok, v := m.Get(KEYWORDS.skipUnused); ok {
			return ToBool(v)
		}
	}
	return false
}

func parseLetLoop(obj Object, formName string, ctx *ParseContext) *LetExpr {
	res := &LetExpr{
		Position: GetPosition(obj),
	}
	bindings := Second(obj.(Seq))
	switch b := bindings.(type) {
	case *Vector:
		if b.count%2 != 0 {
			panic(&ParseError{obj: bindings, msg: formName + " requires an even number of forms in binding vector"})
		}
		if LINTER_MODE && formName != "loop" && b.count == 0 {
			pos := GetPosition(obj)
			printParseWarning(pos, formName+" form with empty bindings vector")
		}
		skipUnused := isSkipUnused(b)
		res.names = make([]Symbol, b.count/2)
		res.values = make([]Expr, b.count/2)
		ctx.PushEmptyLocalFrame()
		defer ctx.PopLocalFrame()

		for i := 0; i < b.count/2; i++ {
			s := b.at(i * 2)
			switch sym := s.(type) {
			case Symbol:
				if sym.ns != nil {
					msg := "Can't let qualified name: " + sym.ToString(false)
					if LINTER_MODE {
						printParseError(GetPosition(s), msg)
					} else {
						panic(&ParseError{obj: s, msg: msg})
					}
				}
				res.names[i] = sym
			default:
				if LINTER_MODE {
					res.names[i] = generateSymbol("linter")
				} else {
					panic(&ParseError{obj: s, msg: "Unsupported binding form: " + sym.ToString(false)})
				}
			}
			var inferredType *Type
			if formName != "letfn" {
				res.values[i] = Parse(b.at(i*2+1), ctx)
				if LINTER_MODE {
					inferredType = res.values[i].InferType()
				}
			}
			ctx.localBindings.AddBinding(res.names[i], i, skipUnused, inferredType)
		}

		if formName == "letfn" {
			for i := 0; i < b.count/2; i++ {
				res.values[i] = Parse(b.at(i*2+1), ctx)
			}
		}

		if formName == "loop" {
			ctx.PushLoopBindings(res.names)
			defer ctx.PopLoopBindings()

			noRecurAllowed := ctx.noRecurAllowed
			ctx.noRecurAllowed = false
			defer func() { ctx.noRecurAllowed = noRecurAllowed }()
		}

		res.body = parseBody(obj.(Seq).Rest().Rest(), ctx)

		if LINTER_MODE {
			if len(res.body) == 0 {
				pos := GetPosition(obj)
				printParseWarning(pos, formName+" form with empty body")
			}

			if !skipUnused {
				var unused []Symbol
				for _, b := range ctx.localBindings.bindings {
					if needsUnusedWarning(b) {
						unused = append(unused, b.name)
					}
				}
				sort.Sort(BySymbolName(unused))
				for _, u := range unused {
					printParseWarning(GetPosition(u), "unused binding: "+u.ToString(false))
				}
			}
		}

	default:
		panic(&ParseError{obj: obj, msg: formName + " requires a vector for its bindings"})
	}
	return res
}

func parseRecur(obj Object, ctx *ParseContext) *RecurExpr {
	if ctx.noRecurAllowed {
		panic(&ParseError{obj: obj, msg: "Cannot recur across try"})
	}
	loopBindings := ctx.GetLoopBindings()
	if loopBindings == nil {
		panic(&ParseError{obj: obj, msg: "No recursion point for recur"})
	}
	seq := obj.(Seq)
	args := parseSeq(seq.Rest(), ctx)
	if len(loopBindings) != len(args) {
		panic(&ParseError{obj: obj, msg: fmt.Sprintf("Mismatched argument count to recur, expected: %d args, got: %d", len(loopBindings), len(args))})
	}
	ctx.recur = true
	return &RecurExpr{
		args:     args,
		Position: GetPosition(obj),
	}
}

func resolveMacro(obj Object, ctx *ParseContext) *Var {
	switch sym := obj.(type) {
	case Symbol:
		if ctx.GetLocalBinding(sym) != nil {
			return nil
		}
		vr, ok := ctx.GlobalEnv.Resolve(sym)
		if !ok || !vr.isMacro || vr.Value == nil {
			return nil
		}
		vr.isUsed = true
		vr.isGloballyUsed = true
		if vr.ns == nil {
			// This very likely represents a Joker
			// bug. E.g. often seen while developing the
			// fast-init (fast-startup) version of
			// Joker. But it's much easier to debug when
			// presented as a parse error (so the
			// "offending" .joke source info is provided)
			// along with the problematic var name.
			panic(&ParseError{obj: obj, msg: fmt.Sprintf("No namespace for %s", vr.name.ToString(false))})
		}
		vr.ns.isUsed = true
		vr.ns.isGloballyUsed = true
		return vr
	default:
		return nil
	}
}

func fixInfo(obj Object, info *ObjectInfo) Object {
	switch s := obj.(type) {
	case Nil:
		return obj
	case Seq:
		objs := make([]Object, 0, 8)
		for !s.IsEmpty() {
			t := fixInfo(s.First(), info)
			objs = append(objs, t)
			s = s.Rest()
		}
		res := NewListFrom(objs...)
		if s, ok := obj.(Meta); ok {
			res.meta = s.GetMeta()
		}
		if objInfo := obj.GetInfo(); objInfo != nil {
			return res.WithInfo(objInfo)
		}
		return res.WithInfo(info)
	case *Vector:
		var res Conjable = EmptyVector()
		for i := 0; i < s.count; i++ {
			t := fixInfo(s.at(i), info)
			res = res.Conj(t)
		}
		res.(*Vector).meta = s.meta
		if objInfo := obj.GetInfo(); objInfo != nil {
			return res.WithInfo(objInfo)
		}
		return res.WithInfo(info)
	case Map:
		res := EmptyArrayMap()
		iter := s.Iter()
		for iter.HasNext() {
			p := iter.Next()
			key := fixInfo(p.Key, info)
			value := fixInfo(p.Value, info)
			res.Add(key, value)
		}
		res.meta = s.(Meta).GetMeta()
		if objInfo := obj.GetInfo(); objInfo != nil {
			return res.WithInfo(objInfo)
		}
		return res.WithInfo(info)
	default:
		return obj
	}
}

func macroexpand1(seq Seq, ctx *ParseContext) Object {
	op := seq.First()
	vr := resolveMacro(op, ctx)
	if vr != nil {
		expr := &MacroCallExpr{
			Position: GetPosition(seq),
			macro:    vr.Value.(Callable),
			args:     ToSlice(seq.Rest().Cons(ctx.localBindings.ToMap()).Cons(seq)),
			name:     varCallableString(vr),
		}
		return fixInfo(Eval(expr, nil), seq.GetInfo())
	} else {
		return seq
	}
}

func reportNotAFunction(pos Position, name string) {
	printParseWarning(pos, name+" is not a function")
}

func getTaggedType(obj Meta) *Type {
	if m := obj.GetMeta(); m != nil {
		if ok, typeName := m.Get(KEYWORDS.tag); ok {
			if typeSym, ok := typeName.(Symbol); ok {
				if t := TYPES[typeSym.name]; t != nil {
					return t
				}
			}
		}
	}
	return nil
}

func getTaggedTypes(obj Meta) []*Type {
	var res []*Type
	if m := obj.GetMeta(); m != nil {
		if ok, typeName := m.Get(KEYWORDS.tag); ok {
			switch typeDecl := typeName.(type) {
			case Symbol:
				if t := TYPES[typeDecl.name]; t != nil {
					res = append(res, t)
				}
			case String:
				parts := strings.Split(typeDecl.S, "|")
				for _, p := range parts {
					if t := TYPES[MakeSymbol(p).name]; t != nil {
						res = append(res, t)
					}
				}
			}
		}
	}
	return res
}

func isTypeOneOf(abstractTypes []*Type, concreteType *Type) bool {
	for _, t := range abstractTypes {
		if IsEqualOrImplements(t, concreteType) {
			return true
		}
	}
	return false
}

func typesString(types []*Type) string {
	var b bytes.Buffer
	for i, t := range types {
		b.WriteString(t.ToString(false))
		if i < len(types)-1 {
			b.WriteString(" or ")
		}
	}
	return b.String()
}

func checkTypes(declaredArgs []Symbol, call *CallExpr) bool {
	res := false
	for i, da := range declaredArgs {
		if declaredTypes := getTaggedTypes(da); len(declaredTypes) > 0 {
			passedType := call.args[i].InferType()
			if passedType != nil {
				if !isTypeOneOf(declaredTypes, passedType) {
					printParseWarning(call.args[i].Pos(), fmt.Sprintf("arg[%d] of %s must have type %s, got %s", i, call.Name(), typesString(declaredTypes), passedType.ToString(false)))
					res = true
				}
			}
		}
	}
	return res
}

func selectArity(expr *FnExpr, passedArgsCount int) *FnArityExpr {
	for _, arity := range expr.arities {
		if len(arity.args) == passedArgsCount {
			return &arity
		}
	}
	if expr.variadic != nil && passedArgsCount >= len(expr.variadic.args)-1 {
		return expr.variadic
	}
	return nil
}

func reportWrongArity(expr *FnExpr, isMacro bool, call *CallExpr, pos Position) bool {
	passedArgsCount := len(call.args)
	if isMacro {
		passedArgsCount += 2
	}
	if v := selectArity(expr, passedArgsCount); v != nil {
		return checkTypes(v.args, call)
	}
	printParseWarning(pos, fmt.Sprintf("Wrong number of args (%d) passed to %s", len(call.args), call.Name()))
	return true
}

func checkArglist(arglist Seq, passedArgsCount int) bool {
	for !arglist.IsEmpty() {
		if v, ok := arglist.First().(*Vector); ok {
			if v.Count() == passedArgsCount ||
				v.Count() >= 2 && v.Nth(v.Count()-2).Equals(SYMBOLS.amp) && passedArgsCount >= (v.Count()-2) {
				return true
			}
		}
		arglist = arglist.Rest()
	}
	return false
}

func setMacroMeta(vr *Var) {
	if vr.meta == nil {
		vr.meta = EmptyArrayMap().Assoc(KEYWORDS.macro, Boolean{B: true}).(Map)
	} else {
		vr.meta = vr.meta.Assoc(KEYWORDS.macro, Boolean{B: true}).(Map)
	}
}

func parseSetMacro(obj Object, ctx *ParseContext) Expr {
	expr := Parse(Second(obj.(Seq)), ctx)
	switch expr := expr.(type) {
	case *LiteralExpr:
		switch vr := expr.obj.(type) {
		case *Var:
			res := &SetMacroExpr{
				vr: vr,
			}
			res.Eval(nil)
			return res
		}
	}
	panic(&ParseError{obj: obj, msg: "set-macro__ argument must be a var"})
}

func isKnownMacros(sym Symbol) (bool, Seq) {
	if KNOWN_MACROS == nil {
		knownMacros := GLOBAL_ENV.CoreNamespace.Resolve("*known-macros*")
		if knownMacros == nil {
			return false, nil
		}
		KNOWN_MACROS = knownMacros
	}
	if ok, v := KNOWN_MACROS.Value.(Map).Get(sym); ok {
		switch v := v.(type) {
		case Seqable:
			return true, v.Seq()
		default:
			return true, nil
		}
	}
	return false, nil
}

func isUnknownCallable(expr Expr) (bool, Seq) {
	if !LINTER_MODE {
		return false, nil
	}
	if c, ok := expr.(*VarRefExpr); ok {
		if c.vr.isMacro {
			return true, nil
		}
		var sym Symbol
		if c.vr.ns != GLOBAL_ENV.CurrentNamespace() && c.vr.ns != GLOBAL_ENV.CoreNamespace {
			sym = Symbol{
				ns:   c.vr.ns.Name.name,
				name: c.vr.name.name,
			}
		} else {
			sym = MakeSymbol(*c.vr.name.name)
		}
		b, s := isKnownMacros(sym)
		if b {
			return b, s
		}
		if c.vr.expr != nil {
			return false, nil
		}
		if sym.ns == nil && c.vr.isFake && c.vr.ns != GLOBAL_ENV.CoreNamespace {
			return true, nil
		}
	}
	return false, nil
}

func areAllLiteralExprs(exprs []Expr) bool {
	for _, expr := range exprs {
		if _, ok := expr.(*LiteralExpr); !ok {
			return false
		}
	}
	return true
}

func getRequireVar(ctx *ParseContext) *Var {
	if REQUIRE_VAR == nil {
		REQUIRE_VAR = ctx.GlobalEnv.CoreNamespace.Resolve("require")
	}
	return REQUIRE_VAR
}

func getReferVar(ctx *ParseContext) *Var {
	if REFER_VAR == nil {
		REFER_VAR = ctx.GlobalEnv.CoreNamespace.Resolve("refer")
	}
	return REFER_VAR
}

func getAliasVar(ctx *ParseContext) *Var {
	if ALIAS_VAR == nil {
		ALIAS_VAR = ctx.GlobalEnv.CoreNamespace.Resolve("alias")
	}
	return ALIAS_VAR
}

func getCreateNsVar(ctx *ParseContext) *Var {
	if CREATE_NS_VAR == nil {
		CREATE_NS_VAR = ctx.GlobalEnv.CoreNamespace.Resolve("create-ns")
	}
	return CREATE_NS_VAR
}

func getInNsVar(ctx *ParseContext) *Var {
	if IN_NS_VAR == nil {
		IN_NS_VAR = ctx.GlobalEnv.CoreNamespace.Resolve("in-ns")
	}
	return IN_NS_VAR
}

func checkCall(expr Expr, isMacro bool, call *CallExpr, pos Position) {
	argsCount := len(call.args)
	switch expr := expr.(type) {
	case *FnExpr:
		reportWrongArity(expr, isMacro, call, pos)
	case *MapExpr:
		if argsCount == 0 || argsCount > 2 {
			printParseWarning(pos, fmt.Sprintf("Wrong number of args (%d) passed to a map", argsCount))
		}
	case *SetExpr:
		if argsCount == 0 || argsCount > 1 {
			printParseWarning(pos, fmt.Sprintf("Wrong number of args (%d) passed to a set", argsCount))
		}
	case *LiteralExpr:
		if _, ok := expr.obj.(Callable); !ok && !expr.isSurrogate {
			reportNotAFunction(pos, call.Name())
			return
		}
		switch expr.obj.(type) {
		case Keyword:
			if argsCount == 0 || argsCount > 2 {
				printParseWarning(pos, fmt.Sprintf("Wrong number of args (%d) passed to %s", argsCount, call.Name()))
			}
		}
	case *RecurExpr:
		reportNotAFunction(pos, call.Name())
	case *ThrowExpr:
		reportNotAFunction(pos, call.Name())
	}
}

func parseList(obj Object, ctx *ParseContext) Expr {
	expanded := macroexpand1(obj.(Seq), ctx)
	if expanded != obj {
		return Parse(expanded, ctx)
	}
	seq := obj.(Seq)
	if seq.IsEmpty() {
		return NewLiteralExpr(obj)
	}

	currentIsUnknownCallableScope := ctx.isUnknownCallableScope
	defer func() {
		ctx.isUnknownCallableScope = currentIsUnknownCallableScope
	}()

	ctx.isUnknownCallableScope = false

	pos := GetPosition(obj)
	first := seq.First()
	if v, ok := first.(Symbol); ok && v.ns == nil {
		switch v.name {
		case STR.quote:
			return NewLiteralExpr(Second(seq))
		case STR._if:
			checkForm(obj, 3, 4)
			if LINTER_MODE && SeqCount(seq) < 4 && WARNINGS.ifWithoutElse {
				printParseWarning(pos, "missing else branch")
			}
			return &IfExpr{
				cond:     Parse(Second(seq), ctx),
				positive: Parse(Third(seq), ctx),
				negative: Parse(Fourth(seq), ctx),
				Position: pos,
			}
		case STR.fn_:
			return parseFn(obj, ctx)
		case STR.let_:
			return parseLet(obj, ctx)
		case STR.letfn_:
			return parseLetfn(obj, ctx)
		case STR.loop_:
			return parseLoop(obj, ctx)
		case STR.recur:
			return parseRecur(obj, ctx)

		// Vars' isMacro has to be properly set during parse stage
		// for linter mode to correctly handle arguments count.
		case STR.setMacro_:
			return parseSetMacro(obj, ctx)

		case STR.def:
			return parseDef(obj, ctx, false)
		case STR.defLinter:
			return parseDef(obj, ctx, true)
		case STR._var:
			checkForm(obj, 2, 2)
			switch sym := Second(seq).(type) {
			case Symbol:
				vr, ok := ctx.GlobalEnv.Resolve(sym)
				if !ok {
					if !LINTER_MODE {
						panic(&ParseError{obj: obj, msg: "Unable to resolve var " + sym.ToString(false) + " in this context"})
					}
					symNs := ctx.GlobalEnv.NamespaceFor(ctx.GlobalEnv.CurrentNamespace(), sym)
					if !ctx.isUnknownCallableScope {
						if symNs == nil || symNs == ctx.GlobalEnv.CurrentNamespace() {
							printParseError(GetPosition(obj), "Unable to resolve symbol: "+sym.ToString(false))
						}
					}
					vr = InternFakeSymbol(symNs, sym)
				}
				vr.isUsed = true
				vr.isGloballyUsed = true
				vr.ns.isUsed = true
				vr.ns.isGloballyUsed = true
				return &LiteralExpr{
					obj:      vr,
					Position: pos,
				}
			default:
				panic(&ParseError{obj: obj, msg: "var's argument must be a symbol"})
			}
		case STR.do:
			res := &DoExpr{
				body:             parseBody(seq.Rest(), ctx),
				Position:         pos,
				isCreatedByMacro: isCreatedByMacro(seq),
			}
			if LINTER_MODE {
				if len(res.body) == 0 {
					printParseWarning(pos, "do form with empty body")
				} else if len(res.body) == 1 {
					printParseWarning(pos, "redundant do form")
				}
			}
			return res
		case STR.throw:
			return &ThrowExpr{
				Position: pos,
				e:        Parse(Second(seq), ctx),
			}
		case STR.try:
			return parseTry(obj, ctx)
		}
	}

	ctx.isUnknownCallableScope = currentIsUnknownCallableScope
	callable := Parse(first, ctx)
	unknown, syms := isUnknownCallable(callable)
	if unknown {
		ctx.isUnknownCallableScope = true
		if syms != nil {
			ctx.linterBindings = ctx.linterBindings.PushFrame()
			defer func() {
				ctx.linterBindings = ctx.linterBindings.PopFrame()
			}()
			for !syms.IsEmpty() {
				if sym, ok := syms.First().(Symbol); ok {
					ctx.linterBindings.AddBinding(sym, 0, true, nil)
				}
				syms = syms.Rest()
			}
		}
	} else {
		ctx.isUnknownCallableScope = false
	}
	res := &CallExpr{
		callable: callable,
		args:     parseSeq(seq.Rest(), ctx),
		Position: pos,
	}
	if LINTER_MODE {
		switch c := res.callable.(type) {
		case *VarRefExpr:
			if c.vr.Value != nil {
				switch f := c.vr.Value.(type) {
				case *Fn:
					if !reportWrongArity(f.fnExpr, c.vr.isMacro, res, pos) {
						require := getRequireVar(ctx)
						refer := getReferVar(ctx)
						alias := getAliasVar(ctx)
						createNs := getCreateNsVar(ctx)
						inNs := getInNsVar(ctx)
						if (c.vr.Value.Equals(require.Value) ||
							c.vr.Value.Equals(alias.Value) ||
							c.vr.Value.Equals(refer.Value) ||
							c.vr.Value.Equals(inNs.Value) ||
							c.vr.Value.Equals(createNs.Value)) &&
							areAllLiteralExprs(res.args) {
							Eval(res, nil)
						}
					}
				case Callable:
					if m := c.vr.GetMeta(); m != nil {
						if ok, arglist := m.Get(KEYWORDS.arglist); ok {
							if arglist, ok := arglist.(Seq); ok {
								if !checkArglist(arglist, len(res.args)) {
									printParseWarning(pos, fmt.Sprintf("Wrong number of args (%d) passed to %s", len(res.args), res.Name()))
								}
							}
						}
					}
					return res
				default:
					reportNotAFunction(pos, res.Name())
				}
			} else {
				checkCall(c.vr.expr, c.vr.isMacro, res, pos)
			}
		default:
			checkCall(res.callable, false, res, pos)
		}
	}
	return res
}

func InternFakeSymbol(ns *Namespace, sym Symbol) *Var {
	if ns != nil {
		fakeSym := Symbol{
			ns:   nil,
			name: sym.name,
		}
		return ns.InternFake(fakeSym)
	}
	fakeSym := Symbol{
		ns:   nil,
		name: STRINGS.Intern(sym.ToString(false)),
	}
	return GLOBAL_ENV.CurrentNamespace().InternFake(fakeSym)
}

func isInteropSymbol(sym Symbol) bool {
	return sym.ns == nil && (strings.HasPrefix(*sym.name, ".") || strings.HasSuffix(*sym.name, ".") || strings.Contains(*sym.name, "$"))
}

func isRecordConstructor(sym Symbol) bool {
	return sym.ns == nil && (strings.HasPrefix(*sym.name, "->") || strings.HasPrefix(*sym.name, "map->"))
}

var fullClassNameRe = regexp.MustCompile(`.+\..+\.[A-Z].+`)

func isJavaSymbol(sym Symbol) bool {
	s := *sym.name
	if sym.ns != nil {
		s = *sym.ns
	}
	return fullClassNameRe.MatchString(s)
}

func MakeVarRefExpr(vr *Var, obj Object) *VarRefExpr {
	vr.isUsed = true
	vr.isGloballyUsed = true
	vr.ns.isUsed = true
	vr.ns.isGloballyUsed = true
	return &VarRefExpr{
		vr:       vr,
		Position: GetPosition(obj),
	}
}

func parseSymbol(obj Object, ctx *ParseContext) Expr {
	sym := obj.(Symbol)
	b := ctx.GetLocalBinding(sym)
	if b != nil {
		b.isUsed = true
		return &BindingExpr{
			binding:  b,
			Position: GetPosition(obj),
		}
	}
	if vr, ok := ctx.GlobalEnv.Resolve(sym); ok {
		return MakeVarRefExpr(vr, obj)
	}
	if sym.ns == nil && TYPES[sym.name] != nil {
		return &LiteralExpr{
			Position: GetPosition(obj),
			obj:      TYPES[sym.name],
		}
	}
	if !LINTER_MODE {
		panic(&ParseError{obj: obj, msg: "Unable to resolve symbol: " + sym.ToString(false)})
	}
	if DIALECT == CLJS && sym.ns == nil {
		// Check if this is a "callable namespace"
		ns := ctx.GlobalEnv.FindNamespace(sym)
		if ns == nil {
			ns = ctx.GlobalEnv.CurrentNamespace().aliases[sym.name]
		}
		if ns != nil {
			ns.isUsed = true
			ns.isGloballyUsed = true
			return NewSurrogateExpr(obj)
		}
		// See if this is a JS interop (i.e. Math.PI)
		parts := strings.Split(sym.Name(), ".")
		if len(parts) > 1 && parts[0] != "" && parts[len(parts)-1] != "" {
			return parseSymbol(DeriveReadObject(obj, MakeSymbol(strings.Join(parts[:len(parts)-1], "."))), ctx)
		}
		// Check if this is a constructor call
		if len(parts) == 2 && parts[0] != "" && parts[len(parts)-1] == "" {
			if vr, ok := ctx.GlobalEnv.Resolve(MakeSymbol(parts[0])); ok {
				return MakeVarRefExpr(vr, obj)
			}
		}
	}
	symNs := ctx.GlobalEnv.NamespaceFor(ctx.GlobalEnv.CurrentNamespace(), sym)
	if symNs == nil || symNs == ctx.GlobalEnv.CurrentNamespace() {
		if isInteropSymbol(sym) || isJavaSymbol(sym) {
			return NewSurrogateExpr(sym)
		}
		if !ctx.isUnknownCallableScope {
			if ctx.linterBindings.GetBinding(sym) == nil {
				printParseError(GetPosition(obj), "Unable to resolve symbol: "+sym.ToString(false))
			}
		}
	}
	return MakeVarRefExpr(InternFakeSymbol(symNs, sym), obj)
}

func Parse(obj Object, ctx *ParseContext) Expr {
	pos := GetPosition(obj)
	var res Expr
	canHaveMeta := false
	switch v := obj.(type) {
	case Nil:
		res = NewLiteralExpr(obj)
	case *Vector:
		canHaveMeta = true
		res = parseVector(v, pos, ctx)
	case Map:
		canHaveMeta = true
		res = parseMap(v, pos, ctx)
	case *MapSet:
		canHaveMeta = true
		res = parseSet(v, pos, ctx)
	case Seq:
		res = parseList(obj, ctx)
	case Symbol:
		res = parseSymbol(obj, ctx)
	default:
		res = NewLiteralExpr(obj)
	}
	if canHaveMeta {
		meta := obj.(Meta).GetMeta()
		if meta != nil {
			return &MetaExpr{
				meta:     parseMap(meta, pos, ctx),
				expr:     res,
				Position: pos,
			}
		}
	}
	return res
}

func TryParse(obj Object, ctx *ParseContext) (expr Expr, err error) {
	defer func() {
		if r := recover(); r != nil {
			PROBLEM_COUNT++
			switch r.(type) {
			case *ParseError:
				err = r.(error)
			case *EvalError:
				err = r.(error)
			case *ExInfo:
				err = r.(error)
			default:
				panic(r)
			}
		}
	}()
	return Parse(obj, ctx), nil
}
