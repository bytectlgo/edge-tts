package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bytectlgo/edge-tts/pkg/edge_tts"
)

func listVoices() error {
	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call ListVoices function from pkg/edge_tts
	voices, err := edge_tts.ListVoices(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get voice list: %v", err)
	}

	// Print header
	fmt.Printf("%-35s %-9s %-22s %-35s\n", "Name", "Gender", "ContentCategories", "VoicePersonalities")
	fmt.Println(strings.Repeat("-", 100))

	// Print voice list
	for _, voice := range voices {
		personalities := strings.Join(voice.StyleList, ", ")
		if personalities == "" {
			personalities = "Friendly, Positive"
		}
		fmt.Printf("%-35s %-9s %-22s %-35s\n",
			voice.ShortName,
			voice.Gender,
			"General",
			personalities)
	}
	return nil
}

func textToSpeech(text, voice, outputFile, subtitleFile string, rate, volume, pitch string) error {
	// Create new TTS configuration
	opts := []edge_tts.Option{
		edge_tts.WithRate(rate),
		edge_tts.WithVolume(volume),
		edge_tts.WithPitch(pitch),
	}

	// Create new Communicate instance
	comm := edge_tts.NewCommunicate(text, voice, opts...)

	// Set timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Save audio to file
	err := comm.Save(ctx, outputFile, subtitleFile)
	if err != nil {
		return fmt.Errorf("Failed to save audio: %v", err)
	}

	if outputFile != "" {
		fmt.Printf("Audio saved to %s\n", outputFile)
	}
	if subtitleFile != "" {
		fmt.Printf("Subtitles saved to %s\n", subtitleFile)
	}
	return nil
}

func main() {
	// Define command line parameters
	listVoicesFlag := flag.Bool("list-voices", false, "List all available voices")
	text := flag.String("text", "", "Text to convert")
	voice := flag.String("voice", "zh-CN-XiaoxiaoNeural", "Voice to use")
	outputMedia := flag.String("write-media", "", "Output audio filename")
	outputSubtitles := flag.String("write-subtitles", "", "Output subtitle filename")
	rate := flag.String("rate", "+0%", "Speech rate adjustment")
	volume := flag.String("volume", "+0%", "Volume adjustment")
	pitch := flag.String("pitch", "+0Hz", "Pitch adjustment")
	flag.Parse()

	// Execute corresponding function based on parameters
	if *listVoicesFlag {
		if err := listVoices(); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Check required parameters
	if *text == "" {
		log.Fatal("Error: --text parameter is required")
	}
	if *outputMedia == "" && *outputSubtitles == "" {
		log.Fatal("Error: --write-media or --write-subtitles parameter is required")
	}

	// Check if output files already exist
	if *outputMedia != "" {
		if _, err := os.Stat(*outputMedia); err == nil {
			fmt.Printf("Warning: File %s already exists and will be overwritten\n", *outputMedia)
		}
	}
	if *outputSubtitles != "" {
		if _, err := os.Stat(*outputSubtitles); err == nil {
			fmt.Printf("Warning: File %s already exists and will be overwritten\n", *outputSubtitles)
		}
	}

	if err := textToSpeech(*text, *voice, *outputMedia, *outputSubtitles, *rate, *volume, *pitch); err != nil {
		log.Fatal(err)
	}
}
