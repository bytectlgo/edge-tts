package edge_tts

import (
	"fmt"
	"time"
)

// SubMaker 用于生成字幕
type SubMaker struct {
	cues []SubCue
}

// SubCue 表示一个字幕片段
type SubCue struct {
	Index   int
	Start   time.Duration
	End     time.Duration
	Text    string
	Words   []string
	WordNum int
}

// NewSubMaker 创建一个新的 SubMaker
func NewSubMaker() *SubMaker {
	return &SubMaker{
		cues: make([]SubCue, 0),
	}
}

// Feed 添加一个字幕片段
func (s *SubMaker) Feed(chunk TTSChunk) {
	if chunk.Type != "WordBoundary" {
		return
	}

	// 创建新的字幕片段
	cue := SubCue{
		Index:   len(s.cues) + 1,
		Start:   time.Duration(chunk.Offset) * time.Microsecond,
		End:     time.Duration(chunk.Offset+chunk.Duration) * time.Microsecond,
		Text:    chunk.Text,
		Words:   make([]string, 0),
		WordNum: 0,
	}

	// 添加单词
	cue.Words = append(cue.Words, chunk.Text)
	cue.WordNum++

	// 添加到字幕列表
	s.cues = append(s.cues, cue)
}

// MergeCues 合并字幕片段
func (s *SubMaker) MergeCues(wordsInCue int) {
	if wordsInCue <= 0 {
		return
	}

	merged := make([]SubCue, 0)
	current := SubCue{
		Index:   1,
		Start:   0,
		End:     0,
		Text:    "",
		Words:   make([]string, 0),
		WordNum: 0,
	}

	for _, cue := range s.cues {
		if current.WordNum == 0 {
			current.Start = cue.Start
			current.End = cue.End
			current.Text = cue.Text
			current.Words = append(current.Words, cue.Text)
			current.WordNum++
		} else if current.WordNum < wordsInCue {
			current.End = cue.End
			current.Text += " " + cue.Text
			current.Words = append(current.Words, cue.Text)
			current.WordNum++
		} else {
			merged = append(merged, current)
			current = SubCue{
				Index:   len(merged) + 1,
				Start:   cue.Start,
				End:     cue.End,
				Text:    cue.Text,
				Words:   []string{cue.Text},
				WordNum: 1,
			}
		}
	}

	if current.WordNum > 0 {
		merged = append(merged, current)
	}

	s.cues = merged
}

// GetSRT 生成SRT格式的字幕
func (s *SubMaker) GetSRT() string {
	var result string
	for _, cue := range s.cues {
		result += fmt.Sprintf("%d\n", cue.Index)
		result += fmt.Sprintf("%s --> %s\n", formatDuration(cue.Start), formatDuration(cue.End))
		result += cue.Text + "\n\n"
	}
	return result
}

// formatDuration 格式化时间持续时间
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, seconds, milliseconds)
}
