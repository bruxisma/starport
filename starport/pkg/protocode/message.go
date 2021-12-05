package protocode

import (
	"fmt"

	"github.com/emicklei/proto"
)

// Message represents a "deconstructed" protobuf Message, where each item in
// the message is more readily available.
//
// The downside to using this type is that the order of items will change when
// it is reconstructed (however, comments will stay in the same location)
type Message struct {
	*proto.Message
	max      int
	fields   []*proto.NormalField
	oneofs   []*proto.OneOfField
	maps     []*proto.MapField
	messages []*proto.Message
	options  []*proto.Option
	enums    []*proto.Enum
}

// NewMessage returns a Message from a proto.Message
func NewMessage(input *proto.Message) *Message {
	message := &Message{
		Message:  input,
		fields:   []*proto.NormalField{},
		oneofs:   []*proto.OneOfField{},
		maps:     []*proto.MapField{},
		messages: []*proto.Message{},
		options:  []*proto.Option{},
		enums:    []*proto.Enum{},
	}
	for _, item := range input.Elements {
		switch value := item.(type) {
		case *proto.NormalField, *proto.OneOfField, *proto.MapField,
			*proto.Message, *proto.Option, *proto.Enum:
			message.Append(value)
		}
	}
	return message
}

// CreateMessagef returns a Message whose name is created from the format
// specifiers provided.
func CreateMessagef(format string, args ...interface{}) *Message {
	return CreateMessage(fmt.Sprintf(format, args...))
}

// CreateMessage returns a Message whose name is the one provided.
func CreateMessage(name string) *Message {
	return NewMessage(&proto.Message{Name: name})
}

// FindField returns a NormalField inside the Message with the name provided
func (message *Message) FindField(name string) (*proto.NormalField, error) {
	if idx := message.IndexOfField(name); idx != -1 {
		return message.fields[idx], nil
	}
	return nil, fmt.Errorf("%w %q in message %q", ErrFieldNotFound, name, message.Name)
}

// FindEnum returns an proto.Enum inside the Message with the name provided
func (message *Message) FindEnum(name string) (*proto.Enum, error) {
	if idx := message.IndexOfEnum(name); idx != -1 {
		return message.enums[idx], nil
	}
	return nil, fmt.Errorf("%w %q in message %q", ErrEnumNotFound, name, message.Name)
}

// IndexOfField returns the index of a field inside the received Message with
// the name provided. If no such field is found, it returns -1
func (message *Message) IndexOfField(name string) int {
	for idx, item := range message.fields {
		if item.Name == name {
			return idx
		}
	}
	return -1
}

// IndexOfEnum returns the index of a enum inside the received Message with
// the name provided. If no such enum is found, it returns -1
func (message *Message) IndexOfEnum(name string) int {
	for idx, item := range message.enums {
		if item.Name == name {
			return idx
		}
	}
	return -1
}

// Extend calls Append on each item provided in the order they are passed to
// the function
func (message *Message) Extend(items ...interface{}) {
	for _, item := range items {
		message.Append(item)
	}
}

// Append takes any type of item that can be placed inside a Message and places
// it into the correct collection.
//
// If a value passed is *not* a know type, this function will panic
func (message *Message) Append(item interface{}) {
	switch value := item.(type) {
	case *proto.NormalField:
		message.AppendNormalField(*value)
	case proto.NormalField:
		message.AppendNormalField(value)
	case *proto.OneOfField:
		message.AppendOneOfField(*value)
	case proto.OneOfField:
		message.AppendOneOfField(value)
	case *proto.MapField:
		message.AppendMapField(*value)
	case proto.MapField:
		message.AppendMapField(value)
	case *proto.Field:
		message.AppendField(*value)
	case proto.Field:
		message.AppendField(value)
	case *proto.Message:
		message.AppendMessage(value)
	case *proto.Enum:
		message.AppendEnum(value)
	case *proto.Option:
		message.AppendOption(value)
	default:
		panic(fmt.Sprintf("protocode.Message.Append provided unknown type %[1]T: %[1]v", value))
	}
}

// AppendNormalField appends a NormalField, but fixes up its Sequence to be
// correct
func (message *Message) AppendNormalField(field proto.NormalField) {
	message.fields = append(message.fields, &field)
	message.fixupFieldSequence(field.Field)
}

// AppendOneOfField appends a OneOfField but fixes up its Sequence to be
// correct
func (message *Message) AppendOneOfField(field proto.OneOfField) {
	message.oneofs = append(message.oneofs, &field)
	message.fixupFieldSequence(field.Field)
}

// AppendMapField appends a OneOfField but fixes up its Sequence to be
// correct
func (message *Message) AppendMapField(field proto.MapField) {
	message.maps = append(message.maps, &field)
	message.fixupFieldSequence(field.Field)
}

// AppendField appends a NormalField constructed from the provided Field, while
// also fixing up its Sequence to be correct
func (message *Message) AppendField(field proto.Field) {
	message.AppendNormalField(proto.NormalField{Field: &field})
}

// AppendRepeatedField appends a NormalField constructed from the provided
// field, while also marking the field as repeated.
func (message *Message) AppendRepeatedField(field proto.Field) {
	message.AppendNormalField(proto.NormalField{Field: &field, Repeated: true})
}

// AppendFields append multiple Fields by calling AppendField for each one
// provided.
func (message *Message) AppendFields(fields ...proto.Field) {
	for _, field := range fields {
		message.AppendField(field)
	}
}

// AppendMessage appends the provided message to the received Message
func (message *Message) AppendMessage(input *proto.Message) {
	message.messages = append(message.messages, input)
}

// AppendOption appends the provided option to the received Message
func (message *Message) AppendOption(option *proto.Option) {
	message.options = append(message.options, option)
}

// AppendEnum appends an enumeration to the received Message
func (message *Message) AppendEnum(enum *proto.Enum) {
	message.enums = append(message.enums, enum)
}

// Proto returns a proto.Message using the received Message
func (message *Message) Proto() *proto.Message {
	length := len(message.fields) +
		len(message.oneofs) +
		len(message.maps) +
		len(message.messages) +
		len(message.options) +
		len(message.enums)
	elements := make([]proto.Visitee, 0, length)
	for _, item := range message.options {
		elements = append(elements, item)
	}
	for _, item := range message.messages {
		elements = append(elements, item)
	}
	for _, item := range message.enums {
		elements = append(elements, item)
	}
	for _, item := range message.fields {
		elements = append(elements, item)
	}
	for _, item := range message.oneofs {
		elements = append(elements, item)
	}
	for _, item := range message.maps {
		elements = append(elements, item)
	}
	message.Message.Elements = elements
	return message.Message
}

func (message *Message) fixupFieldSequence(field *proto.Field) {
	message.max++
	field.Sequence = message.max
}
