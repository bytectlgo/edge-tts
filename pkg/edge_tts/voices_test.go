package edge_tts

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestListVoices(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	voices, err := ListVoices(ctx)
	if err != nil {
		t.Fatalf("ListVoices failed: %v", err)
	}

	if len(voices) == 0 {
		t.Fatal("No voices returned")
	}

	for _, voice := range voices {
		fmt.Println(voice.ShortName)
	}

}
