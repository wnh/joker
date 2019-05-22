// This file is generated by generate-std.joke script. Do not edit manually!

package os

import (
	. "github.com/candid82/joker/core"
)

var osNamespace = GLOBAL_ENV.EnsureNamespace(MakeSymbol("joker.os"))



var args_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 0:
		_res := commandArgs()
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var chdir_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 1:
		dirname := ExtractString(_args, 0)
		_res := chdir(dirname)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var cwd_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 0:
		_res := getwd()
		return MakeString(_res)

	default:
		PanicArity(_c)
	}
	return NIL
}

var env_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 0:
		_res := env()
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var exec_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 2:
		name := ExtractString(_args, 0)
		opts := ExtractMap(_args, 1)
		_res := execute(name, opts)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var exit_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 1:
		code := ExtractInt(_args, 0)
		_res := NIL
		ExitJoker(code)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var ls_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 1:
		dirname := ExtractString(_args, 0)
		_res := readDir(dirname)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var mkdir_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 2:
		name := ExtractString(_args, 0)
		perm := ExtractInt(_args, 1)
		_res := mkdir(name, perm)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var set_env_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 2:
		key := ExtractString(_args, 0)
		value := ExtractString(_args, 1)
		_res := setEnv(key, value)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var sh_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case true:
		CheckArity(_args, 1, 999)
		name := ExtractString(_args, 0)
		arguments := ExtractStrings(_args, 1)
		_res := sh("", nil, name, arguments)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var sh_from_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case true:
		CheckArity(_args, 2, 999)
		dir := ExtractString(_args, 0)
		name := ExtractString(_args, 1)
		arguments := ExtractStrings(_args, 2)
		_res := sh(dir, nil, name, arguments)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

var stat_ Proc = func(_args []Object) Object {
	_c := len(_args)
	switch {
	case _c == 1:
		filename := ExtractString(_args, 0)
		_res := stat(filename)
		return _res

	default:
		PanicArity(_c)
	}
	return NIL
}

func init() {

	osNamespace.ResetMeta(MakeMeta(nil, "Provides a platform-independent interface to operating system functionality.", "1.0"))

	
	osNamespace.InternVar("args", args_,
		MakeMeta(
			NewListFrom(NewVectorFrom()),
			`Returns a sequence of the command line arguments, starting with the program name (normally, joker).`, "1.0"))

	osNamespace.InternVar("chdir", chdir_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("dirname"))),
			`Chdir changes the current working directory to the named directory. If there is an error, an exception will be thrown. Returns nil.`, "1.0"))

	osNamespace.InternVar("cwd", cwd_,
		MakeMeta(
			NewListFrom(NewVectorFrom()),
			`Returns a rooted path name corresponding to the current directory. If the current directory can
  be reached via multiple paths (due to symbolic links), cwd may return any one of them.`, "1.0"))

	osNamespace.InternVar("env", env_,
		MakeMeta(
			NewListFrom(NewVectorFrom()),
			`Returns a map representing the environment.`, "1.0"))

	osNamespace.InternVar("exec", exec_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("name"), MakeSymbol("opts"))),
			`Executes the named program with the given arguments. opts is a map with the following keys (all optional):
  :args - vector of arguments (all arguments must be strings),
  :dir - if specified, working directory will be set to this value before executing the program,
  :stdin - if specified, provides stdin for the program. Can be either a string or :pipe keyword.
  If it's a string, the string's content will serve as stdin for the program. If it's :pipe,
  Joker's stdin will be redirected to the program's stdin.
  Returns a map with the following keys:
  :success - whether or not the execution was successful,
  :err-msg (present iff :success if false) - string capturing error object returned by Go runtime
  :exit - exit code of program (or attempt to execute it),
  :out - string capturing stdout of the program,
  :err - string capturing stderr of the program.`, "1.0"))

	osNamespace.InternVar("exit", exit_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("code"))),
			`Causes the current program to exit with the given status code.`, "1.0"))

	osNamespace.InternVar("ls", ls_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("dirname"))),
			`Reads the directory named by dirname and returns a list of directory entries sorted by filename.`, "1.0"))

	osNamespace.InternVar("mkdir", mkdir_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("name"), MakeSymbol("perm"))),
			`Creates a new directory with the specified name and permission bits.`, "1.0"))

	osNamespace.InternVar("set-env", set_env_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("key"), MakeSymbol("value"))),
			`Sets the specified key to the specified value in the environment.`, "1.0"))

	osNamespace.InternVar("sh", sh_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("name"), MakeSymbol("&"), MakeSymbol("arguments"))),
			`Executes the named program with the given arguments. Returns a map with the following keys:
      :success - whether or not the execution was successful,
      :err-msg (present iff :success if false) - string capturing error object returned by Go runtime
      :exit - exit code of program (or attempt to execute it),
      :out - string capturing stdout of the program,
      :err - string capturing stderr of the program.`, "1.0"))

	osNamespace.InternVar("sh-from", sh_from_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("dir"), MakeSymbol("name"), MakeSymbol("&"), MakeSymbol("arguments"))),
			`Executes the named program with the given arguments and working directory set to dir.
  Returns a map with the following keys:
      :success - whether or not the execution was successful,
      :err-msg (present iff :success if false) - string capturing error object returned by Go runtime
      :exit - exit code of program (or attempt to execute it),
      :out - string capturing stdout of the program,
      :err - string capturing stderr of the program.`, "1.0"))

	osNamespace.InternVar("stat", stat_,
		MakeMeta(
			NewListFrom(NewVectorFrom(MakeSymbol("filename"))),
			`Returns a map describing the named file. The info map has the following attributes:
  :name - base name of the file
  :size - length in bytes for regular files; system-dependent for others
  :mode - file mode bits
  :modtime - modification time
  :dir? - true if file is a directory`, "1.0"))

}
