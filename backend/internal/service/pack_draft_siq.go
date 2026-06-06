package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"log"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
	"gopkg.in/vansante/go-ffprobe.v2"
)

// ── XML types (unexported) ────────────────────────────────────────────────────

type siqPackage struct {
	XMLName xml.Name  `xml:"package"`
	Name    string    `xml:"name,attr"`
	Rounds  siqRounds `xml:"rounds"`
}

type siqRounds struct {
	Rounds []siqRound `xml:"round"`
}

type siqRound struct {
	Name   string    `xml:"name,attr"`
	Type   string    `xml:"type,attr"`
	Themes siqThemes `xml:"themes"`
}

type siqThemes struct {
	Themes []siqTheme `xml:"theme"`
}

type siqTheme struct {
	Name      string        `xml:"name,attr"`
	Info      *siqThemeInfo `xml:"info"`
	Questions siqQuestions  `xml:"questions"`
}

type siqThemeInfo struct {
	Comments string `xml:"comments"`
}

type siqQuestions struct {
	Questions []siqQuestion `xml:"question"`
}

type siqQuestion struct {
	Price  int       `xml:"price,attr"`
	Type   string    `xml:"type,attr"`
	Params siqParams `xml:"params"`
	Right  siqRight  `xml:"right"`
}

type siqParams struct {
	Params []siqParam `xml:"param"`
}

type siqParam struct {
	Name      string        `xml:"name,attr"`
	Items     []siqItem     `xml:"item"`
	NumberSet *siqNumberSet `xml:"numberSet"`
	CharData  string        `xml:",chardata"`
}

type siqItem struct {
	ItemType string `xml:"type,attr"`
	IsRef    string `xml:"isRef,attr"`
	Content  string `xml:",chardata"`
}

type siqNumberSet struct {
	Minimum int `xml:"minimum,attr"`
}

type siqRight struct {
	Answers []siqAnswer `xml:"answer"`
}

type siqAnswer struct {
	Content string `xml:",chardata"`
}

// ── Small helpers ─────────────────────────────────────────────────────────────

var whitespaceRe = regexp.MustCompile(`\s+`)

func normalizeText(s string) string {
	return strings.TrimSpace(whitespaceRe.ReplaceAllString(s, " "))
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func siqTypeToQuestionType(t string) domain.QuestionType {
	switch t {
	case "secret":
		return domain.CatInBag
	case "stake", "stakeAll":
		return domain.Auction
	default:
		return domain.Regular
	}
}

func itemTypeToFolder(itemType string) string {
	switch itemType {
	case "audio":
		return "Audio"
	case "video":
		return "Video"
	default:
		return "Images"
	}
}

// extractContent returns (text, mediaZipPath) from a content param.
func extractContent(param *siqParam) (text string, mediaZipPath string) {
	if param == nil {
		return "", ""
	}
	var texts []string
	for _, item := range param.Items {
		content := normalizeText(item.Content)
		if item.IsRef == "True" && content != "" {
			if mediaZipPath == "" {
				itemType := item.ItemType
				if itemType == "" {
					itemType = "image"
				}
				mediaZipPath = itemTypeToFolder(itemType) + "/" + content
			}
		} else if content != "" {
			texts = append(texts, content)
		}
	}
	return strings.Join(texts, " "), mediaZipPath
}

func extractAnswers(right siqRight) []string {
	answers := make([]string, 0, len(right.Answers))
	for _, a := range right.Answers {
		if t := normalizeText(a.Content); t != "" {
			answers = append(answers, t)
		}
	}
	return answers
}

func buildComment(text, mediaPath string, reg func(string) *domain.Attachment) *domain.Comment {
	att := reg(mediaPath)
	if text == "" && att == nil {
		return nil
	}
	return &domain.Comment{Text: strPtr(text), Attachment: att}
}

func findParam(params []siqParam, name string) *siqParam {
	for i := range params {
		if params[i].Name == name {
			return &params[i]
		}
	}
	return nil
}

// parseSIQ unzips r, parses content.xml, uploads all referenced media to
// storage, and returns a PackDraft with fully-populated Attachment structs.
func (s *PackDraftService) parseSIQ(ctx context.Context, r io.ReaderAt, size int64) (*domain.PackDraft, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, custerr.NewBadRequestErr("invalid siq file: not a valid zip archive")
	}

	zipIndex := make(map[string]*zip.File, len(zr.File))
	for _, f := range zr.File {
		name := f.Name
		if decoded, err := url.PathUnescape(name); err == nil {
			name = decoded
		}
		zipIndex[strings.ToLower(whitespaceRe.ReplaceAllString(name, " "))] = f
	}

	xmlZipFile := zipIndex["content.xml"]
	if xmlZipFile == nil {
		return nil, custerr.NewBadRequestErr("invalid siq file: content.xml not found")
	}
	rc, err := xmlZipFile.Open()
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	defer rc.Close()

	xmlBytes, err := io.ReadAll(rc)
	if err != nil {
		return nil, custerr.NewInternalErr(err)
	}
	xmlBytes = regexp.MustCompile(`\s+xmlns(?::\w+)?="[^"]*"`).ReplaceAll(xmlBytes, nil)

	var pkg siqPackage
	if err := xml.Unmarshal(xmlBytes, &pkg); err != nil {
		return nil, custerr.NewBadRequestErr("invalid siq file: cannot parse content.xml: " + err.Error())
	}

	// processMedia uploads one zip entry to storage, probes it with ffprobe,
	// and returns a fully populated *domain.Attachment.
	// On any non-fatal error it returns nil or a best-effort result.
	processMedia := func(zipPath string) *domain.Attachment {
		f := zipIndex[strings.ToLower(zipPath)]
		if f == nil {
			log.Printf("siq import: media not found in zip: %s", zipPath)
			return nil
		}
		fr, err := f.Open()
		if err != nil {
			log.Printf("siq import: cannot open %s: %v", zipPath, err)
			return nil
		}
		data, err := io.ReadAll(fr)
		fr.Close()
		if err != nil {
			log.Printf("siq import: cannot read %s: %v", zipPath, err)
			return nil
		}

		filename := filepath.Base(f.Name)
		mimeType := mime.TypeByExtension(filepath.Ext(filename))
		key := s.attachmentService.generateKey(filename, false)

		if err := s.storage.Upload(ctx, storage.UploadInput{
			Key:      key,
			Size:     int64(len(data)),
			MimeType: mimeType,
			Reader:   bytes.NewReader(data),
		}); err != nil {
			log.Printf("siq import: failed to upload %s: %v", zipPath, err)
			return nil
		}

		tmpFile, err := os.CreateTemp("", "sgame-probe-*")
		if err != nil {
			log.Printf("siq import: failed to create temp file: %v", err)
			return &domain.Attachment{Key: key, MimeType: mimeType, Size: int64(len(data)), Type: domain.Image}
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		if _, err := io.Copy(tmpFile, bytes.NewReader(data)); err != nil {
			log.Printf("siq import: failed to buffer %s for probing: %v", zipPath, err)
			return &domain.Attachment{Key: key, MimeType: mimeType, Size: int64(len(data)), Type: domain.Image}
		}

		probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		probeData, err := ffprobe.ProbeURL(probeCtx, tmpFile.Name())
		if err != nil {
			log.Printf("siq import: failed to probe %s: %v", zipPath, err)
			return &domain.Attachment{Key: key, MimeType: mimeType, Size: int64(len(data)), Type: domain.Image}
		}

		att := &domain.Attachment{
			Key:      key,
			MimeType: mimeType,
			Size:     int64(len(data)),
			Type:     s.attachmentService.attachmentTypeFromProbe(probeData),
		}
		if att.Type == domain.Video || att.Type == domain.Audio {
			att.Duration = probeData.Format.DurationSeconds
		} else {
			att.Duration = DEFAULT_ATTACHMENT_DURATION
		}
		return att
	}

	// Collect all unique media paths across every question (both question and
	// answer/comment params) so we can deduplicate and upload in parallel.
	mediaSet := make(map[string]struct{})
	for _, r := range pkg.Rounds.Rounds {
		for _, theme := range r.Themes.Themes {
			for _, q := range theme.Questions.Questions {
				_, qMedia := extractContent(findParam(q.Params.Params, "question"))
				_, aMedia := extractContent(findParam(q.Params.Params, "answer"))
				if qMedia != "" {
					mediaSet[qMedia] = struct{}{}
				}
				if aMedia != "" {
					mediaSet[aMedia] = struct{}{}
				}
			}
		}
	}

	// Process all media concurrently using a fixed pool of 8 workers.
	type mediaResult struct {
		path string
		att  *domain.Attachment
	}
	jobs := make(chan string)
	results := make(chan mediaResult)

	for range 8 {
		go func() {
			for zipPath := range jobs {
				results <- mediaResult{zipPath, processMedia(zipPath)}
			}
		}()
	}

	go func() {
		for p := range mediaSet {
			jobs <- p
		}
		close(jobs)
	}()

	attMap := make(map[string]*domain.Attachment, len(mediaSet))
	for i := 0; i < len(mediaSet); i++ {
		r := <-results
		attMap[r.path] = r.att
	}

	lookupMedia := func(zipPath string) *domain.Attachment {
		if zipPath == "" {
			return nil
		}
		return attMap[zipPath]
	}

	var rounds []domain.Round
	var finalRound domain.FinalRound

	for _, r := range pkg.Rounds.Rounds {
		if r.Type == "final" {
			for _, theme := range r.Themes.Themes {
				if len(theme.Questions.Questions) == 0 {
					continue
				}
				q := theme.Questions.Questions[0]
				params := q.Params.Params
				qText, qMedia := extractContent(findParam(params, "question"))
				aText, aMedia := extractContent(findParam(params, "answer"))

				finalRound.Categories = append(finalRound.Categories, domain.FinalRoundCategory{
					HiddenFinalRoundCategory: domain.HiddenFinalRoundCategory{
						Name: theme.Name,
					},
					Question: domain.FinalRoundQuestion{
						HiddenFinalRoundQuestion: domain.HiddenFinalRoundQuestion{
							Category:   theme.Name,
							Text:       strPtr(qText),
							Attachment: lookupMedia(qMedia),
						},
						Answers: extractAnswers(q.Right),
						Comment: buildComment(aText, aMedia, lookupMedia),
					},
				})
			}
			continue
		}

		var categories []domain.Category
		for _, theme := range r.Themes.Themes {
			var comment *string
			if theme.Info != nil {
				comment = strPtr(theme.Info.Comments)
			}

			var questions []domain.Question
			for i, q := range theme.Questions.Questions {
				params := q.Params.Params
				qText, qMedia := extractContent(findParam(params, "question"))
				aText, aMedia := extractContent(findParam(params, "answer"))

				price := q.Price
				if q.Type == "secret" {
					if pp := findParam(params, "price"); pp != nil && pp.NumberSet != nil {
						price = pp.NumberSet.Minimum
					}
				}

				questions = append(questions, domain.Question{
					HiddenQuestion: domain.HiddenQuestion{
						Round:    r.Name,
						Category: theme.Name,
						Index:    i,
						Value:    price,
					},
					Type:       siqTypeToQuestionType(q.Type),
					Text:       strPtr(qText),
					Attachment: lookupMedia(qMedia),
					Answers:    extractAnswers(q.Right),
					Comment:    buildComment(aText, aMedia, lookupMedia),
				})
			}

			if len(questions) > 0 {
				categories = append(categories, domain.Category{
					Name:      theme.Name,
					Comment:   comment,
					Questions: questions,
				})
			}
		}

		if len(categories) > 0 {
			rounds = append(rounds, domain.Round{
				Name:       r.Name,
				Categories: categories,
			})
		}
	}

	return &domain.PackDraft{
		Name:       pkg.Name,
		Type:       domain.Private,
		Rounds:     rounds,
		FinalRound: finalRound,
	}, nil
}
