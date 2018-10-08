package test

import (
	"testing"

	. "github.com/kocircuit/kocircuit/lang/go/eval"
	_ "github.com/kocircuit/kocircuit/lang/go/eval/macros"
	"github.com/kocircuit/kocircuit/lang/go/runtime"
)

func TestEval(t *testing.T) {
	RegisterEvalGateAt("test", "Gate", new(testNamedGate))
	RegisterEvalGateAt("test", "Underscore", new(testUnderscoreGate))
	tests := &EvalTests{T: t, Test: evalTests}
	tests.Play(runtime.NewContext())
}

var evalTests = []*EvalTest{
	{ // select, repeat
		Enabled: true,
		File: `
		G(u, w) {
			p: (a: u, b: w)
			q: (a: w, b: u)
			return: (
				s: (p.a, q.b) // []byte
				t: (p.b, q.a) // []float64
			)
		}
		Main(x, y) { return: G(u: x, w: y) }
		`,
		Arg: struct { // deconstruct/construct
			Ko_x byte    `ko:"name=x"`
			Ko_y float64 `ko:"name=y"`
		}{
			Ko_x: 7,
			Ko_y: 3.3,
		},
		Result: struct {
			Ko_s []byte    `ko:"name=s"`
			Ko_t []float64 `ko:"name=t"`
		}{
			Ko_s: []byte{7, 7},
			Ko_t: []float64{3.3, 3.3},
		},
	},
	{ // select and repeated
		Enabled: true,
		File: `
		Main(x) {
			_: (a: 1, a: 2, a: x)
			return: _.a
		}
		`,
		Arg: struct {
			Ko_x int64 `ko:"name=x"`
		}{
			Ko_x: 7,
		},
		Result: []int64{1, 2, 7},
	},
	{ // empty
		Enabled: true,
		File: `
		Empty() { return: empty() }
		empty(dontPass) { return: dontPass }
		Main() { return: Empty() }
		`,
		Arg:    struct{}{},
		Result: (*struct{})(nil),
	},
	{ // boolean gates
		Enabled: true,
		File: `
		Main() { 
			return: Xor(
				true
				And(
					Or(true, false)
					Or(Xor(true, true), false)
				)
			)
		}`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // augmentation and invocation
		Enabled: true,
		File: `
		Pass(p?) { return: p }
		Main(x) {
			callMe: Pass[x]
			return: callMe()
		}
		`,
		Arg: struct {
			Ko_x byte `ko:"name=x"`
		}{
			Ko_x: 7,
		},
		Result: byte(7),
	},
	{ // integer macros
		Enabled: true,
		File: `
		import "integer"
		Main() {
			return: And(
				integer.Equal(-1, integer.Negative(1))
				Not(integer.Less(3, 2))
				integer.Equal(
					integer.Prod(2, 3)
					integer.Sum(1, 5)
				)
				integer.Equal(integer.Moduli(7, 5), 2)
				integer.Equal(integer.Ratio(7, 2), 3)
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // len
		Enabled: true,
		File: `
		Main() {
			return: And(
				Equal(Len(), 0)
				Equal(Len(5), 1)
				Equal(Len(5, 5), 2)
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // equal
		Enabled: true,
		File: `
		Main() {
			return: And(
				Not(Equal("a", "a", "c"))
				Equal((1), 1)
				Equal(
					(a: "a", b: 1)
					(a: "a", b: (1))
				)
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // Int8, ..., Uint8, ..., Float32, ...
		Enabled: true,
		File: `
		Main() {
			return: And(
				Equal(Int8(0), Int8(256))
				Equal(Uint8(0), Uint8(256))
				Equal(Int8(255), Int8(-1))
				Equal(Float32(-3.14e7), Float32(-3.14e7))
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // yield
		Enabled: true,
		File: `
		Main() {
			return: And(
				Equal(
					Yield(if: false, else: (1))
					1
				)
				Equal(
					Len(Yield(if: true, then: (1, 2, 3)))
					3
				)
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // hash
		Enabled: true,
		File: `
		import "integer"
		NotEqual(x?) { 
			return: Not(Equal(x))
		}
		Main() {
			return: And(
				Equal((a: 1), (a: 1, b: Empty()))
				Not(Equal((a: ()), ()))
				Not(Equal(Hash(a: ()), Hash(Merge())))	
				Equal(Hash("foo"), Hash(Merge("foo")))
				Equal(Hash(), Hash(()))
				Equal(Hash(Empty()), Hash(Merge()))
				NotEqual(Hash(Int64(1)), Hash(Int32(1)))
				NotEqual(Hash(1, 2), Hash(1, 2, 3))
				Equal(Hash(a: "a", b: 1), Hash(a: "a", b: 1))
				Not(Equal(Hash(a: "a", b: ()), Hash(a: "a")))
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // range
		Enabled: true,
		File: `
		import "integer"
		Main() {
			return: And(
				// test range without stop
				Equal(
					Range(
						start: 0
						over: (1, 2, 3, 4, 5)
						with: sum(carry, elem) {
							return: (
								emit: integer.Negative(elem)
								carry: integer.Sum(carry, elem)
							)
						}
					)
					(image: (-1, -2, -3, -4, -5), residue: 15)
				)
				// test range with stop
				Equal(
					Range(
						start: 0
						over: (1, 2, 3, 4, 5)
						with: sum
						stop: stopPredicate(carry?) {
							return: Less(3, carry)
						}
					)
					(image: (-1, -2, -3), residue: 6)
				)
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // loop
		Enabled: true,
		File: `
		import "integer"
		Main() {
			return: And(
				Equal(
					Loop(
						start: 0
						step: step(carry?) { return: integer.Sum(carry, 1) }
						stop: stop(carry?) { return: Equal(carry, 33) }
					)
					33
				)
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // merge
		Enabled: true,
		File: `
		Main() {
			return: And(
				Equal(
					Merge((1, 2, 3), (11, 13, 17))
					(1, 2, 3, 11, 13, 17)
				)
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // spin, wait
		Enabled: true,
		File: `
		Main() {
			return: Spin(["Hello, world!"]).Wait()
		}
		`,
		Arg:    struct{}{},
		Result: "Hello, world!",
	},
	{ // take
		Enabled: true,
		File: `
		Main() {
			return: And(
				Equal(Take(1, 2, 3).first, 1)
				Equal(Take("a", "b").remainder, "b")
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // parallel, sequential
		Enabled: true,
		File: `
		import "integer"
		F(x?) { return: integer.Sum(x ,2) }
		G(x?) { return: integer.Prod(x ,2) }
		Main() {
			return: Equal(
				integer.Sum(
					Parallel(F[3]) // 5
					Sequential(G[5]) // 10
				)
				15
			)
		}
		`,
		Arg:    struct{}{},
		Result: true,
	},
	{ // format macro
		Enabled: true,
		File: `
		Main(x) {
			return: Format(
				format: "abc %% def %% hij %%"
				args: (x, "Y", "Z")
				withString: Return
				withArg: Return
			)
		}
		`,
		Arg: struct {
			Ko_x string `ko:"name=x"`
		}{
			Ko_x: "X",
		},
		Result: []string{"abc ", "X", " def ", "Y", " hij ", "Z", ""},
	},
	{ // template macro
		Enabled: true,
		File: `
		Main(x) {
			return: Template(
				template: "abc {{a}} def {{b.c}} hij {{b.d}}"
				args: (a: x, b: (c: "Y", d: "Z"))
				withString: Return
				withArg: Return
			)
		}
		`,
		Arg: struct {
			Ko_x string `ko:"name=x"`
		}{
			Ko_x: "X",
		},
		Result: []string{"abc ", "X", " def ", "Y", " hij ", "Z"},
	},
	{ // recover/panic macro
		Enabled: true,
		File: `
		Main(x) {
			return: Recover(
				invoke: panicOnAll[x]
				panic: Return
			)
		}
		panicOnAll(x?) {
			return: Panic(x)
		}
		`,
		Arg: struct {
			Ko_x string `ko:"name=x"`
		}{
			Ko_x: "msg",
		},
		Result: "msg",
	},
	{ // recover/panic macro through parallel execution
		Enabled: true,
		File: `
		Main(x) {
			return: Recover(
				invoke: Parallel[panicOnAll[x]]
				panic: Return
			)
		}
		panicOnAll(x?) {
			return: Panic(x)
		}
		`,
		Arg: struct {
			Ko_x string `ko:"name=x"`
		}{
			Ko_x: "msg",
		},
		Result: "msg",
	},
	{ // test named values and map extraction
		Enabled: true,
		File: `
		import "test"
		Main(x) {
			return: test.Gate(
				int64: x
				same: test.Gate(int64: 2)
				ss: (key: "k1", value: "v1")
				ss: (key: "k2", value: "v2")
			)
		}
		`,
		Arg: struct {
			Ko_x int64 `ko:"name=x"`
		}{
			Ko_x: 3,
		},
		Result: &testNamedGate{
			Int64: 3,
			Same: &testNamedGate{
				Int64: 2,
			},
			SS: map[string]string{
				"k1": "v1",
				"k2": "v2",
			},
		},
	},
	{ // test when
		Enabled: true,
		File: `
		W(x?) {
			return: When(have: x, then: Return, else: Return[5])
		}
		Main(x) {
			return: (a: W(x), b: W())
		}
		`,
		Arg: struct {
			Ko_x int64 `ko:"name=x"`
		}{
			Ko_x: 3,
		},
		Result: struct {
			Ko_a int64 `ko:"name=a"`
			Ko_b int64 `ko:"name=b"`
		}{
			Ko_a: 3,
			Ko_b: 5,
		},
	},
	{ // test all
		Enabled: true,
		File: `
		Pass(pass) { return: pass }
		HaveBoth(x, y) {
			return: When(
				have: All(x: x, y: y)
				then: Pass[pass: true]
				else: Return[false]
			)
		}
		Main(x) {
			return: And(
				HaveBoth(x: x, y: "abc")
				Not(HaveBoth(y: "abc"))
			)
		}
		`,
		Arg: struct {
			Ko_x int64 `ko:"name=x"`
		}{
			Ko_x: 3,
		},
		Result: true,
	},
	// idiom
	{ // 24: test switch idiom
		Enabled: true,
		File: `
		Main() {
			return: And(test1, test2, test3)
			test1: Equal(
				Switch(
					case: Return[]
					case: Return[1]
					case: Return[2]
				)
				1
			)
			test2: Equal(
				Switch(
					case: Return[]
					case: Return[]
					otherwise: Return[3]
				)
				3
			)
			test3: Equal(
				Switch(
					case: Return[]
					case: Return[]
				)
				Empty()
			)
		}
		`,
		Arg:    nil,
		Result: true,
	},
	{ // test sort
		Enabled: true,
		File: `
		less(left, right) {
			return: Less(left.f2, right.f2)
		}
		Main() {
			return: Equal(
				Sort(
					over: (f1: "abc", f2: 123)
					over: (f1: "def", f2: 321)
					over: (f1: "foo", f2: 213)
					less: less
				)
				((f1: "abc", f2: 123), (f1: "foo", f2: 213), (f1: "def", f2: 321))
			)
		}
		`,
		Arg:    nil,
		Result: true,
	},
	{ // test underscore arguments
		Enabled: true,
		File: `
		import "test"
		Main() {
			return: Equal(
				test.Underscore(_: 1)
				0
			)
		}
		`,
		Arg:    nil,
		Result: true,
	},
}

type testNamedGate struct {
	Int64 int64             `ko:"name=int64"`
	Same  *testNamedGate    `ko:"name=same"`
	SS    map[string]string `ko:"name=ss"`
}

func (g *testNamedGate) Play(ctx *runtime.Context) *testNamedGate {
	return g
}

type testUnderscoreGate struct {
	Underscore int64 `ko:"name=_"`
}

func (g *testUnderscoreGate) Play(ctx *runtime.Context) int64 {
	return g.Underscore
}
