// This file is generated by generate-std.joke script. Do not edit manually!

package json

import (
	. "github.com/candid82/joker/core"
)

var jsonNamespace = GLOBAL_ENV.EnsureNamespace(MakeSymbol("joker.json"))



var read_string_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 1:
		s := ExtractString(_args, 0)
		_res := readString(s)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var write_string_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 1:
		v := ExtractObject(_args, 0)
		_res := writeString(v)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

func init() {

	jsonNamespace.ResetMeta(MakeMeta(nil, "Implements encoding and decoding of JSON as defined in RFC 4627.", "1.0"))

	
	jsonNamespace.InternVar("read-string", read_string_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("s"))),
			`Parses the JSON-encoded data and return the result as a Joker value.`, "1.0"))

	jsonNamespace.InternVar("write-string", write_string_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("v"))),
			`Returns the JSON encoding of v.`, "1.0"))

}
