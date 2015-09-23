package goha

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"strings"
)

type credentials struct {
	username   string
	password   string
	realm      string
	nonce      string
	digestURI  string
	algorithm  string
	cnonce     string
	opaque     string
	qop        string
	nonceCount int
	method     string
}

func newCredentials(username, password, header, uri, method string) *credentials {
	d := newDigestHeader(header)

	realm := d.realm()
	nonce := d.nonce()
	algorithm := d.algorithm()
	opaque := d.opaque()
	qop := d.qop()
	cnonce := randomNonce()

	return &credentials{username, password, realm, nonce, uri, algorithm, cnonce, opaque, qop, 0, method}
}

func (c *credentials) authHeader() string {
	var sl []string

	sl = append(sl, `username="`+c.username+`"`)
	sl = append(sl, `realm="`+c.realm+`"`)
	sl = append(sl, `nonce="`+c.nonce+`"`)
	sl = append(sl, `uri="`+c.digestURI+`"`)
	sl = append(sl, `response="`+c.response()+`"`)

	if c.opaque != "" {
		sl = append(sl, `opaque="`+c.opaque+`"`)
	}

	if c.qop != "" {
		sl = append(sl, "qop="+c.qop)
		sl = append(sl, "nc="+c.nonceCountStr())
		sl = append(sl, `cnonce="`+c.cnonce+`"`)
	}

	if c.algorithm != "" {
		sl = append(sl, "algorithm="+c.algorithm)
	}

	return "Digest " + strings.Join(sl, ", ")
}

func (c *credentials) response() string {
	c.nonceCount++

	if c.qop == "" {
		return h(c.ha1(), c.nonce, c.ha2())
	}

	return h(c.ha1(), c.nonce, c.nonceCountStr(), c.cnonce, c.qop, c.ha2())
}

func (c *credentials) nonceCountStr() string {
	return fmt.Sprintf("%08x", c.nonceCount)
}

func (c *credentials) ha1() string {
	return h(c.username, c.realm, c.password)
}

func (c *credentials) ha2() string {
	return h(c.method, c.digestURI)
}

type digestHeader struct {
	header string
}

func newDigestHeader(header string) digestHeader {
	return digestHeader{header}
}

func (d digestHeader) realm() string {
	return parseDirective(d.header, "realm")
}

func (d digestHeader) nonce() string {
	return parseDirective(d.header, "nonce")
}

func (d digestHeader) algorithm() string {
	return parseDirective(d.header, "algorithm")
}

func (d digestHeader) opaque() string {
	return parseDirective(d.header, "opaque")
}

func (d digestHeader) qop() string {
	return parseDirective(d.header, "qop")
}

func h(parts ...string) string {
	var content bytes.Buffer

	for _, part := range parts {
		content.WriteString(part)
		content.WriteString(":")
	}

	data := content.String()
	data = data[:len(data)-1]

	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

func randomNonce() string {
	b := make([]byte, 8)
	io.ReadFull(rand.Reader, b)

	return fmt.Sprintf("%x", b)[:16]
}

func parseDirective(header, name string) string {
	index := strings.Index(header, name)

	if index == -1 {
		return ""
	}

	start := 1 + index + strings.Index(header[index:], `"`)
	end := start + strings.Index(header[start:], `"`)

	return strings.TrimSpace(header[start:end])
}