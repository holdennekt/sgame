package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/holdennekt/sgame/backend/internal/domain"
	"github.com/holdennekt/sgame/backend/internal/interface/storage"
	"github.com/holdennekt/sgame/backend/pkg/custerr"
)

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
	Type      string        `xml:"type,attr"`
	Items     []siqItem     `xml:"item"`
	Params    []siqParam    `xml:"param"`
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

// buildOptionsMap returns label→text for each child param of an answerOptions group.
func buildOptionsMap(param *siqParam) map[string]string {
	if param == nil {
		return nil
	}
	m := make(map[string]string, len(param.Params))
	for i := range param.Params {
		text, _ := extractContent(&param.Params[i])
		if text != "" {
			m[param.Params[i].Name] = text
		}
	}
	return m
}

// appendSelectOptions appends "A. ...\nB. ..." lines to qText in document order.
func appendSelectOptions(qText string, param *siqParam) string {
	if param == nil || len(param.Params) == 0 {
		return qText
	}
	var lines []string
	for i := range param.Params {
		text, _ := extractContent(&param.Params[i])
		if text != "" {
			lines = append(lines, param.Params[i].Name+". "+text)
		}
	}
	if len(lines) == 0 {
		return qText
	}
	if qText != "" {
		return qText + "\n" + strings.Join(lines, "\n")
	}
	return strings.Join(lines, "\n")
}

// siqPendingUpload is a media file read from the zip but not yet uploaded
type siqPendingUpload struct {
	key  string
	data []byte
}

// parseSIQ unzips r, parses content.xml, and returns a PackDraft with
// fully-populated Attachment structs (status pending) plus the still-unsent
// media bytes for the caller to upload once the draft has been persisted.
func (s *PackDraftService) parseSIQ(r io.ReaderAt, size int64) (*domain.PackDraft, []siqPendingUpload, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, nil, custerr.NewBadRequestErr("invalid siq file: not a valid zip archive")
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
		return nil, nil, custerr.NewBadRequestErr("invalid siq file: content.xml not found")
	}
	rc, err := xmlZipFile.Open()
	if err != nil {
		return nil, nil, custerr.NewInternalErr(err)
	}
	defer func() { _ = rc.Close() }()

	xmlBytes, err := io.ReadAll(rc)
	if err != nil {
		return nil, nil, custerr.NewInternalErr(err)
	}
	xmlBytes = regexp.MustCompile(`\s+xmlns(?::\w+)?="[^"]*"`).ReplaceAll(xmlBytes, nil)

	var pkg siqPackage
	if err := xml.Unmarshal(xmlBytes, &pkg); err != nil {
		return nil, nil, custerr.NewBadRequestErr("invalid siq file: cannot parse content.xml: " + err.Error())
	}

	// readMedia reads one zip entry and returns a pending *domain.Attachment
	// plus its raw bytes for later upload
	readMedia := func(zipPath string) (*domain.Attachment, []byte) {
		f := zipIndex[strings.ToLower(zipPath)]
		if f == nil {
			slog.Warn("siq import: media not found in zip", "path", zipPath)
			return nil, nil
		}
		fr, err := f.Open()
		if err != nil {
			slog.Error("siq import: cannot open media", "path", zipPath, "err", err)
			return nil, nil
		}
		data, err := io.ReadAll(fr)
		_ = fr.Close()
		if err != nil {
			slog.Error("siq import: cannot read media", "path", zipPath, "err", err)
			return nil, nil
		}

		att := &domain.Attachment{
			Key:    s.attachmentService.generateKey(false),
			Size:   int64(len(data)),
			Status: domain.AttachmentPending,
		}
		return att, data
	}

	// Collect all unique media paths
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

	// Read all media concurrently using a pool of up to 8 workers
	type mediaResult struct {
		path string
		att  *domain.Attachment
		data []byte
	}
	jobs := make(chan string)
	results := make(chan mediaResult)

	for range min(len(mediaSet), 8) {
		go func() {
			for zipPath := range jobs {
				att, data := readMedia(zipPath)
				results <- mediaResult{zipPath, att, data}
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
	var uploads []siqPendingUpload
	for i := 0; i < len(mediaSet); i++ {
		r := <-results
		attMap[r.path] = r.att
		if r.att != nil {
			uploads = append(uploads, siqPendingUpload{key: r.att.Key, data: r.data})
		}
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

				answers := extractAnswers(q.Right)
				answerTypeParam := findParam(params, "answerType")
				if answerTypeParam != nil && normalizeText(answerTypeParam.CharData) == "select" {
					optParam := findParam(params, "answerOptions")
					qText = appendSelectOptions(qText, optParam)
					optMap := buildOptionsMap(optParam)
					for j, a := range answers {
						if text, ok := optMap[a]; ok {
							answers[j] = text
						}
					}
				}

				questions = append(questions, domain.Question{
					HiddenQuestion: domain.HiddenQuestion{
						Round:    r.Name,
						Category: theme.Name,
						Index:    i,
						Value:    q.Price,
					},
					Type:       siqTypeToQuestionType(q.Type),
					Text:       strPtr(qText),
					Attachment: lookupMedia(qMedia),
					Answers:    answers,
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
	}, uploads, nil
}

// uploadSIQMedia uploads every pending SIQ media file to raw storage concurrently
func (s *PackDraftService) uploadSIQMedia(ctx context.Context, uploads []siqPendingUpload) {
	jobs := make(chan siqPendingUpload)
	var wg sync.WaitGroup
	for range min(len(uploads), 8) {
		wg.Go(func() {
			for u := range jobs {
				if err := s.rawStorage.Upload(ctx, storage.UploadInput{
					Key:    u.key,
					Size:   int64(len(u.data)),
					Reader: bytes.NewReader(u.data),
				}); err != nil {
					slog.Error("siq import: failed to upload media", "key", u.key, "err", err)
					s.attachmentService.markFailed(ctx, u.key, "failed to upload media from the imported file")
				}
			}
		})
	}
	for _, u := range uploads {
		jobs <- u
	}
	close(jobs)
	wg.Wait()
}
