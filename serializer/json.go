package serializer

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func ProtobuffToJSON(message proto.Message) ([]byte, error) {
	marshaller := protojson.MarshalOptions{
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
		Indent:          "   ",
		UseProtoNames:   true,
	}
	// marshaller := jsonpb.Marshaler{
	// 	EnumsAsInts:  false,
	// 	EmitDefaults: true,
	// 	Indent:       "   ",
	// 	OrigName:     true,
	// }
	return marshaller.Marshal(message)
}
