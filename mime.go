package mime

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

const sixtyFour = 64

var (
	ErrEmpty     = errors.New("empty")
	ErrChar      = errors.New("invalid character")
	ErrDuplicate = errors.New("duplicate")
)

var Unknown = Mime{
	Main: "unknown",
	Sub:  "unknown",
}

type Mime struct {
	Main   string
	Sub    string
	Suffix string
	Params map[string]string
}

func (m Mime) String() string {
	var str strings.Builder

	str.WriteString(m.Main)
	str.WriteRune(slash)
	str.WriteString(m.Sub)

	if m.Suffix != "" {
		str.WriteRune(plus)
		str.WriteString(m.Suffix)
	}

	for k, v := range m.Params {
		str.WriteRune(semicolon)
		str.WriteString(k)
		str.WriteRune(equal)
		str.WriteRune(quote)
		str.WriteString(v)
		str.WriteRune(quote)
	}

	return str.String()
}

const (
	dot        = '.'
	slash      = '/'
	plus       = '+'
	minus      = '-'
	semicolon  = ';'
	quote      = '"'
	equal      = '='
	space      = ' '
	tab        = '\t'
	caret      = '^'
	underscore = '_'
	dollar     = '$'
	ampersand  = '&'
	bang       = '!'
	pound      = '#'
)

func Parse(str string) (Mime, error) {
	var (
		mt    Mime
		err   error
		delim byte
	)
	if len(str) == 0 {
		return Unknown, ErrEmpty
	}
	rs := strings.NewReader(str)
	_, mt.Main, err = parseName(rs, func(b byte) bool {
		return b == slash
	})
	if err != nil {
		return Unknown, err
	}

	delim, mt.Sub, err = parseName(rs, func(b byte) bool {
		return b == plus || b == semicolon
	})
	if err != nil && !errors.Is(err, io.EOF) {
		return Unknown, err
	}

	mt.Params = make(map[string]string)
	if delim == plus {
		delim, mt.Suffix, err = parseName(rs, func(b byte) bool {
			return b == semicolon
		})
		if err != nil && !errors.Is(err, io.EOF) {
			return Unknown, err
		}
	}
	if delim == semicolon {
		if rs.Len() == 0 {
			return Unknown, fmt.Errorf("no parameter given after semicolon")
		}
		for rs.Len() > 0 {
			skipBlank(rs)
			k, v, err := parseKeyValue(rs)
			if err != nil {
				return Unknown, err
			}
			if _, ok := mt.Params[k]; ok {
				return Unknown, fmt.Errorf("%w: %s", ErrDuplicate, k)
			}
			mt.Params[k] = v
		}
	}

	return mt, err
}

func parseName(rs *strings.Reader, isTerm func(byte) bool) (byte, string, error) {
	b, err := rs.ReadByte()
	if err != nil {
		return 0, "", ErrEmpty
	}
	if !isAlpha(b) {
		return 0, "", fmt.Errorf("%w: not an alphanumeric character", ErrChar)
	}

	var str strings.Builder
	str.WriteByte(b)
	for i := 1; i < sixtyFour; i++ {
		b, err := rs.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return 0, strings.ToLower(str.String()), nil
			}
			return 0, "", fmt.Errorf("%w: %s after %s", ErrChar, err, str.String())
		}
		if isTerm(b) {
			return b, strings.ToLower(str.String()), nil
		}
		if !isValid(b) {
			return 0, "", fmt.Errorf("%w: %s (%c)", ErrChar, str.String(), b)
		}
		str.WriteByte(b)
	}
	return 0, "", fmt.Errorf("%w: termination character not found", ErrChar)
}

func parseKeyValue(rs *strings.Reader) (string, string, error) {
	_, key, err := parseName(rs, func(b byte) bool {
		return b == equal
	})
	if err != nil {
		return "", "", err
	}
	delim, err := rs.ReadByte()
	if err != nil {
		return "", "", err
	}
	if delim != quote {
		rs.UnreadByte()
		delim = semicolon
	}
	delim, val, err := parseName(rs, func(b byte) bool {
		return b == delim
	})
	return key, val, nil
}

func skipBlank(rs *strings.Reader) {
	for {
		if b, _ := rs.ReadByte(); !isBlank(b) {
			break
		}
	}
	rs.UnreadByte()
}

func isValid(b byte) bool {
	return isAlpha(b) || isPunct(b)
}

func isPunct(b byte) bool {
	return b == caret || b == underscore || b == dollar || b == ampersand ||
		b == bang || b == pound || b == dot || b == minus
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlpha(b byte) bool {
	return isLetter(b) || isDigit(b)
}

func isBlank(b byte) bool {
	return b == space || b == tab
}
