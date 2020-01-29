package zipstream

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"testing"
)

func testReader(t *testing.T, s []byte) {

	var wbuf bytes.Buffer
	for j := 0; j < 2; j++ {
		z := zip.NewWriter(&wbuf)
		for i := 0; i < 2; i++ {
			zw, err := z.Create("tmp")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := zw.Write(s); err != nil {
				t.Fatal(err)
			}
		}

		if err := z.Close(); err != nil {
			t.Fatal(err)
		}
	}

	zr := NewReader(&wbuf)
	for j := 0; j < 2; j++ {
		fcount := 0
		for {
			_, err := zr.Next()
			if err != nil {
				if err != io.EOF {
					t.Fatal(err)
				}
				if fcount != 2 {
					t.Fatal("Embeded file missing", j, fcount, err)
				}
				break // No more files
			}
			fcount++
			s2, err := ioutil.ReadAll(zr)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Compare(s, s2) != 0 {
				t.Fatal("Decompressed data does not match original")
			}
		}
	}
}

func TestReader(t *testing.T) {
	testReader(t, []byte(`<poc><firstName>Juan</firstName></poc>`))

	s := new(bytes.Buffer)
	io.Copy(s, io.LimitReader(rand.New(rand.NewSource(1)), 16384))
	testReader(t, s.Bytes())
}

func TestHeaderShortRead(t *testing.T) {
	zbuf := bytes.Buffer{}
	zwtr := zip.NewWriter(&zbuf)
	zw, err := zwtr.Create("tmp.xml")
	if err != nil {
		t.Fatal(err)
	}

	//Write file 1
	if _, err := zw.Write([]byte(`<poc><firstName>Juan</firstName><lastName>RodRiehakelkjd lkbug</lastName><department>Engineering Department</department></poc>`)); err != nil {
		t.Fatal(err)
	}

	if _, err := io.Copy(zw, io.LimitReader(rand.New(rand.NewSource(1)), 16384)); err != nil {
		t.Fatal(err)
	}

	zw, err = zwtr.Create("tmp.json")
	if err != nil {
		t.Fatal(err)
	}

	//Write file 2
	if _, err := zw.Write([]byte(`{"proc":{"firstName":"Juan","lastName":"RodRiehakelkjd","department":"Engineering Department"}}`)); err != nil {
		t.Fatal(err)
	}

	if _, err := io.Copy(zw, io.LimitReader(rand.New(rand.NewSource(3)), 16384)); err != nil {
		t.Fatal(err)
	}

	//Close the zip writer
	if err := zwtr.Close(); err != nil {
		t.Fatal(err)
	}

	//Find the second magic marker so we can test the header break issue
	idx := bytes.Index(zbuf.Bytes()[4:], []byte{0x50, 0x4b, 0x03, 0x04})
	if idx < 0 {
		panic("Unable to find magic file marker")
	}

	//Get the magic marker
	//b := zbuf.Bytes()[:idx+4]
	b := zbuf.Bytes()
	b = b[:idx+6]

	//Read
	zr := NewReader(bytes.NewReader(b))
	for {
		//We are waiting for an unexepected EOF
		_, err := zr.Next()
		if err == io.ErrUnexpectedEOF {
			break
		} else if err != nil {
			t.Fatal(err.Error())
		}
	}
}
