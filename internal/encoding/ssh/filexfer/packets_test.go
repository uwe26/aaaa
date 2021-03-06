package filexfer

import (
	"bytes"
	"testing"
)

func TestRawPacket(t *testing.T) {
	const (
		id      = 42
		errMsg  = "eof"
		langTag = "en"
	)

	p := &RawPacket{
		Type:      PacketTypeStatus,
		RequestID: id,
		Data: Buffer{
			b: []byte{
				0x00, 0x00, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x03, 'e', 'o', 'f',
				0x00, 0x00, 0x00, 0x02, 'e', 'n',
			},
		},
	}

	data, err := p.MarshalBinary()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	want := []byte{
		0x00, 0x00, 0x00, 22,
		101,
		0x00, 0x00, 0x00, 42,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 3, 'e', 'o', 'f',
		0x00, 0x00, 0x00, 2, 'e', 'n',
	}

	if !bytes.Equal(data, want) {
		t.Errorf("RawPacket.Marshal() = %X, but wanted %X", data, want)
	}

	*p = RawPacket{}

	if err := p.ReadFrom(bytes.NewReader(data), nil, DefaultMaxPacketLength); err != nil {
		t.Fatal("unexpected error:", err)
	}

	if p.Type != PacketTypeStatus {
		t.Errorf("RawPacket.UnmarshalBinary(): Type was %v, but expected %v", p.Type, PacketTypeStat)
	}

	if p.RequestID != uint32(id) {
		t.Errorf("RawPacket.UnmarshalBinary(): RequestID was %d, but expected %d", p.RequestID, id)
	}

	want = []byte{
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 3, 'e', 'o', 'f',
		0x00, 0x00, 0x00, 2, 'e', 'n',
	}

	if !bytes.Equal(p.Data.Bytes(), want) {
		t.Fatalf("RawPacket.UnmarshalBinary(): Data was %X, but expected %X", p.Data, want)
	}

	var resp StatusPacket
	resp.UnmarshalPacketBody(&p.Data)

	if resp.StatusCode != StatusEOF {
		t.Errorf("UnmarshalPacketBody(RawPacket.Data): StatusCode was %v, but expected %v", resp.StatusCode, StatusEOF)
	}

	if resp.ErrorMessage != errMsg {
		t.Errorf("UnmarshalPacketBody(RawPacket.Data): ErrorMessage was %q, but expected %q", resp.ErrorMessage, errMsg)
	}

	if resp.LanguageTag != langTag {
		t.Errorf("UnmarshalPacketBody(RawPacket.Data): LanguageTag was %q, but expected %q", resp.LanguageTag, langTag)
	}
}

func TestRequestPacket(t *testing.T) {
	const (
		id   = 42
		path = "foo"
	)

	p := &RequestPacket{
		RequestID: id,
		Request: &StatPacket{
			Path: path,
		},
	}

	data, err := p.MarshalBinary()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	want := []byte{
		0x00, 0x00, 0x00, 12,
		17,
		0x00, 0x00, 0x00, 42,
		0x00, 0x00, 0x00, 3, 'f', 'o', 'o',
	}

	if !bytes.Equal(data, want) {
		t.Errorf("RequestPacket.Marshal() = %X, but wanted %X", data, want)
	}

	*p = RequestPacket{}

	if err := p.ReadFrom(bytes.NewReader(data), nil, DefaultMaxPacketLength); err != nil {
		t.Fatal("unexpected error:", err)
	}

	if p.RequestID != uint32(id) {
		t.Errorf("RequestPacket.UnmarshalBinary(): RequestID was %d, but expected %d", p.RequestID, id)
	}

	req, ok := p.Request.(*StatPacket)
	if !ok {
		t.Fatalf("unexpected Request type was %T, but expected %T", p.Request, req)
	}

	if req.Path != path {
		t.Errorf("RequestPacket.UnmarshalBinary(): Request.Path was %q, but expected %q", req.Path, path)
	}
}
