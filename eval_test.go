package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func expectProgramToReturn(t *testing.T, program string, expected Value) {
	ctx := NewContext("/tmp")
	ctx.LoadBuiltins()
	val, err := ctx.Eval(strings.NewReader(program))
	if err != nil {
		t.Errorf("Did not expect program to exit with error: %s", err.Error())
	}
	if val == nil {
		t.Errorf("Return value of program should not be nil")
	} else if !val.Eq(expected) {
		t.Errorf(fmt.Sprintf("Expected and returned values don't match: %s != %s",
			strconv.Quote(expected.String()),
			strconv.Quote(val.String())))
	}
}

func TestEvalEmptyProgram(t *testing.T) {
	expectProgramToReturn(t, "", null)
	expectProgramToReturn(t, "   \n", null)
}

func TestCommentProgram(t *testing.T) {
	expectProgramToReturn(t, "// this is a comment", null)
	expectProgramToReturn(t, "// this is a comment\n", null)
}

func TestCommentInBinaryExpr(t *testing.T) {
	expectProgramToReturn(t, "1 + // this is a comment\n2", IntValue(3))
}

func TestCommentAndNewline(t *testing.T) {
	expectProgramToReturn(t, "1 + 2 // this is a comment\n", IntValue(3))
}

func TestIdentifierAfterComment(t *testing.T) {
	expectProgramToReturn(t, "x := 10 // this is a comment\nx + x", IntValue(20))
}

func TestEmptyLiteral(t *testing.T) {
	expectProgramToReturn(t, "_", empty)
}

func TestNullLiteral(t *testing.T) {
	expectProgramToReturn(t, "?", null)
}

func TestStringLiteral(t *testing.T) {
	expectProgramToReturn(t, "'Hello, World!\\n'", MakeString("Hello, World!\n"))
}

func TestQuotedStringLiteral(t *testing.T) {
	expectProgramToReturn(t, "'a\\'b'", MakeString("a'b"))
}

func TestStringLiteralOverflow(t *testing.T) {
	expectProgramToReturn(t, "'\\", MakeString(""))
	expectProgramToReturn(t, "'\\'", MakeString("'"))
	expectProgramToReturn(t, "'\\x'", MakeString("x"))
	expectProgramToReturn(t, "'\\x1'", MakeString("x1"))
	expectProgramToReturn(t, "'\\x1g'", MakeString("x1g"))
}

func TestHexStringLiteral(t *testing.T) {
	expectProgramToReturn(t, "'a\\x!'", MakeString("ax!"))
	expectProgramToReturn(t, "'a\\x1!'", MakeString("ax1!"))
	expectProgramToReturn(t, "'a\\x0a!'", MakeString("a\n!"))
	expectProgramToReturn(t, "'a\\x0A!'", MakeString("a\n!"))
	expectProgramToReturn(t, "'a\\x1z!'", MakeString("ax1z!"))

	nonAsciiStringValue := StringValue([]byte{0x98})
	expectProgramToReturn(t, "'\\x98'", &nonAsciiStringValue) // test out of ascii range
}

func TestIntegerLiteral(t *testing.T) {
	expectProgramToReturn(t, "64710", IntValue(64710))
}

func TestFloatLiteral(t *testing.T) {
	expectProgramToReturn(t, "100.0", FloatValue(100))
	expectProgramToReturn(t, "3.141592", FloatValue(3.141592))
}

func TestAtomLiteral(t *testing.T) {
	atomNames := []string{
		"_?", "if", "fn", "with", "true", "false", "_if", "not_found_404",
	}

	for _, atomName := range atomNames {
		expectProgramToReturn(t, ":"+atomName, AtomValue(atomName))
	}
}

func TestKeywordLikeAtomLiteral(t *testing.T) {
	expectProgramToReturn(t, ":if", AtomValue("if"))
	expectProgramToReturn(t, ":fn", AtomValue("fn"))
	expectProgramToReturn(t, ":with", AtomValue("with"))
	expectProgramToReturn(t, ":true", AtomValue("true"))
	expectProgramToReturn(t, ":false", AtomValue("false"))
}

func TestListLiteral(t *testing.T) {
	expectProgramToReturn(t, `[1, [2, 'three'], :four]`, MakeList(
		IntValue(1),
		MakeList(
			IntValue(2),
			MakeString("three"),
		),
		AtomValue("four"),
	))
}

func TestObjectLiteral(t *testing.T) {
	expectProgramToReturn(t, `{ a: 'ay', :be: 200, 100: {('d' + 'i'): :dee } }`, ObjectValue{
		"a":  MakeString("ay"),
		"be": IntValue(200),
		"100": ObjectValue{
			"di": AtomValue("dee"),
		},
	})
}

func TestListStringify(t *testing.T) {
	expectProgramToReturn(t, `
		[
			string(:atomValue)
			string(3)
			string(2.51)
		]
	`, MakeList(
		MakeString("atomValue"),
		MakeString("3"),
		MakeString("2.51"),
	))
}

func TestObjectStringify(t *testing.T) {
	expectProgramToReturn(t, `
		x := {
			first: {}
			second: :two
			_third: {
				_fourth: 'four'
			}
		}
		x |> string()
	`, MakeString("{_third: {_fourth: 'four'}, first: {}, second: :two}"))
}

func TestFunctionDefAndCall(t *testing.T) {
	expectProgramToReturn(t, `fn getThree() { x := 4, 3 }, getThree()`, IntValue(3))
}

func TestFunctionDefWithEmpty(t *testing.T) {
	expectProgramToReturn(t, `fn getThird(_, _, third) third, getThird(1, 2, 3)`, IntValue(3))
}

func TestFunctionCreatesScope(t *testing.T) {
	expectProgramToReturn(t, `
	x := 3
	fn defineX x := 10
	defineX()
	x
	`, IntValue(3))
}

func TestBlockCreatesScope(t *testing.T) {
	expectProgramToReturn(t, `
	x := 3
	{ x := 10 }
	x
	`, IntValue(3))
}

func TestEmptyFunctionBody(t *testing.T) {
	expectProgramToReturn(t, `
	fn do {
		a: :bee
	}
	do()
	`, ObjectValue{
		"a": AtomValue("bee"),
	})
}

func TestObjectLiteralFunctionBody(t *testing.T) {
	expectProgramToReturn(t, `
	fn do {}
	do()
	`, null)
}

func TestLocalAssignment(t *testing.T) {
	expectProgramToReturn(t, `x := 100, y := 200, x`, IntValue(100))
}

func TestChainedLocalAssignment(t *testing.T) {
	expectProgramToReturn(t, `
	a := b := 10
	a + b
	`, IntValue(20))
}

func TestChainedNonlocalAssignment(t *testing.T) {
	expectProgramToReturn(t, `
	a := b := 0
	{
		a <- b <- 20
	}
	a + b
	`, IntValue(40))
}

func TestDestructureList(t *testing.T) {
	expectProgramToReturn(t, `
	list := [1, 2, 3]
	[a] := list
	[_, _, b, c] := list
	[a, b, c]
	`, MakeList(
		IntValue(1),
		IntValue(3),
		null,
	))
}

func TestDestructureObject(t *testing.T) {
	expectProgramToReturn(t, `
	obj := {
		a: 'ay'
		b: 'bee'
		12: 'see'
	}
	{a: a} := obj
	{:b: b, 10 + 2: see} := {'whatever': dee} := obj
	[a, b, see, dee]
	`, MakeList(
		MakeString("ay"),
		MakeString("bee"),
		MakeString("see"),
		null,
	))
}

func TestDestrctureToReassignList(t *testing.T) {
	expectProgramToReturn(t, `
	v := [:aa, :bbb]
	[v, w] := v
	v
	`, AtomValue("aa"))
}

func TestDestrctureToReassignObject(t *testing.T) {
	expectProgramToReturn(t, `
	a := {a: :aa, b: :bbb}
	{a: a} := a
	a
	`, AtomValue("aa"))
}

func TestUnderscoreVarNames(t *testing.T) {
	expectProgramToReturn(t, `
	_a := 'A'
	b_ := 'B'
	c_d := 'CD'
	_a + b_ + c_d
	`, MakeString("ABCD"))
}

func TestNonlocalAssignment(t *testing.T) {
	expectProgramToReturn(t, `
	x := 100
	y := 200
	fn do {
		x <- x + 100
		y := y + 100
	}
	do()
	x + y
	`, IntValue(400))
}

func TestPushToString(t *testing.T) {
	expectProgramToReturn(t, `
	s := 'hi'
	[s << 'world', s]
	`, MakeList(
		MakeString("hiworld"),
		MakeString("hiworld"),
	))
}

func TestPushToList(t *testing.T) {
	expectProgramToReturn(t, `
	arr := [:a]
	[arr << :b, arr]
	`, MakeList(
		MakeList(AtomValue("a"), AtomValue("b")),
		MakeList(AtomValue("a"), AtomValue("b")),
	))
}

func TestPushArrowPrecedence(t *testing.T) {
	expectProgramToReturn(t, `
	arr := [2] << 1 + 3
	arr << 10 << 20
	arr << x := 100
	`, MakeList(
		IntValue(2),
		IntValue(4),
		IntValue(10),
		IntValue(20),
		IntValue(100),
	))
}

func TestUnaryExpr(t *testing.T) {
	expectProgramToReturn(t, `!true`, oakFalse)
	expectProgramToReturn(t, `!(false | true)`, oakFalse)

	expectProgramToReturn(t, `-546`, IntValue(-546))
	expectProgramToReturn(t, `-3.250`, FloatValue(-3.25))
}

func TestUnaryBindToProperty(t *testing.T) {
	expectProgramToReturn(t, `!!false`, oakFalse)
	expectProgramToReturn(t, `--3`, IntValue(3))
	expectProgramToReturn(t, `
	obj := {k: false, n: 10}
	[!obj.k, -obj.n]
	`, MakeList(
		oakTrue,
		IntValue(-10),
	))
}

func TestBasicBinaryExpr(t *testing.T) {
	expectProgramToReturn(t, `2 * 3 + 1`, IntValue(7))
	expectProgramToReturn(t, `1 + 2 * 3`, IntValue(7))
}

func TestFloatDivide(t *testing.T) {
	expectProgramToReturn(t, "10 / 4", FloatValue(2.5))
}

func TestOrderedBinaryExpr(t *testing.T) {
	expectProgramToReturn(t, `-1.5 + -3.5 - 5 / 5 * 2`, FloatValue(-7))
	expectProgramToReturn(t, `(-1.5 + -3.5 - 5) / 5 * 2`, FloatValue(-4))
}

func TestBinaryExprWithParens(t *testing.T) {
	expectProgramToReturn(t, `(1 + 2) / 3 - 1 + (10 + (20 / 5)) % 3`, FloatValue(2))
}

func TestLongBinaryExprWithPrecedence(t *testing.T) {
	expectProgramToReturn(t, `x := 1 + 2 * 3 + 4 / 2 + 10 % 4, x % 5 + x`, FloatValue(12))
}

func TestBinaryExprWithComplexTerms(t *testing.T) {
	expectProgramToReturn(t, `
	fn double(n) 2 * n
	fn decrement(n) n - 1
	double(10) + if decrement(10) { 9 -> 2, _ -> 1 } + 8
	`, IntValue(30))
}

func TestBinaryExprWithinComplexTermsWithinBinaryExpr(t *testing.T) {
	expectProgramToReturn(t, `
	fn inc(n) n + 1
	2 * inc(3 + 4)
	`, IntValue(16))
}

func TestStringCompare(t *testing.T) {
	expectProgramToReturn(t, `
	[
		// empty string
		'long string' > ''
		// length comparison
		'hi' < 'hiworld'
		// lexicographical
		'heels' < 'hi'
		// space
		'abc' > ' abc'
		// equality
		'abc' = 'abc'
	]
	`, MakeList(oakTrue, oakTrue, oakTrue, oakTrue, oakTrue))
}

func TestAndOperator(t *testing.T) {
	expectProgramToReturn(t, `
	[
		true & true
		true & false
		false & true
		false & false
		'abcd' & '    '
		'    ' & 'wxyz'
		'abcdef' & '   '
		'   ' & 'abcdef'
	]
	`, MakeList(
		oakTrue,
		oakFalse,
		oakFalse,
		oakFalse,
		MakeString("    "),
		MakeString("    "),
		MakeString("   \x00\x00\x00"),
		MakeString("   \x00\x00\x00"),
	))
}

func TestXorOperator(t *testing.T) {
	expectProgramToReturn(t, `
	[
		true ^ true
		true ^ false
		false ^ true
		false ^ false
		'ABCD' ^ '    '
		'    ' ^ 'WXYZ'
		'ABCDEF' ^ '   '
		'   ' ^ 'ABCDEF'
	]
	`, MakeList(
		oakFalse,
		oakTrue,
		oakTrue,
		oakFalse,
		MakeString("abcd"),
		MakeString("wxyz"),
		MakeString("abcDEF"),
		MakeString("abcDEF"),
	))
}

func TestOrOperator(t *testing.T) {
	expectProgramToReturn(t, `
	[
		true | true
		true | false
		false | true
		false | false
		'ABCD' | '    '
		'    ' | 'WXYZ'
		'ABCDEF' | '   '
		'   ' | 'ABCDEF'
	]
	`, MakeList(
		oakTrue,
		oakTrue,
		oakTrue,
		oakFalse,
		MakeString("abcd"),
		MakeString("wxyz"),
		MakeString("abcDEF"),
		MakeString("abcDEF"),
	))
}

func TestShortCircuitingAnd(t *testing.T) {
	expectProgramToReturn(t, `
	x := 3
	[
		false & 2
		false & { x <- 10, true }
		x
	]
	`, MakeList(oakFalse, oakFalse, IntValue(3)))
}

func TestShortCircuitingOr(t *testing.T) {
	expectProgramToReturn(t, `
	x := 3
	[
		true | 2
		true | { x <- 10, false }
		x
	]
	`, MakeList(oakTrue, oakTrue, IntValue(3)))
}

func TestEmptyIfExpr(t *testing.T) {
	expectProgramToReturn(t, `if 100 {}`, null)
}

func TestEmptyIfCondition(t *testing.T) {
	expectProgramToReturn(t, `if {
		false -> 10
		true -> 20
		_ -> 30
	}`, IntValue(20))
}

func TestBasicIfExpr(t *testing.T) {
	expectProgramToReturn(t, `if 2 * 2 {
		? -> 100
		{ a: 'b' } -> 200
		5 -> 'five'
		4 -> 'four'
	}`, MakeString("four"))
}

func TestIfExprWithMultiTarget(t *testing.T) {
	for _, i := range []int{11, 12, 13} {
		expectProgramToReturn(t, fmt.Sprintf(`if %d {
			10 -> :wrong
			11, 5 + 7, { 10 + 3 } -> :right
			_ -> :wrong2
		}`, i), AtomValue("right"))
	}
}

func TestNestedIfExpr(t *testing.T) {
	expectProgramToReturn(t, `if 3 {
		10, if true {
			true -> 10
			_ -> 3
		} -> 'hi'
		100, 3 -> 'hello'
	}`, MakeString("hello"))
}

func TestIfExprWithEmpty(t *testing.T) {
	expectProgramToReturn(t, `if 10 + 2 {
		12 -> 'twelve'
		_ -> 'wrong'
	}`, MakeString("twelve"))
}

func TestIfExprWithAssignmentCond(t *testing.T) {
	expectProgramToReturn(t, `if x := 2 + 4 {
		6 -> x * x
		_ -> x
	}`, IntValue(36))
}

func TestIfExprInFunction(t *testing.T) {
	expectProgramToReturn(t, `
	fn even?(n) if n % 2 {
		0 -> true
		_ -> false
	}
	even?(100)
	`, oakTrue)
}

func TestComplexIfExprTarget(t *testing.T) {
	expectProgramToReturn(t, `
	fn double(n) 2 * n
	fn xyz(n) if n {
		1 + 2 -> :abc
		2 * double(3) -> :xyz
		_ -> false
	}
	[xyz(3), xyz(12), xyz(24)]
	`, MakeList(
		AtomValue("abc"),
		AtomValue("xyz"),
		oakFalse,
	))
}

func TestBasicWithExpr(t *testing.T) {
	expectProgramToReturn(t, `fn add(a, b) { a + b }, with add(10) 40`, IntValue(50))
}

func TestWithExprWithCallback(t *testing.T) {
	expectProgramToReturn(t, `fn applyThrice(x, f) f(f(f(x))), with applyThrice(10) fn(n) n + 1`, IntValue(13))
}

func TestRecursiveFunction(t *testing.T) {
	expectProgramToReturn(t, `
	fn times(n, f) {
		fn sub(i) if i {
			n -> ?
			_ -> {
				f(i)
				sub(i + 1)
			}
		}
		sub(0)
	}

	counter := 0
	with times(10) fn(i) {
		counter <- counter + i * 10
	}
	counter
	`, IntValue(450))
}

func TestRecursiveFunctionOnList(t *testing.T) {
	expectProgramToReturn(t, `
	fn each(list, f) {
		fn sub(i) if i {
			len(list) -> ?
			_ -> {
				f(list.(i))
				sub(i + 1)
			}
		}
		sub(0)
	}

	sum := 0
	list := [1, 2, 3, 4, 5]
	with each(list) fn(it) {
		sum <- sum + it
	}
	sum
	`, IntValue(15))
}

func TestCurriedFunctionDef(t *testing.T) {
	expectProgramToReturn(t, `
	addThree := fn(a) fn(b) fn(c) {
		a + b + c
	}

	almost := addThree(15)(20)
	almost(8)
	`, IntValue(15+20+8))
}

// string ops

func TestStringAccess(t *testing.T) {
	expectProgramToReturn(t, `
	s := 'Hello, World!'
	[
		s.0 + s.2
		s.-2
		s.15
	]
	`, MakeList(MakeString("Hl"), null, null))
}

func TestStringAssign(t *testing.T) {
	expectProgramToReturn(t, `
	s := {
		payload: 'Magnolia'
	}
	t := s.payload
	[s.payload.3 := 'pie', t]
	`, MakeList(
		MakeString("Magpieia"),
		MakeString("Magpieia"),
	))
}

func TestStringAppendByPush(t *testing.T) {
	expectProgramToReturn(t, `
	s := {
		payload: 'Oak'
	}
	[s.payload << ' language', s.payload]
	`, MakeList(
		MakeString("Oak language"),
		MakeString("Oak language"),
	))
}

func TestStringAppendByAssign(t *testing.T) {
	expectProgramToReturn(t, `
	s := {
		payload: 'Oak'
	}
	t := s.payload
	[s.payload.(len(s.payload)) := ' language', s.payload]
	`, MakeList(
		MakeString("Oak language"),
		MakeString("Oak language"),
	))
}

// list ops

func TestListAccess(t *testing.T) {
	expectProgramToReturn(t, `
	s := [1, 2, 3, 4, 5]
	[
		s.0 + s.3
		s.-2
		s.15
	]
	`, MakeList(IntValue(5), null, null))
}

func TestListAssign(t *testing.T) {
	result := MakeList(
		IntValue(1),
		IntValue(2),
		MakeString("three"),
		IntValue(4),
	)

	expectProgramToReturn(t, `
	s := {
		numbers: [1, 2, 3, 4]
	}
	t := s.numbers
	[s.numbers.2 := 'three', t]
	`, MakeList(result, result))
}

func TestListAppendByPush(t *testing.T) {
	result := MakeList(
		IntValue(1),
		IntValue(2),
		IntValue(3),
		IntValue(4),
		IntValue(100),
	)

	expectProgramToReturn(t, `
	s := {
		numbers: [1, 2, 3, 4]
	}
	t := s.numbers
	[s.numbers << 100, t]
	`, MakeList(result, result))
}

func TestListAppendByAssign(t *testing.T) {
	result := MakeList(
		IntValue(1),
		IntValue(2),
		IntValue(3),
		IntValue(4),
		IntValue(100),
	)

	expectProgramToReturn(t, `
	s := {
		numbers: [1, 2, 3, 4]
	}
	[s.numbers.(len(s.numbers)) := 100, s.numbers]
	`, MakeList(result, result))
}

// object ops

func TestObjectAccess(t *testing.T) {
	expectProgramToReturn(t, `
	obj := {
		a: 'ay'
		b: 'bee'
		c: ['see', {
			d: 'd'
		}]
	}
	[
		obj.c.(1).:d
		obj.c.(1).(:d)
	]
	`, MakeList(
		MakeString("d"),
		MakeString("d"),
	))
}

func TestObjectAssign(t *testing.T) {
	expectProgramToReturn(t, `
	obj := {
		a: 'ay'
		b: 'bee'
		c: ['see', {
			d: 'd'
		}]
	}
	[
		obj.c.(1).:e := 'hello_e'
		obj.c.(1).(:f) := 'hello_f'
		obj.c
	]
	`, MakeList(
		ObjectValue{
			"d": MakeString("d"),
			"e": MakeString("hello_e"),
			"f": MakeString("hello_f"),
		},
		ObjectValue{
			"d": MakeString("d"),
			"e": MakeString("hello_e"),
			"f": MakeString("hello_f"),
		},
		MakeList(MakeString("see"), ObjectValue{
			"d": MakeString("d"),
			"e": MakeString("hello_e"),
			"f": MakeString("hello_f"),
		}),
	))
}

func TestObjectDelete(t *testing.T) {
	expectProgramToReturn(t, `
	obj := {
		a: 'ay'
		b: 'bee'
		c: {
			d: 'dee'
			e: 'ee'
		}
	}
	[
		obj.nonexistent := _
		obj.b := { 1, 2, _ }
		obj
		obj.c.d := _
		obj.c
	]
	`, MakeList(
		ObjectValue{
			"a": MakeString("ay"),
			"c": ObjectValue{
				"e": MakeString("ee"),
			},
		},
		ObjectValue{
			"a": MakeString("ay"),
			"c": ObjectValue{
				"e": MakeString("ee"),
			},
		},
		ObjectValue{
			"a": MakeString("ay"),
			"c": ObjectValue{
				"e": MakeString("ee"),
			},
		},
		ObjectValue{
			"e": MakeString("ee"),
		},
		ObjectValue{
			"e": MakeString("ee"),
		},
	))
}

func TestSinglePipe(t *testing.T) {
	expectProgramToReturn(t, `
	fn append(a, b) a + b
	'hello' |> append('world')
	`, MakeString("helloworld"))
}

func TestMultiPipe(t *testing.T) {
	expectProgramToReturn(t, `
	fn append(a, b) a + b
	'hello' |> append('world') |> append('!')
	`, MakeString("helloworld!"))
}

func TestComplexPipe(t *testing.T) {
	expectProgramToReturn(t, `
	lib := {
		add1: fn(n) n + 1
		double: fn(n) 2 * n
	}
	fn getAdder(env) { env.add1 }
	100 |> lib.add1() |> lib.double() |> getAdder(lib)()
	`, IntValue(203))
}

func TestPipeWithExpr(t *testing.T) {
	expectProgramToReturn(t, `
	fn add(a, b) a + b
	fn double(n) 2 * n
	fn apply(x, f) f(x)

	10 |> add(20) |> with apply() fn(n) n |> double() + 40
	`, IntValue(100))
}

func TestExtraArgs(t *testing.T) {
	expectProgramToReturn(t, `
	fn getExtra(a, b, c) {
		[b, c]
	}
	getExtra(1, ?)
	`, MakeList(null, null))
}

func TestRestArgs(t *testing.T) {
	expectProgramToReturn(t, `
	fn getRest(first, rest...) {
		rest
	}
	getRest(1, 2, 3, 4, 5)
	`, MakeList(
		IntValue(2),
		IntValue(3),
		IntValue(4),
		IntValue(5),
	))
}
