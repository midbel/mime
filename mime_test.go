package mime

import (
  "testing"
)

func TestParse(t *testing.T) {
  mimetypes := []struct{
    Mime string
    Valid bool
  } {
    {Mime: "", Valid: false},
    {Mime: "extralongmimetypewithmorethansixtyfourcharactersformaintype", Valid: false},
    {Mime: "application", Valid: false},
    {Mime: "application/", Valid: false},
    {Mime: "application/cbor+", Valid: false},
    {Mime: "application/cbor;", Valid: false},
    {Mime: "?application/cbor", Valid: false},
    {Mime: "application/?cbor", Valid: false},
    {Mime: "application/cbor", Valid: true},
    {Mime: "application/cbor", Valid: true},
    {Mime: "application/cbor+rfc;type=basic", Valid: true},
    {Mime: "text/xml", Valid: true},
    {Mime: "text/xml;charset=UTF-8", Valid: true},
    {Mime: "text/xml;charset=\"UTF-8\"", Valid: true},
  }
  for i, mt := range mimetypes {
    _, err := Parse(mt.Mime)
    if mt.Valid && err != nil {
      t.Errorf("%d) fail parsing %s: %s", i+1, mt.Mime, err)
      continue
    }
    if !mt.Valid && err == nil {
      t.Errorf("%d) invalid mime %q parsed succesfully", i+1, mt.Mime)
    }
  }
}
