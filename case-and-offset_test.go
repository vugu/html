package html

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"testing"

	atom "github.com/vugu/html/atom"
)

// TODO: figure out where this goes in the test suite

func TestTokenizerOffset(t *testing.T) {

	var buf bytes.Buffer
	linelen := 34 // includes newline
	fmt.Fprintf(&buf, "<html>                           \n")
	fmt.Fprintf(&buf, "<body>                           \n")
	for i := 0; i < 20000; i++ { // make this high enough to force reallocation and reading in blocks
		fmt.Fprintf(&buf, "<div id=\"%08d\">%08d</div>\n", i, i)
	}
	fmt.Fprintf(&buf, "</body>                          \n")
	fmt.Fprintf(&buf, "</html>                          \n")
	fmt.Fprintf(&buf, "                                 \n")
	fmt.Fprintf(&buf, "                                 \n")

	// log.Printf("DATA:\n%s", buf.Bytes())

	divnum := 0

	z := NewTokenizer(bytes.NewReader(buf.Bytes()))
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

			zText := z.Text()
			if zText[0] == '0' {

				vi := -1
				fmt.Sscanf(string(zText), "%d", &vi)

				zoff := z.Offset()
				if (float64(zoff-19)/float64(linelen))-2 != float64(vi) {
					t.Logf("BAD TEXT OFFSET: zoff = %d, vi = %d", zoff, vi)
					t.Fail()
				}

			}

		case StartTagToken:
			tn, _ := z.TagName()

			if bytes.Compare(tn, []byte("div")) == 0 {

				k, v, _ := z.TagAttr()
				if bytes.Compare(k, []byte("id")) != 0 {
					t.Errorf("unknown k: %s", k)
				}

				vi := -1
				fmt.Sscanf(string(v), "%d", &vi)

				zoff := z.Offset()
				if (float64(zoff)/float64(linelen))-2 != float64(vi) {
					t.Logf("BAD DIV OFFSET: zoff = %d, vi = %d", zoff, vi)
					t.Fail()
				}

				divnum++
			}

		case EndTagToken:

			tn, _ := z.TagName()
			if bytes.Compare(tn, []byte("div")) == 0 {
				zoff := z.Offset()
				if (float64(zoff-27)/float64(linelen))-2 != float64(divnum-1) {
					b := buf.Bytes()[zoff : zoff+24]
					t.Logf("BAD DIV CLOSE OFFSET: zoff = %d, divnum = %d (bytes at offset: %q)", zoff, divnum, b)
					t.Fail()
				}
			}

		}
	}

}

func TestParserOffset(t *testing.T) {

	var buf bytes.Buffer
	// linelen := 34 // includes newline
	fmt.Fprintf(&buf, "<html>                           \n")
	fmt.Fprintf(&buf, "<body>                           \n")
	for i := 0; i < 20; i++ { // make this high enough to force reallocation and reading in blocks
		fmt.Fprintf(&buf, "<div id=\"%08d\">%08d</div>\n", i, i)
	}
	fmt.Fprintf(&buf, "</body>                          \n")
	fmt.Fprintf(&buf, "</html>                          \n")
	fmt.Fprintf(&buf, "                                 \n")
	fmt.Fprintf(&buf, "                                 \n")

	n, err := Parse(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var visit func(n *Node)
	visit = func(n *Node) {

		// t.Logf("Node Type=%v, Data=%q, DataAtom=%v", n.Type, n.Data, n.DataAtom)
		// for _, a := range n.Attr {
		// 	t.Logf("  %s=%q", a.Key, a.Val)
		// }

		if n.DataAtom == atom.Div {
			var idv int
			for _, a := range n.Attr {
				if a.Key == "id" {
					idv, _ = strconv.Atoi(a.Val)
					break

				}
			}
			t.Logf("n.Offset = %d, idv = %d", n.Offset, idv)
			// n.Offset
		}

		if n.FirstChild != nil {
			visit(n.FirstChild)
		}
		if n.NextSibling != nil {
			visit(n.NextSibling)
		}
	}
	visit(n)

}

// func TestTokenizerPreserveCase(t *testing.T) {

// 	inHTML := `<!doctype html>
// <html>
// <body>
// 	<Div id="test1" Class="something"></Div> <!-- tag that matches an Atom -->
// 	<some-Other id="test2" othER-attr="blah"></some-Other> <!-- random tag not an Atom -->
// 	Some other random text here.
// </body>
// </html>`
// 	inHTMLB := []byte(inHTML)
// 	// defer func() {
// 	// 	t.Logf("inHTMLB after: %s", inHTMLB)
// 	// }()

// 	z := NewTokenizer(bytes.NewReader(inHTMLB))
// 	// z.PreserveCase(true)

// loop:
// 	for {
// 		tt := z.Next()
// 		t.Logf("Offset: %d", z.Offset())
// 		switch tt {
// 		case ErrorToken:
// 			if z.Err() == io.EOF {
// 				break loop
// 			}
// 			t.Error(z.Err())
// 			t.FailNow()
// 		case TextToken:
// 			t.Logf("TextToken: %s", z.Text())
// 		case StartTagToken:
// 			tn, tno, ha := z.TagNameAndOrig()
// 			t.Logf("StartTagToken, tag=%s origTagName=%s, hasAttr=%v", tn, tno, ha)
// 			tns := string(tn)

// 			if tns == "some-other" || tns == "div" {
// 				t.Logf("Should not have gotten lower case %q as element name", tns)
// 				t.Fail()
// 			}

// 			var k, ko, v []byte
// 			for ha {
// 				k, ko, v, ha = z.TagAttrAndOrig()
// 				t.Logf(" attr: %s (orig=%s) = %q", k, ko, v)
// 				if string(k) == "class" {
// 					t.Logf("Should not have gotten lower case %q as attribute name", string(k))
// 					t.Fail()
// 				}
// 			}

// 		case EndTagToken:
// 			tn, tno, _ := z.TagNameAndOrig()
// 			t.Logf("EndTagToken, tag=%s (origTagName=%s)", tn, tno)
// 		default:
// 			t.Logf("Other Token: %v", tt)
// 		}
// 	}

// }

func TestParserPreserveCase(t *testing.T) {

	if s := atom.String([]byte("Div")); s != "Div" {
		t.Logf("atom.String() returned %q instead of Div", s)
		t.Fail()
	}

	inHTML := `<!doctype html>
<html>
<body>
	<Div id="test1" Class="something"></Div> <!-- tag that matches an Atom -->
	<some-Other id="test2" othER-attr="blah"></some-Other> <!-- random tag not an Atom -->
	Some other random text here.
</body>
</html>`
	inHTMLB := []byte(inHTML)

	node, err := ParseWithOptions(bytes.NewReader(inHTMLB))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	checked := 0

	var visit func(n *Node)
	visit = func(n *Node) {

		if n.DataAtom == atom.Div {
			if n.OrigData != "Div" {
				t.Logf("Expected Div got: %q", n.OrigData)
				t.Fail()
			}
			checked++
			for _, a := range n.Attr {
				if a.Key == "class" {
					if a.OrigKey != "Class" {
						t.Logf("Expected Class got: %q", a.OrigKey)
						t.Fail()
					}
					checked++
				}
			}
		}

		if n.Data == "some-other" {
			if n.OrigData != "some-Other" {
				t.Logf("Expected some-Other got: %q", n.OrigData)
				t.Fail()
			}
			checked++
			for _, a := range n.Attr {
				if a.Key == "other-attr" {
					if a.OrigKey != "othER-attr" {
						t.Logf("Expected othER-attr got: %q", a.OrigKey)
						t.Fail()
					}
					checked++
				}
			}
		}

		if n.FirstChild != nil {
			visit(n.FirstChild)
		}
		if n.NextSibling != nil {
			visit(n.NextSibling)
		}
	}
	visit(node)

	if checked != 4 {
		t.Errorf("expected 4 checks but did %v", checked)
		t.Fail()
	}

}

// func TestParserPreserveCase(t *testing.T) {

// 	inHTML := `<!doctype html>
// <html>
// <body>
// 	<Div id="test1" Class="something"></Div> <!-- tag that matches an Atom -->
// 	<some-Other id="test2" othER-attr="blah"></some-Other> <!-- random tag not an Atom -->
// 	Some other random text here.
// </body>
// </html>`
// 	inHTMLB := []byte(inHTML)

// 	// node, err := ParseWithOptions(bytes.NewReader(inHTMLB), ParseOptionPreserveCase(true))
// 	node, err := ParseWithOptions(bytes.NewReader(inHTMLB))
// 	if err != nil {
// 		t.Error(err)
// 		t.FailNow()
// 	}

// 	var visit func(n *Node)
// 	visit = func(n *Node) {
// 		t.Logf("Node Type=%v, Data=%q, DataAtom=%v", n.Type, n.Data, n.DataAtom)
// 		for _, a := range n.Attr {
// 			t.Logf("  %s=%q", a.Key, a.Val)
// 		}
// 		if n.FirstChild != nil {
// 			visit(n.FirstChild)
// 		}
// 		if n.NextSibling != nil {
// 			visit(n.NextSibling)
// 		}
// 	}
// 	visit(node)

// }
