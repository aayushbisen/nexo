package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Value struct {
	Kind    byte // '+', '-', '$', '*'
	Str     string
	Integer int64
	Array   []Value
}

func (v Value) Write(wr io.Writer) error {
	switch v.Kind {
	case '+':
		fmt.Fprintf(wr, "+%s\r\n", v.Str)
	case '-':
		fmt.Fprintf(wr, "-%s\r\n", v.Str)
	case '$':
		fmt.Fprintf(wr, "$%d\r\n%s\r\n", len(v.Str), v.Str)
	case '*':
		fmt.Fprintf(wr, "*%d\r\n", len(v.Array))
		for _, val := range v.Array {
			val.Write(wr)
		}
	}

	return nil
}

func Read(r *bufio.Reader) (Value, error) {

	kind, err := r.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch kind {
	case '+':
		line, err := r.ReadString('\n')
		if err != nil {
			return Value{}, err
		}
		// line = "OK\r\n" — need to trim \r\n
		return Value{Kind: '+', Str: line[:len(line)-2]}, nil
	case '-':
		line, err := r.ReadString('\n')
		if err != nil {
			return Value{}, err
		}
		return Value{Kind: '-', Str: line[:len(line)-2]}, nil
	case '$':
		countStr, err := r.ReadString('\n')
		if err != nil {
			return Value{}, err
		}
		countInt, err := strconv.Atoi(countStr[:len(countStr)-2])
		if err != nil {
			return Value{}, err
		}
		buf := make([]byte, countInt)
		_, err = io.ReadFull(r, buf)

		r.ReadByte()
		r.ReadByte()

		return Value{Kind: '$', Str: string(buf)}, nil
	case '*':
		countStr, err := r.ReadString('\n')
		if err != nil {
			return Value{}, err
		}
		countInt, err := strconv.Atoi(countStr[:len(countStr)-2])
		buf := make([]Value, 0, countInt)
		for i := 0; i < countInt; i++ {
			val, err := Read(r)
			if err != nil {
				return Value{}, err
			}
			buf = append(buf, val)
		}
		return Value{Kind: '*', Array: buf}, nil
	}

	return Value{}, nil
}
