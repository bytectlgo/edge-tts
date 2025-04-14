package edge_tts

import (
	"fmt"
	"strings"
	"time"
)

// SubMaker 用于生成字幕
type SubMaker struct {
	cues []SubCue
}

// SubCue 表示一个字幕片段
type SubCue struct {
	Index int
	Start time.Duration
	End   time.Duration
	Text  string
}

// NewSubMaker 创建一个新的 SubMaker
func NewSubMaker() *SubMaker {
	return &SubMaker{
		cues: make([]SubCue, 0),
	}
}

// Feed 添加一个字幕片段
func (s *SubMaker) Feed(chunk TTSChunk) error {
	if chunk.Type != "WordBoundary" {
		return fmt.Errorf("invalid message type, expected 'WordBoundary'")
	}

	// 检查文本内容是否为空
	if chunk.Text == "" {
		return nil
	}

	// 创建新的字幕片段，注意时间单位转换为微秒
	cue := SubCue{
		Index: len(s.cues) + 1, // SRT格式要求从1开始
		Start: time.Duration(float64(chunk.Offset/10)) * time.Microsecond,
		End:   time.Duration(float64((chunk.Offset+chunk.Duration)/10)) * time.Microsecond,
		Text:  chunk.Text,
	}

	// 添加到字幕列表
	s.cues = append(s.cues, cue)
	return nil
}

// MergeCues 合并字幕片段
func (s *SubMaker) MergeCues(words int) error {
	if words <= 0 {
		return fmt.Errorf("invalid number of words to merge, expected > 0")
	}

	if len(s.cues) == 0 {
		return nil
	}

	newCues := make([]SubCue, 0)
	currentCue := s.cues[0]

	for _, cue := range s.cues[1:] {
		// 计算当前字幕中的单词数
		wordCount := len(strings.Fields(currentCue.Text))
		if wordCount < words {
			// 合并字幕
			currentCue = SubCue{
				Index: currentCue.Index,
				Start: currentCue.Start,
				End:   cue.End,
				Text:  currentCue.Text + " " + cue.Text,
			}
		} else {
			newCues = append(newCues, currentCue)
			currentCue = cue
		}
	}
	newCues = append(newCues, currentCue)
	s.cues = newCues
	return nil
}

// GetSRT 生成SRT格式的字幕
func (s *SubMaker) GetSRT() string {
	var result string
	for _, cue := range s.cues {
		// 跳过空文本的字幕
		if cue.Text == "" {
			continue
		}
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
