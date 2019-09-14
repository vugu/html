package html

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestLineCounterText(t *testing.T) {

	txt := "this\nis \r\na   \ntest\n"
	txtb := []byte(txt)

	var buf bytes.Buffer
	for i := 0; i < 10000; i++ {
		buf.Write(txtb)
	}

	lc := NewLineCounter(bytes.NewReader(buf.Bytes()))

	ioutil.ReadAll(lc)

	check := func(in, outLine, outOff int) {
		a, b := lc.ForOffset(in)
		if outLine != a {
			t.Errorf("input %d expected line num %d but got %d", in, outLine, a)
		}
		if outOff != b {
			t.Errorf("input %d expected line offset %d but got %d", in, outOff, b)
		}
	}

	check(0, 1, 0)
	check(1, 1, 0)
	check(4, 1, 0)
	check(5, 2, 5)
	check(6, 2, 5)
	check(9, 2, 5)
	check(10, 3, 10)
	check(11, 3, 10)
	check(15, 4, 15)
	check(19, 4, 15)
	check(20, 5, 20)
	check(21, 5, 20)
	check(20000000, 40001, 200000)

}

func TestLineCounterHTML(t *testing.T) {

	// simple test case against HTML, let's see if it works when we put all this together

	h := `<!doctype html>
<html>
	<body>
		<div id="testing">blah</div>
	</body>
</html>`

	hb := []byte(h)

	lc := NewLineCounter(bytes.NewReader(hb))
	node, err := Parse(lc)
	if err != nil {
		t.Fatal(err)
	}

	blahoff := -1
	cc := 0
	var visit func(n *Node)
	visit = func(n *Node) {

		if n.Type == TextNode && n.Data == "blah" {
			cc++
			blahoff = n.Offset
		}

		if n.FirstChild != nil {
			visit(n.FirstChild)
		}
		if n.NextSibling != nil {
			visit(n.NextSibling)
		}
	}
	visit(node)

	line, lineOff := lc.ForOffset(blahoff)
	if !(blahoff == 51 && line == 4 && lineOff == 31) {
		t.Errorf("blahoff=%d, line=%d, lineOff=%d", blahoff, line, lineOff)
	}
}
