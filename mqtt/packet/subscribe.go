package packet

import "bytes"

type Subscribe struct {
	id     int
	topics []Topic
}

type Topic struct {
	Name string
	qos  int
}

func (s *Subscribe) AddTopic(topic Topic) {
	s.topics = append(s.topics, topic)
}

func (s *Subscribe) PacketType() int {
	return subscribeType
}

func (s *Subscribe) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	// Empty topic list is incorrect. Server must disconnect.
	if len(s.topics) == 0 {
		return packet
	}

	variablePart.Write(encodeUint16(uint16(s.id)))

	for _, topic := range s.topics {
		variablePart.Write(encodeString(topic.Name))
		// TODO Check that QOS is valid
		variablePart.WriteByte(byte(topic.qos))
	}

	fixedHeaderFlags := 2 // mandatory value
	fixedHeader := (subscribeType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func decodeSubscribe(payload []byte) *Subscribe {
	subscribe := new(Subscribe)
	return subscribe
}

/*
  defstruct id: 0, qos: 0, topics: []

  def encode(%MQTT.Packet.Subscribe{topics: []}) do
    {:error, :empty_topics}
  end

  def encode(packet) do
    fixed_header_flag = 2
    variable_header = << packet.id :: size(16) >>

    qos = << packet.qos :: size(8) >>
    topics = packet.topics
             |> Enum.reduce(<<>>, fn(topic, acc) -> acc <> encode_string(topic) <> qos end)

    variable_part = variable_header <> topics

    remaining_size = remaining_length(byte_size(variable_part))

    fixed_header = << @mqtt_type :: size(4),  fixed_header_flag :: size(4)>> <> remaining_size
    fixed_header <> variable_part
  end

  def decode(_hflag, _rest) do
    %MQTT.Packet.Subscribe{}
  end
*/
