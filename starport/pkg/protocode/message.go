package protocode

import (
	"fmt"

	"github.com/emicklei/proto"
)

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
		case *proto.NormalField:
			message.AppendNormalField(value)
		case *proto.OneOfField:
			message.AppendOneOfField(value)
		case *proto.MapField:
			message.AppendMapField(value)
		case *proto.Message:
			message.AppendMessage(value)
		case *proto.Option:
			message.AppendOption(value)
		case *proto.Enum:
			message.AppendEnum(value)
		}
	}
	return message
}

func CreateMessagef(format string, args ...interface{}) *Message {
	return CreateMessage(fmt.Sprintf(format, args...))
}

func CreateMessage(name string) *Message {
	return NewMessage(&proto.Message{Name: name})
}

func (message *Message) FindField(name string) (*proto.NormalField, error) {
	if idx := message.IndexOfField(name); idx != -1 {
		return message.fields[idx], nil
	}
	return nil, fmt.Errorf("%w %q in message %q", ErrFieldNotFound, name, message.Name)
}

func (message *Message) FindEnum(name string) (*proto.Enum, error) {
	if idx := message.IndexOfEnum(name); idx != -1 {
		return message.enums[idx], nil
	}
	return nil, fmt.Errorf("%w %q in message %q", ErrEnumNotFound, name, message.Name)
}

func (message *Message) IndexOfField(name string) int {
	for idx, item := range message.fields {
		if item.Name == name {
			return idx
		}
	}
	return -1
}

func (message *Message) IndexOfEnum(name string) int {
	for idx, item := range message.enums {
		if item.Name == name {
			return idx
		}
	}
	return -1
}

func (message *Message) AppendNormalField(field *proto.NormalField) {
	message.fields = append(message.fields, field)
	message.fixupFieldSequence(field.Field)
}

func (message *Message) AppendOneOfField(field *proto.OneOfField) {
	message.oneofs = append(message.oneofs, field)
	message.fixupFieldSequence(field.Field)
}

func (message *Message) AppendMapField(field *proto.MapField) {
	message.maps = append(message.maps, field)
	message.fixupFieldSequence(field.Field)
}

func (message *Message) AppendField(field *proto.Field) {
	message.AppendNormalField(&proto.NormalField{Field: field})
}

func (message *Message) AppendRepeatedField(field *proto.Field) {
	message.AppendNormalField(&proto.NormalField{Field: field, Repeated: true})
}

func (message *Message) AppendFields(fields ...*proto.Field) {
	for _, field := range fields {
		message.AppendField(field)
	}
}

func (message *Message) AppendMessage(input *proto.Message) {
	message.messages = append(message.messages, input)
}

func (message *Message) AppendOption(option *proto.Option) {
	message.options = append(message.options, option)
}

func (message *Message) AppendEnum(enum *proto.Enum) {
	message.enums = append(message.enums, enum)
}

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
	if field.Sequence != 0 {
		message.max = maxInt(message.max, field.Sequence)
		return
	}
	message.max++
	field.Sequence = message.max
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
