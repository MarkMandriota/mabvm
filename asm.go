package mabvm

import (
	"fmt"
)

type AsmParser struct {
	src string
	pos int
}

func (ap *AsmParser) Reset(src string) {
	ap.src = src
	ap.pos = 0
}

func (ap *AsmParser) readByte() {
	ap.pos++
}

func (ap *AsmParser) peekByte() byte {
	if ap.pos < len(ap.src) {
		return ap.src[ap.pos]
	}

	return 0
}

func (ap *AsmParser) Parse(dst []Opcode) ([]Opcode, error) {
	ret := dst

	if ret == nil {
		ret = make([]byte, 0, 1<<16)
	}

	for {
		op, err := ap.parseOpcode()
		if err != nil {
			return nil, fmt.Errorf("error in instruction %d: %v", len(ret), err)
		}
		if op == 0 {
			return ret, nil
		}

		ret = append(ret, op)
	}
}

func (ap *AsmParser) parseOpcode() (op Opcode, err error) {
	ap.nextSect()

	if ap.peekByte() != ':' {
		return 0, nil
	}

	ap.readByte()

	switch ap.peekByte() {
	case 'S':
		op = SJ
	case 'D':
		op = DJ
	case 'C':
		op = CJ
	case 'V':
		op = VJ
	default:
		return 0, fmt.Errorf("unexpected character with code %d in table section: expected table character (S, D, C, V)", ap.peekByte())
	}

	ap.readByte()

	if ap.peekByte() != ':' {
		return op, nil
	}

	ap.readByte()

	op |= ap.testFlag('I', IF)
	op |= ap.testFlag('E', EF)
	op |= ap.testFlag('M', MF)

	if ap.peekByte() != ':' {
		if b := ap.peekByte(); !(isVoid(b) || b == 0) {
			return 0, fmt.Errorf("unexpected character with code %d in flag section: expected flag character (I, E, M)", b)
		}
		return op, nil
	}

	ap.readByte()

	op |= ap.testFlag('L', LC)
	op |= ap.testFlag('E', EC)
	op |= ap.testFlag('G', GC)

	if b := ap.peekByte(); !(isVoid(b) || b == 0) {
		return 0, fmt.Errorf("unexpected character with code %d in conditional section: expected conditional flag character (L, E, G)", b)
	}

	return op, nil
}

func (ap *AsmParser) testFlag(name byte, code Opcode) Opcode {
	if ap.peekByte() == name {
		ap.readByte()
		return code
	}

	return 0
}

func (ap *AsmParser) nextSect() {
	for !isSectOrNull(ap.peekByte()) {
		ap.readByte()
	}
}

func (ap *AsmParser) nextPush() {
	for !isBegPushOrSectOrNull(ap.peekByte()) {
		ap.readByte()
	}
}

func (ap *AsmParser) readPushes() []string {
	pushes := make([]string, 0)

	for ap.nextPush(); ap.peekByte() == '<'; ap.nextPush() {
		beg := ap.pos

		for isEndPush(ap.peekByte()) {
			pushes = append(pushes, ap.src[beg:ap.pos])
		}
	}

	return pushes
}

func isVoid(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t' || b == '\v'
}

func isSectOrNull(b byte) bool {
	return b == ':' || b == 0
}

func isBegPushOrSectOrNull(b byte) bool {
	return b == '<' || isSectOrNull(b)
}

func isEndPush(b byte) bool {
	return b == '>' || b == 0
}
