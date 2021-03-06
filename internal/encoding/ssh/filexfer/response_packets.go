package filexfer

// StatusPacket defines the SSH_FXP_STATUS packet.
//
// Specified in https://tools.ietf.org/html/draft-ietf-secsh-filexfer-02#section-7
type StatusPacket struct {
	StatusCode   Status
	ErrorMessage string
	LanguageTag  string
}

// MarshalPacket returns p as a two-part binary encoding of p.
func (p *StatusPacket) MarshalPacket(reqid uint32, b []byte) (header, payload []byte, err error) {
	buf := NewBuffer(b)
	if buf.Cap() < 9 {
		// uint32(error/status code) + string(error message) + string(language tag)
		size := 4 + 4 + len(p.ErrorMessage) + 4 + len(p.LanguageTag)
		buf = NewMarshalBuffer(size)
	}

	buf.StartPacket(PacketTypeStatus, reqid)
	buf.AppendUint32(uint32(p.StatusCode))
	buf.AppendString(p.ErrorMessage)
	buf.AppendString(p.LanguageTag)

	return buf.Packet(payload)
}

// UnmarshalPacketBody unmarshals the packet body from the given Buffer.
// It is assumed that the uint32(request-id) has already been consumed.
func (p *StatusPacket) UnmarshalPacketBody(buf *Buffer) (err error) {
	statusCode, err := buf.ConsumeUint32()
	if err != nil {
		return err
	}
	p.StatusCode = Status(statusCode)

	if p.ErrorMessage, err = buf.ConsumeString(); err != nil {
		return err
	}

	if p.LanguageTag, err = buf.ConsumeString(); err != nil {
		return err
	}

	return nil
}

// HandlePacket defines the SSH_FXP_HANDLE packet.
type HandlePacket struct {
	Handle string
}

// MarshalPacket returns p as a two-part binary encoding of p.
func (p *HandlePacket) MarshalPacket(reqid uint32, b []byte) (header, payload []byte, err error) {
	buf := NewBuffer(b)
	if buf.Cap() < 9 {
		size := 4 + len(p.Handle) // string(handle)
		buf = NewMarshalBuffer(size)
	}

	buf.StartPacket(PacketTypeHandle, reqid)
	buf.AppendString(p.Handle)

	return buf.Packet(payload)
}

// UnmarshalPacketBody unmarshals the packet body from the given Buffer.
// It is assumed that the uint32(request-id) has already been consumed.
func (p *HandlePacket) UnmarshalPacketBody(buf *Buffer) (err error) {
	if p.Handle, err = buf.ConsumeString(); err != nil {
		return err
	}

	return nil
}

// DataPacket defines the SSH_FXP_DATA packet.
type DataPacket struct {
	Data []byte
}

// MarshalPacket returns p as a two-part binary encoding of p.
func (p *DataPacket) MarshalPacket(reqid uint32, b []byte) (header, payload []byte, err error) {
	buf := NewBuffer(b)
	if buf.Cap() < 9 {
		size := 4 // uint32(len(data)); data content in payload
		buf = NewMarshalBuffer(size)
	}

	buf.StartPacket(PacketTypeData, reqid)
	buf.AppendUint32(uint32(len(p.Data)))

	return buf.Packet(p.Data)
}

// UnmarshalPacketBody unmarshals the packet body from the given Buffer.
// It is assumed that the uint32(request-id) has already been consumed.
func (p *DataPacket) UnmarshalPacketBody(buf *Buffer) (err error) {
	if p.Data, err = buf.ConsumeByteSlice(); err != nil {
		return err
	}

	return nil
}

// NamePacket defines the SSH_FXP_NAME packet.
type NamePacket struct {
	Entries []*NameEntry
}

// MarshalPacket returns p as a two-part binary encoding of p.
func (p *NamePacket) MarshalPacket(reqid uint32, b []byte) (header, payload []byte, err error) {
	buf := NewBuffer(b)
	if buf.Cap() < 9 {
		size := 4 // uint32(len(entries))

		for _, e := range p.Entries {
			size += e.Len()
		}

		buf = NewMarshalBuffer(size)
	}

	buf.StartPacket(PacketTypeName, reqid)
	buf.AppendUint32(uint32(len(p.Entries)))

	for _, e := range p.Entries {
		e.MarshalInto(buf)
	}

	return buf.Packet(payload)
}

// UnmarshalPacketBody unmarshals the packet body from the given Buffer.
// It is assumed that the uint32(request-id) has already been consumed.
func (p *NamePacket) UnmarshalPacketBody(buf *Buffer) (err error) {
	count, err := buf.ConsumeUint32()
	if err != nil {
		return err
	}

	p.Entries = make([]*NameEntry, 0, count)

	for i := uint32(0); i < count; i++ {
		var e NameEntry
		if err := e.UnmarshalFrom(buf); err != nil {
			return err
		}

		p.Entries = append(p.Entries, &e)
	}

	return nil
}

// AttrsPacket defines the SSH_FXP_ATTRS packet.
type AttrsPacket struct {
	Attrs Attributes
}

// MarshalPacket returns p as a two-part binary encoding of p.
func (p *AttrsPacket) MarshalPacket(reqid uint32, b []byte) (header, payload []byte, err error) {
	buf := NewBuffer(b)
	if buf.Cap() < 9 {
		size := p.Attrs.Len() // ATTRS(attrs)
		buf = NewMarshalBuffer(size)
	}

	buf.StartPacket(PacketTypeAttrs, reqid)
	p.Attrs.MarshalInto(buf)

	return buf.Packet(payload)
}

// UnmarshalPacketBody unmarshals the packet body from the given Buffer.
// It is assumed that the uint32(request-id) has already been consumed.
func (p *AttrsPacket) UnmarshalPacketBody(buf *Buffer) (err error) {
	return p.Attrs.UnmarshalFrom(buf)
}
