package garbledreplacer

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

var ErrInvalidUTF8 = errors.New("invalid UTF-8 character")

func NewTransformer(enc encoding.Encoding, replaceRune rune) transform.Transformer {
	e := enc.NewEncoder()
	return transform.Chain(&replacer{
		replaceRune: replaceRune,
		enc:         e,
	}, e)
}

type replacer struct {
	transform.NopResetter

	enc         *encoding.Encoder
	replaceRune rune
}

var _ transform.Transformer = (*replacer)(nil)

func (t *replacer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	_src := src
	if len(_src) == 0 && atEOF {
		return
	}

	for len(_src) > 0 {
		r, n := utf8.DecodeRune(_src)
		if r == utf8.RuneError {
			// INFO: Assume only multibyte characters are split
			// 	 If there is any other pattern, it needs to be handled properly.
			err = transform.ErrShortSrc
			break
		}

		buf := _src[:n]
		if _, encErr := t.enc.Bytes(buf); encErr != nil {
			// Replace strings that cannot be converted
			buf = []byte(string(t.replaceRune))
		}
		if nDst+len(buf) > len(dst) {
			// over destination buffer
			err = transform.ErrShortDst
			break
		}
		dstN := copy(dst[nDst:], buf)
		if dstN <= 0 {
			break
		}
		nSrc += n
		nDst += dstN
		_src = _src[n:]
	}
	return
}
