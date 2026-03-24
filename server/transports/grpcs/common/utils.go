package common

import (
	"google.golang.org/protobuf/types/known/structpb"
)

func ErrString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func DecodeStruct(s *structpb.Struct) map[string]any {
	if s == nil {
		return nil
	}
	return s.AsMap()
}
