package lexer

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
)

type regexHandler func(lex *lexer, regex *regexp.Regexp)

type buffer struct {
	reader io.Reader
	cache  []byte
	eof    bool
	pos    int
	// TODO threading
}

func (b *buffer) debug() {
	fmt.Printf("Pos %d\nConsumed: %s\nRemaining:%s\n", b.pos, b.cache[:b.pos], b.cache[b.pos:])
}

type fork struct {
	buffer *buffer
	pos    int
}

func newBuffer(reader io.Reader) *buffer {
	return &buffer{reader: reader}
}

func (b *buffer) fork() *fork {
	return &fork{b, b.pos}
}

func (b *buffer) advanceCache(count int) error {
	tmp := make([]byte, count)
	read, err := b.reader.Read(tmp)
	b.cache = append(b.cache, tmp[:read]...)
	// b.debug()
	if err == io.EOF {
		b.eof = true
		err = nil
	}
	return err
}

func (b *buffer) advance(length int) {
	if len(b.cache) < (length + b.pos) {
		panic("Cannot advance beyone what has been read")
	}
	b.pos += length
}

func (b *buffer) subString(start int, end int) string {
	return (string)(b.cache[b.pos+start : b.pos+end])
}

func (f *fork) Read(b []byte) (count int, err error) {
	if f.buffer.eof && f.pos >= len(f.buffer.cache) {
		return 0, io.EOF
	}
	cacheSize := len(f.buffer.cache)
	desiredCacheSize := len(b) // TODO: This is wrong
	if desiredCacheSize > cacheSize {
		err = f.buffer.advanceCache(desiredCacheSize - cacheSize)
	}
	count = copy(b, f.buffer.cache[f.pos:])
	f.pos += count
	if f.buffer.eof && err == nil {
		err = io.EOF
	}
	return
}

type regexPattern struct {
	regex   *regexp.Regexp
	handler regexHandler
}

type lexer struct {
	patterns []regexPattern
	tokens   []Token
	source   *buffer
	pos      int
}

func Tokenize(source io.Reader) []Token {
	lex := createLexer(source)

	for !lex.at_eof() {
		matched := false

		for _, pattern := range lex.patterns {
			// lex.source.debug()
			loc := pattern.regex.FindReaderIndex(lex.remainder())
			if loc != nil && loc[0] == 0 {
				pattern.handler(lex, pattern.regex)
				matched = true
				break
			}
		}

		if !matched {
			panic(
				fmt.Sprintf(
					"Lexer::Error -> unrecognized token near %s",
					string(lex.remainderString()),
				),
			)
		}
	}
	lex.push(NewToken(EOF, "EOF"))
	return lex.tokens
}

func defaultHandler(kind TokenKind, value string) regexHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		lex.advance(len(value))
		lex.push(NewToken(kind, value))
	}
}

func stringHandler(l *lexer, regex *regexp.Regexp) {
	index := regex.FindReaderIndex(l.remainder())
	if index == nil {
		panic("Should have found something")
	}
	match := l.source.subString(index[0], index[1]) // [l.pos+index[0] : l.pos+index[1]]
	l.push(NewToken(IDENTIFIER, match))
	l.advance(len(match))
}

func (l *lexer) advance(count int) {
	l.pos += count
	l.source.pos += count
}

func (l *lexer) push(token Token) {
	l.tokens = append(l.tokens, token)
}

func (l *lexer) at_eof() bool {
	return l.source.eof && l.pos >= len(l.source.cache)
}

func (l *lexer) remainderString() string {
	return string(l.source.cache[l.pos:])
}

func (l *lexer) remainder() io.RuneReader {
	return bufio.NewReader(l.source.fork()) //bytes.NewBufferString(l.source[l.pos:])
}

func createLexer(source io.Reader) *lexer {
	return &lexer{
		pos:    0,
		source: newBuffer(source),
		tokens: []Token{},
		patterns: []regexPattern{
			{regexp.MustCompile("</"), defaultHandler(TAG_CLOSE_START, "</")},
			{regexp.MustCompile("<"), defaultHandler(TAG_START, "<")},
			{regexp.MustCompile("/>"), defaultHandler(TAG_CLOSE_END, "/>")},
			{regexp.MustCompile(">"), defaultHandler(TAG_END, ">")},
			{regexp.MustCompile(`\w+`), stringHandler},
		},
	}
}
