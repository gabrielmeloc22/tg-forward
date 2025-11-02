package telegram

import (
	"bytes"
	"testing"

	"github.com/gotd/td/crypto"
	"github.com/gotd/td/session"
)

func TestTelethonSessionRoundTrip(t *testing.T) {
	authKey := bytes.Repeat([]byte{'a'}, 256)
	var key crypto.Key
	copy(key[:], authKey)
	authKeyWithID := key.WithID()

	tests := []struct {
		name string
		data *session.Data
	}{
		{
			name: "IPv4",
			data: &session.Data{
				DC:        2,
				Addr:      "192.168.0.1:443",
				AuthKey:   authKeyWithID.Value[:],
				AuthKeyID: authKeyWithID.ID[:],
			},
		},
		{
			name: "IPv6",
			data: &session.Data{
				DC:        2,
				Addr:      "[2001:db8::1]:443",
				AuthKey:   authKeyWithID.Value[:],
				AuthKeyID: authKeyWithID.ID[:],
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := encodeTelethonSession(tt.data)
			if err != nil {
				t.Fatalf("encodeTelethonSession() error = %v", err)
			}

			t.Logf("Encoded session: %s (length: %d)", encoded, len(encoded))

			decoded, err := decodeTelethonSession(encoded)
			if err != nil {
				t.Fatalf("decodeTelethonSession() error = %v", err)
			}

			if decoded.DC != tt.data.DC {
				t.Errorf("DC mismatch: got %d, want %d", decoded.DC, tt.data.DC)
			}

			if decoded.Addr != tt.data.Addr {
				t.Errorf("Addr mismatch: got %s, want %s", decoded.Addr, tt.data.Addr)
			}

			if !bytes.Equal(decoded.AuthKey, tt.data.AuthKey) {
				t.Errorf("AuthKey mismatch")
			}

			if !bytes.Equal(decoded.AuthKeyID, tt.data.AuthKeyID) {
				t.Errorf("AuthKeyID mismatch")
			}
		})
	}
}

func TestDecodeTelethonSession(t *testing.T) {
	tests := []struct {
		name    string
		session string
		wantErr bool
	}{
		{
			name:    "Invalid session",
			session: "invalid",
			wantErr: true,
		},
		{
			name:    "Empty session",
			session: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeTelethonSession(tt.session)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeTelethonSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncodeTelethonSession(t *testing.T) {
	tests := []struct {
		name    string
		data    *session.Data
		wantErr bool
	}{
		{
			name: "Invalid address format",
			data: &session.Data{
				DC:      2,
				Addr:    "invalid-address",
				AuthKey: make([]byte, 256),
			},
			wantErr: true,
		},
		{
			name: "Invalid IP",
			data: &session.Data{
				DC:      2,
				Addr:    "invalid-ip:443",
				AuthKey: make([]byte, 256),
			},
			wantErr: true,
		},
		{
			name: "Invalid port",
			data: &session.Data{
				DC:      2,
				Addr:    "192.168.0.1:invalid",
				AuthKey: make([]byte, 256),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encodeTelethonSession(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeTelethonSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
