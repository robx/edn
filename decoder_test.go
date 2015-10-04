// Copyright 2015 Jean Niklas L'orange.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edn

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
)

// your basic unit tests.. unfinished, probably.

func TestIntReading(t *testing.T) {
	intStrs := [...]string{"0", "1", "+100", "-982", "8223372036854775808", "-5N", "-0N"}
	ints := [...]int64{0, 1, 100, -982, 8223372036854775808, -5, 0}
	for i, istr := range intStrs {
		var n int64
		err := UnmarshalString(istr, &n)
		if err != nil {
			t.Errorf("int64 '%s' failed, but expected success", istr)
		} else if n != ints[i] {
			t.Errorf("int64 '%s' was decoded to %d, but expected %d", istr, n, ints[i])
		}
	}
}

func TestBigIntReading(t *testing.T) {
	const huge = "32317006071311007300714876688669951960444102669715484032130345427524655138867890893197201411522913463688717960921898019494119559150490921095088152386448283120630877367300996091750197750389652106796057638384067568276792218642619756161838094338476170470581645852036305042887575891541065808607552399123930385521914333389668342420684974786564569494856176035326322058077805659331026192708460314150258592864177116725943603718461857357598351152301645904403697613233287231227125684710820209725157101726931323469678542580656697935045997268352998638215525166389647960126939249806625440700685819469589938384356951833568218188663"

	bigIntStrs := [...]string{"0", "1", "-1N", "0N", huge + "N"}

	_1 := func(v *big.Int, _ bool) *big.Int { return v }
	bigInts := [...]*big.Int{
		big.NewInt(0), big.NewInt(1), big.NewInt(-1),
		big.NewInt(0), _1(big.NewInt(0).SetString(huge, 10)),
	}
	for i, istr := range bigIntStrs {
		var n *big.Int
		err := UnmarshalString(istr, &n)
		if err != nil {
			t.Errorf("*big.Int '%s' failed, but expected success", istr)
		} else if n.Cmp(bigInts[i]) != 0 {
			t.Errorf("*big.Int '%s' was decoded to %s, but expected %s", istr, n, bigInts[i])
		}
	}
}

func TestBigFloat(t *testing.T) {
	const huge = "123456789123456789123456789123456789123456789123456789.123456789"

	bigFloatStrs := [...]string{"0", "1M", "-0.1M", "1.1e-10M", huge + "M"}

	bigFloat := func(s string) *big.Float {
		f, _, err := big.ParseFloat(s, 10, 192, big.ToNearestEven)
		if err != nil {
			t.Fatal(err)
		}
		return f
	}

	bigFloats := [...]*big.Float{
		bigFloat("0"), bigFloat("1"), bigFloat("-0.1"),
		bigFloat("1.1e-10"), bigFloat(huge),
	}
	for i, istr := range bigFloatStrs {
		var n *big.Float
		err := UnmarshalString(istr, &n)
		if err != nil {
			t.Errorf("*big.Float '%s' failed, but expected success", istr)
			t.Error(err)
		} else if n.Cmp(bigFloats[i]) != 0 {
			t.Errorf("*big.Float '%s' was decoded to %s, but expected %s", istr, n, bigFloats[i])
		}
	}
}

func TestArray(t *testing.T) {
	stringArray := `("foo" "bar" "baz")`
	stringExpected := [...]string{"foo", "bar", "baz"}
	var sa [3]string
	err := UnmarshalString(stringArray, &sa)
	if err != nil {
		t.Error(`expected '("foo" "bar" "baz")' to decode fine, but didn't`)
	} else {
		for i, expected := range stringExpected {
			if expected != sa[i] {
				t.Errorf(`Element %d in '("foo" "bar" "baz")' (%q) was encoded to %q`,
					i, expected, sa[i])
			}
		}
	}
}

func TestStruct(t *testing.T) {
	type Animal struct {
		Name string
		Type string `edn:"kind"`
	}
	type Person struct {
		Name      string
		Birthyear int `edn:"born"`
		Pets      []Animal
	}
	hans := `{:name "Hans",
            :born 1970,
            :pets [{:name "Cap'n Jack" :kind "Sparrow"}
                   {:name "Freddy" :kind "Cockatiel"}]}`
	goHans := Person{"Hans", 1970,
		[]Animal{{"Cap'n Jack", "Sparrow"}, {"Freddy", "Cockatiel"}}}
	var ednHans Person
	err := UnmarshalString(hans, &ednHans)
	if err != nil {
		t.Error("Error when decoding Hans")
	} else if !reflect.DeepEqual(goHans, ednHans) {
		t.Error("EDN Hans is not equal to Go hans")
	}
}

func TestRec(t *testing.T) {
	type Node struct {
		Left  *Node
		Val   int
		Right *Node
	}
	// here we're using symbols
	tree := "{left {left {val 3} val 5 right nil} val 10 right {val 15 right {val 17}}}"
	goTree := Node{Left: &Node{Left: &Node{Val: 3}, Val: 5, Right: nil},
		Val: 10, Right: &Node{Val: 15, Right: &Node{Val: 17}}}
	var ednTree Node
	err := UnmarshalString(tree, &ednTree)
	if err != nil {
		t.Errorf("Couldn't unmarshal tree: %s", err.Error())
	} else if !reflect.DeepEqual(goTree, ednTree) {
		t.Error("Mismatch between the Go tree and the tree encoded as EDN")
	}
}

func TestDiscard(t *testing.T) {
	var s Symbol
	discarding := "#_ #zap #_ xyz foo bar"
	expected := Symbol("bar")
	err := UnmarshalString(discarding, &s)
	if err != nil {
		t.Errorf("Expected '#_ #zap #_ xyz foo bar' to successfully read")
		t.Log(err.Error())
	} else if expected != s {
		t.Error("Mismatch between the Go symbol and the symbol encoded as EDN")
	}

	discarding = "#_ #foo #foo #foo #_#_bar baz zip quux"
	expected = Symbol("quux")
	err = UnmarshalString(discarding, &s)
	if err != nil {
		t.Errorf("Expected '#_ #foo #foo #foo #_#_bar baz zip quux' to successfully read")
		t.Log(err.Error())
	} else if expected != s {
		t.Error("Mismatch between the Go symbol and the symbol encoded as EDN")
	}
}

// Test that we can read self-defined unmarshalEDNs
type testUnmarshalEDN string

func (t *testUnmarshalEDN) UnmarshalEDN(bs []byte) (err error) {
	var kw Keyword
	err = Unmarshal(bs, &kw)
	if err == nil && string(kw) != "all" {
		return fmt.Errorf("testUnmarshalEDN must be :all if it's a keyword, not %s", kw)
	}
	if err == nil {
		*t = testUnmarshalEDN(kw)
		return
	}
	// try to parse set of keywords
	var m map[Keyword]bool
	err = Unmarshal(bs, &m)
	if err == nil {
		*t = "set elements"
	}
	return
}

func TestUnmarshalEDN(t *testing.T) {
	var tm testUnmarshalEDN
	data := ":all"
	expected := testUnmarshalEDN("all")
	err := UnmarshalString(data, &tm)
	if err != nil {
		t.Errorf("Expected ':all' to successfully read into testUnmarshalEDN")
		t.Log(err.Error())
	} else if expected != tm {
		t.Error("Mismatch between testUnmarshalEDN unmarshaling and the expected value")
		t.Logf("Was %s, expected %s", tm, expected)
	}

	data = "#{:foo :bar :baz}"
	expected = testUnmarshalEDN("set elements")
	err = UnmarshalString(data, &tm)
	if err != nil {
		t.Errorf("Expected '#{:foo :bar :baz}' to successfully read into testUnmarshalEDN")
		t.Log(err.Error())
	} else if expected != tm {
		t.Error("Mismatch between testUnmarshalEDN unmarshaling and the expected value")
		t.Logf("Was %s, expected %s", tm, expected)
	}

	data = "#{:all #{:foo :bar :baz}}"

	var tms map[testUnmarshalEDN]bool
	err = UnmarshalString(data, &tms)
	if err != nil {
		t.Errorf("Expected '#{:all #{:foo :bar :baz}}' to successfully read into a map[testUnmarshalEDN]bool")
		t.Log(err.Error())
	} else {
		fail := false
		if len(tms) != 2 {
			fail = true
		}
		if !tms[testUnmarshalEDN("all")] {
			fail = true
		}
		if !tms[testUnmarshalEDN("set elements")] {
			fail = true
		}
		if fail {
			t.Error("Mismatch between testUnmarshalEDN unmarshaling and the expected value")
			t.Logf("Was %s", tm)
		}
	}
}

type vectorCounter int

func (v *vectorCounter) UnmarshalEDN(bs []byte) (err error) {
	var vec []interface{}
	err = Unmarshal(bs, &vec)
	if err != nil {
		return
	}
	*v = vectorCounter(len(vec))
	return
}

func TestVectorCounter(t *testing.T) {
	var v vectorCounter
	data := "[foo bar baz]"

	var expected vectorCounter = 3
	err := UnmarshalString(data, &v)
	if err != nil {
		t.Errorf("Expected '%s' to successfully read into vectorCounter", data)
		t.Log(err.Error())
	} else if expected != v {
		t.Error("Mismatch between vectorCounter unmarshaling and the expected value")
		t.Logf("Was %d, expected %d", v, expected)
	}

	data = "(a b c d e f)"
	expected = 6
	err = UnmarshalString(data, &v)
	if err != nil {
		t.Errorf("Expected '%s' to successfully read into vectorCounter", data)
		t.Log(err.Error())
	} else if expected != v {
		t.Error("Mismatch between vectorCounter unmarshaling and the expected value")
		t.Logf("Was %d, expected %d", v, expected)
	}

	data = `[[a b c][d e f g h],[#_3 z 2 \c]()["c d e"](2 3.0M)]`
	var vs []vectorCounter
	expected2 := []vectorCounter{3, 5, 3, 0, 1, 2}
	err = UnmarshalString(data, &vs)
	if err != nil {
		t.Errorf("Expected '%s' to successfully read into []vectorCounter", data)
		t.Log(err.Error())
	} else if !reflect.DeepEqual(vs, expected2) {
		t.Errorf("Mismatch between %#v and %#v", vs, expected2)
	}

	data = `{[a b c] "quux", [] "frob"}`
	var vmap map[vectorCounter]string
	expected3 := map[vectorCounter]string{3: "quux", 0: "frob"}
	err = UnmarshalString(data, &vmap)
	if err != nil {
		t.Errorf("Expected '%s' to successfully read into map[vectorCounter]string", data)
		t.Log(err.Error())
	} else if !reflect.DeepEqual(vmap, expected3) {
		t.Errorf("Mismatch between %#v and %#v", vmap, expected3)
	}
}

type mapCounter int

func (mc *mapCounter) UnmarshalEDN(bs []byte) (err error) {
	var m map[interface{}]interface{}
	err = Unmarshal(bs, &m)
	if err != nil {
		return
	}
	*mc = mapCounter(len(m))
	return
}

func TestMapCounter(t *testing.T) {
	var mc mapCounter
	data := `{:a :b :c :d 1 0 nil foo}`

	var expected mapCounter = 4
	err := UnmarshalString(data, &mc)
	if err != nil {
		t.Errorf("Expected '%s' to successfully read into mapCounter", data)
		t.Log(err.Error())
	} else if expected != mc {
		t.Error("Mismatch between mapCounter unmarshaling and the expected value")
		t.Logf("Was %d, expected %d", mc, expected)
	}
}
