package main

import (
	"fmt"
	"google.golang.org/grpc/encoding"
)

type ExecuteRunRequest struct {
	RunID   string
	AgentID string
	Input   string
}
type ExecuteRunResponse struct {
	RunID  string
	Status string
	Result string
}
type protoWireCodec struct{}

func (protoWireCodec) Name() string { return "proto" }
func putString(dst []byte, field byte, value string) []byte {
	if value == "" {
		return dst
	}
	dst = append(dst, field<<3|2)
	n := len(value)
	for n >= 128 {
		dst = append(dst, byte(n)|128)
		n >>= 7
	}
	dst = append(dst, byte(n))
	return append(dst, value...)
}
func (protoWireCodec) Marshal(v any) ([]byte, error) {
	var b []byte
	switch x := v.(type) {
	case *ExecuteRunRequest:
		b = putString(b, 1, x.RunID)
		b = putString(b, 2, x.AgentID)
		b = putString(b, 3, x.Input)
	case *ExecuteRunResponse:
		b = putString(b, 1, x.RunID)
		b = putString(b, 2, x.Status)
		b = putString(b, 3, x.Result)
	default:
		return nil, fmt.Errorf("unsupported message %T", v)
	}
	return b, nil
}
func readString(b []byte, i *int) (string, error) {
	n, shift := 0, 0
	for {
		if *i >= len(b) || shift > 28 {
			return "", fmt.Errorf("invalid protobuf length")
		}
		c := int(b[*i])
		*i++
		n |= (c & 127) << shift
		if c < 128 {
			break
		}
		shift += 7
	}
	if n < 0 || *i+n > len(b) {
		return "", fmt.Errorf("invalid protobuf payload")
	}
	s := string(b[*i : *i+n])
	*i += n
	return s, nil
}
func decode(b []byte, fields []string) error {
	for i := 0; i < len(b); {
		tag := b[i]
		i++
		if tag&7 != 2 {
			return fmt.Errorf("unsupported protobuf wire type")
		}
		field := int(tag >> 3)
		s, err := readString(b, &i)
		if err != nil {
			return err
		}
		if field >= 1 && field <= len(fields) {
			fields[field-1] = s
		}
	}
	return nil
}
func (protoWireCodec) Unmarshal(b []byte, v any) error {
	switch x := v.(type) {
	case *ExecuteRunRequest:
		f := make([]string, 3)
		if err := decode(b, f); err != nil {
			return err
		}
		x.RunID, x.AgentID, x.Input = f[0], f[1], f[2]
	case *ExecuteRunResponse:
		f := make([]string, 3)
		if err := decode(b, f); err != nil {
			return err
		}
		x.RunID, x.Status, x.Result = f[0], f[1], f[2]
	default:
		return fmt.Errorf("unsupported message %T", v)
	}
	return nil
}

var _ encoding.Codec = protoWireCodec{}
