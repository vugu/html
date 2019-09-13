package html

import (
	"bytes"
	"io"
	"testing"
)

// TODO: figure out where this goes in the test suite

func TestTokenizerPreserveCase(t *testing.T) {

	inHTML := `<!doctype html>
<html>
<body>
	<Div id="test1" Class="something"></Div> <!-- tag that matches an Atom -->
	<some-Other id="test2" othER-attr="blah"></some-Other> <!-- random tag not an Atom -->
	Some other random text here.
</body>
</html>`
	inHTMLB := []byte(inHTML)
	// defer func() {
	// 	t.Logf("inHTMLB after: %s", inHTMLB)
	// }()

	z := NewTokenizer(bytes.NewReader(inHTMLB))
	z.PreserveCase(true)

loop:
	for {
		tt := z.Next()
		switch tt {
		case ErrorToken:
			if z.Err() == io.EOF {
				break loop
			}
			t.Error(z.Err())
			t.FailNow()
		case TextToken:
			t.Logf("TextToken: %s", z.Text())
		case StartTagToken:
			tn, ha := z.TagName()
			t.Logf("StartTagToken, tag=%s, hasAttr=%v", tn, ha)
			tns := string(tn)

			if tns == "some-other" || tns == "div" {
				t.Logf("Should not have gotten lower case %q as element name", tns)
				t.Fail()
			}

			var k, v []byte
			for ha {
				k, v, ha = z.TagAttr()
				t.Logf(" attr: %s = %q", k, v)
				if string(k) == "class" {
					t.Logf("Should not have gotten lower case %q as attribute name", string(k))
					t.Fail()
				}
			}

		case EndTagToken:
			tn, _ := z.TagName()
			t.Logf("EndTagToken, tag=%s", tn)
		default:
			t.Logf("Other Token: %v", tt)
		}
	}

}

func TestParserPreserveCase(t *testing.T) {

	inHTML := `<!doctype html>
<html>
<body>
	<Div id="test1" Class="something"></Div> <!-- tag that matches an Atom -->
	<some-Other id="test2" othER-attr="blah"></some-Other> <!-- random tag not an Atom -->
	Some other random text here.
</body>
</html>`
	inHTMLB := []byte(inHTML)

	node, err := ParseWithOptions(bytes.NewReader(inHTMLB), ParseOptionPreserveCase(true))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var visit func(n *Node)
	visit = func(n *Node) {
		t.Logf("Node Type=%v, Data=%q, DataAtom=%v", n.Type, n.Data, n.DataAtom)
		for _, a := range n.Attr {
			t.Logf("  %s=%q", a.Key, a.Val)
		}
		if n.FirstChild != nil {
			visit(n.FirstChild)
		}
		if n.NextSibling != nil {
			visit(n.NextSibling)
		}
	}
	visit(node)

}
