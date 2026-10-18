package service

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"hash/crc32"
	"io"
	"sync/atomic"
	"testing"

	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type zipEntry struct {
	name    string
	data    []byte
	method  uint16
	raw     bool
	rawSize uint64
	crc     uint32
}

func buildZip(t *testing.T, entries ...zipEntry) *bytes.Reader {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		if e.raw {
			w, err := zw.CreateRaw(&zip.FileHeader{
				Name: e.name, Method: e.method, CRC32: e.crc,
				CompressedSize64: uint64(len(e.data)), UncompressedSize64: e.rawSize,
			})
			require.NoError(t, err)
			_, err = w.Write(e.data)
			require.NoError(t, err)
			continue
		}
		w, err := zw.CreateHeader(&zip.FileHeader{Name: e.name, Method: e.method})
		require.NoError(t, err)
		_, err = w.Write(e.data)
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return bytes.NewReader(buf.Bytes())
}

func siqXML(mediaFiles ...string) []byte {
	var items string
	for _, m := range mediaFiles {
		items += fmt.Sprintf(`<param name="question" type="content"><item type="image" isRef="True">%s</item></param>`, m)
	}
	if items == "" {
		items = `<param name="question" type="content"><item>text</item></param>`
	}
	return []byte(`<package name="p"><rounds><round name="r"><themes><theme name="t"><questions>` +
		`<question price="100"><params>` + items + `</params><right><answer>a</answer></right></question>` +
		`</questions></theme></themes></round></rounds></package>`)
}

func zeros(n int) []byte { return make([]byte, n) }

// rawDeflateEntry stores a deflated payload but declares a (false) uncompressed size
func rawDeflateEntry(t *testing.T, name string, payload []byte, declaredSize uint64) zipEntry {
	t.Helper()
	var compressed bytes.Buffer
	fw, err := flate.NewWriter(&compressed, flate.BestCompression)
	require.NoError(t, err)
	_, err = fw.Write(payload)
	require.NoError(t, err)
	require.NoError(t, fw.Close())
	return zipEntry{name: name, data: compressed.Bytes(), method: zip.Deflate, rawSize: declaredSize, raw: true, crc: crc32.ChecksumIEEE(payload)}
}

func TestParseSIQLimits(t *testing.T) {
	tests := []struct {
		name    string
		entries func() []zipEntry
		wantErr string
	}{
		{
			name: "valid pack with media",
			entries: func() []zipEntry {
				return []zipEntry{
					{name: "content.xml", data: siqXML("a.png"), method: zip.Deflate},
					{name: "Images/a.png", data: []byte("png"), method: zip.Store},
				}
			},
		},
		{
			name: "media over per-file limit",
			entries: func() []zipEntry {
				return []zipEntry{
					{name: "content.xml", data: siqXML("big.png"), method: zip.Deflate},
					{name: "Images/big.png", data: zeros(SIQ_MAX_MEDIA_FILE_SIZE + 1), method: zip.Store},
				}
			},
			wantErr: "per-file limit",
		},
		{
			name: "content.xml over its limit",
			entries: func() []zipEntry {
				return []zipEntry{
					{name: "content.xml", data: append(siqXML(), zeros(SIQ_MAX_CONTENT_XML_SIZE)...), method: zip.Store},
				}
			},
			wantErr: "per-file limit",
		},
		{
			name: "compressed bomb over per-file limit",
			entries: func() []zipEntry {
				return []zipEntry{
					{name: "content.xml", data: siqXML("bomb.png"), method: zip.Deflate},
					{name: "Images/bomb.png", data: zeros(SIQ_MAX_MEDIA_FILE_SIZE + 1), method: zip.Deflate},
				}
			},
			wantErr: "per-file limit",
		},
		{
			name: "too many entries",
			entries: func() []zipEntry {
				entries := []zipEntry{{name: "content.xml", data: siqXML(), method: zip.Deflate}}
				for i := range SIQ_MAX_ENTRIES {
					entries = append(entries, zipEntry{name: fmt.Sprintf("x/%d", i), method: zip.Store})
				}
				return entries
			},
			wantErr: "entries",
		},
	}

	svc := &PackDraftService{attachmentService: &AttachmentService{}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := buildZip(t, tt.entries()...)

			draft, uploads, err := svc.parseSIQ(r, r.Size())

			if tt.wantErr == "" {
				require.NoError(t, err)
				require.NotNil(t, draft)
				assert.Len(t, uploads, 1)
				return
			}
			var badRequest custerr.BadRequestErr
			require.ErrorAs(t, err, &badRequest)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestSIQLimitReader(t *testing.T) {
	tests := []struct {
		name    string
		entry   int64
		budget  int64
		size    int
		wantErr error
	}{
		{name: "within limits", entry: 100, budget: 100, size: 100},
		{name: "over entry limit", entry: 100, budget: 1000, size: 101, wantErr: errSIQEntryTooLarge},
		{name: "over total budget", entry: 1000, budget: 100, size: 101, wantErr: errSIQTotalTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var budget atomic.Int64
			budget.Store(tt.budget)
			r := &siqLimitReader{r: bytes.NewReader(zeros(tt.size)), left: tt.entry, budget: &budget}

			data, err := io.ReadAll(r)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, data, tt.size)
		})
	}
}

func TestParseSIQUnderDeclaredSizeIsSkipped(t *testing.T) {
	svc := &PackDraftService{attachmentService: &AttachmentService{}}
	r := buildZip(t,
		zipEntry{name: "content.xml", data: siqXML("liar.png"), method: zip.Deflate},
		rawDeflateEntry(t, "Images/liar.png", zeros(50<<20), 1000),
	)

	_, uploads, err := svc.parseSIQ(r, r.Size())

	require.NoError(t, err)
	assert.Empty(t, uploads)
}
