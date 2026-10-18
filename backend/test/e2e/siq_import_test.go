package e2e

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/holdennekt/sgame/backend/test/e2e/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const siqContentXML = `<package name="E2E SIQ"><rounds><round name="Round 1"><themes><theme name="Theme"><questions>` +
	`<question price="100"><params>` +
	`<param name="question" type="content"><item type="image" isRef="True">a.png</item></param>` +
	`</params><right><answer>42</answer></right></question>` +
	`</questions></theme></themes></round></rounds></package>`

func buildSIQ(t *testing.T, media []byte, method uint16) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range []struct {
		name string
		data []byte
	}{{"content.xml", []byte(siqContentXML)}, {"Images/a.png", media}} {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: e.name, Method: method})
		require.NoError(t, err)
		_, err = w.Write(e.data)
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

type importResult struct {
	status int
	body   string
}

func importRequest(t *testing.T, app *testhelper.TestApp, session, contentType string, body io.Reader) (*http.Request, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, app.Server.URL+"/api/packs/drafts/import", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.AddCookie(&http.Cookie{Name: testhelper.SessionCookieName, Value: session})
	return req, nil
}

func importSIQ(t *testing.T, app *testhelper.TestApp, session string, siq []byte) importResult {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("siq", "pack.siq")
	require.NoError(t, err)
	_, err = part.Write(siq)
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	req, err := importRequest(t, app, session, mw.FormDataContentType(), &buf)
	require.NoError(t, err)
	resp, err := app.Server.Client().Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	out, _ := io.ReadAll(resp.Body)
	return importResult{resp.StatusCode, string(out)}
}

// heldImport is an import whose body is still being uploaded, so its handler keeps the slots.
type heldImport struct {
	t      *testing.T
	pw     *io.PipeWriter
	mw     *multipart.Writer
	result chan importResult
}

func startHeldImport(t *testing.T, app *testhelper.TestApp, session string) *heldImport {
	t.Helper()
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	req, err := importRequest(t, app, session, mw.FormDataContentType(), pr)
	require.NoError(t, err)

	h := &heldImport{t: t, pw: pw, mw: mw, result: make(chan importResult, 1)}
	go func() {
		resp, err := app.Server.Client().Do(req)
		if err != nil {
			h.result <- importResult{0, err.Error()}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		out, _ := io.ReadAll(resp.Body)
		h.result <- importResult{resp.StatusCode, string(out)}
	}()
	t.Cleanup(func() { _ = pw.CloseWithError(io.ErrClosedPipe) })
	return h
}

func (h *heldImport) finish(siq []byte) importResult {
	h.t.Helper()
	go func() {
		part, err := h.mw.CreateFormFile("siq", "pack.siq")
		if err == nil {
			_, err = part.Write(siq)
		}
		if err == nil {
			err = h.mw.Close()
		}
		_ = h.pw.CloseWithError(err)
	}()
	select {
	case res := <-h.result:
		return res
	case <-time.After(15 * time.Second):
		h.t.Fatal("held import did not finish")
		return importResult{}
	}
}

func waitForUserLock(t *testing.T, userId string) {
	t.Helper()
	rds := newRedisClient(t)
	require.Eventually(t, func() bool {
		return rds.Exists(t.Context(), "import:user:"+userId).Val() == 1
	}, 5*time.Second, 20*time.Millisecond, "import handler never took the user lock")
}

func TestSIQImport(t *testing.T) {
	app := newApp(t)
	validSIQ := buildSIQ(t, []byte("not really a png"), zip.Deflate)

	t.Run("guest is forbidden", func(t *testing.T) {
		guest := app.GuestSession(t, "GuestImporter")

		res := importSIQ(t, app, guest, validSIQ)

		assert.Equal(t, http.StatusForbidden, res.status, res.body)
	})

	t.Run("valid pack becomes a draft", func(t *testing.T) {
		session, _ := app.Register(t, "siq_valid", "password123")

		res := importSIQ(t, app, session, validSIQ)
		require.Equal(t, http.StatusCreated, res.status, res.body)

		var created struct {
			Id string `json:"id"`
		}
		require.NoError(t, json.Unmarshal([]byte(res.body), &created))
		require.NotEmpty(t, created.Id)

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, app.Server.URL+"/api/packs/drafts/"+created.Id, nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: testhelper.SessionCookieName, Value: session})
		resp, err := app.Server.Client().Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		out, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode, string(out))
		assert.Contains(t, string(out), "E2E SIQ")
	})

	t.Run("zip bomb is rejected", func(t *testing.T) {
		session, _ := app.Register(t, "siq_bomb", "password123")
		bomb := buildSIQ(t, make([]byte, 101<<20), zip.Deflate)
		require.Less(t, len(bomb), 1<<20, "bomb should be tiny on the wire")

		res := importSIQ(t, app, session, bomb)

		assert.Equal(t, http.StatusBadRequest, res.status, res.body)
		assert.Contains(t, res.body, "per-file limit")
	})

	t.Run("failed import frees the slot", func(t *testing.T) {
		session, _ := app.Register(t, "siq_retry", "password123")

		res := importSIQ(t, app, session, []byte("definitely not a zip"))
		assert.Equal(t, http.StatusBadRequest, res.status, res.body)

		res = importSIQ(t, app, session, validSIQ)
		assert.Equal(t, http.StatusCreated, res.status, res.body)
	})

	t.Run("second import from the same user is rejected", func(t *testing.T) {
		session, userId := app.Register(t, "siq_double", "password123")
		held := startHeldImport(t, app, session)
		waitForUserLock(t, userId)

		res := importSIQ(t, app, session, validSIQ)
		assert.Equal(t, http.StatusTooManyRequests, res.status, res.body)
		assert.Contains(t, res.body, "already have an import")

		assert.Equal(t, http.StatusCreated, held.finish(validSIQ).status)
		assert.Equal(t, http.StatusCreated, importSIQ(t, app, session, validSIQ).status)
	})

	t.Run("pod limit rejects imports from other users", func(t *testing.T) {
		var held []*heldImport
		for i := range 2 {
			session, userId := app.Register(t, fmt.Sprintf("siq_pod_%d", i), "password123")
			held = append(held, startHeldImport(t, app, session))
			waitForUserLock(t, userId)
		}

		session, _ := app.Register(t, "siq_pod_extra", "password123")
		res := importSIQ(t, app, session, validSIQ)
		assert.Equal(t, http.StatusTooManyRequests, res.status, res.body)
		assert.True(t, strings.Contains(res.body, "busy"), res.body)

		for _, h := range held {
			assert.Equal(t, http.StatusCreated, h.finish(validSIQ).status)
		}
		assert.Equal(t, http.StatusCreated, importSIQ(t, app, session, validSIQ).status)
	})
}
