package drive

import (
	"bytes"
	"testing"
)

func TestCompression(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "Simple JSON",
			data: `{"name":"test","value":123}`,
		},
		{
			name: "Complex JSON",
			data: `{
				"profile": {
					"id": 1,
					"name": "John Doe",
					"email": "john@example.com"
				},
				"companies": [
					{"id": 1, "name": "Company A"},
					{"id": 2, "name": "Company B"}
				]
			}`,
		},
		{
			name: "Empty JSON",
			data: `{}`,
		},
		{
			name: "Large JSON",
			data: `{"data":"` + string(bytes.Repeat([]byte("x"), 10000)) + `"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := []byte(tt.data)

			// Compress
			compressed, err := compressJSON(original)
			if err != nil {
				t.Fatalf("Failed to compress: %v", err)
			}

			// Should be smaller (except for very small data)
			if len(original) > 100 && len(compressed) >= len(original) {
				t.Errorf("Compression didn't reduce size: original=%d, compressed=%d",
					len(original), len(compressed))
			}

			// Decompress
			decompressed, err := decompressJSON(compressed)
			if err != nil {
				t.Fatalf("Failed to decompress: %v", err)
			}

			// Should match original
			if !bytes.Equal(original, decompressed) {
				t.Errorf("Decompressed data doesn't match original")
			}
		})
	}
}

func TestCompressionErrors(t *testing.T) {
	t.Run("Invalid gzip data", func(t *testing.T) {
		_, err := decompressJSON([]byte("not gzip data"))
		if err == nil {
			t.Error("Expected error for invalid gzip data")
		}
	})

	t.Run("Empty data", func(t *testing.T) {
		_, err := decompressJSON([]byte{})
		if err == nil {
			t.Error("Expected error for empty data")
		}
	})
}

func BenchmarkCompression(b *testing.B) {
	data := []byte(`{
		"profile": {"id": 1, "name": "Test User"},
		"companies": [
			{"id": 1, "name": "Company 1"},
			{"id": 2, "name": "Company 2"}
		],
		"jobs": [
			{"id": 1, "title": "Job 1", "company_id": 1},
			{"id": 2, "title": "Job 2", "company_id": 2}
		]
	}`)

	b.Run("Compress", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = compressJSON(data)
		}
	})

	compressed, _ := compressJSON(data)

	b.Run("Decompress", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = decompressJSON(compressed)
		}
	})
}
