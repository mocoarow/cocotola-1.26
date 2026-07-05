package gateway

import (
	"context"
	"fmt"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

// TTSClient wraps Google Cloud Text-to-Speech for the limited use case of
// turning a single utterance into an OGG/Opus byte stream.
type TTSClient struct {
	client       *texttospeech.Client
	encoding     AudioEncoding
	sampleRateHz int32
}

// NewTTSClient returns a TTS client. The caller owns the lifetime; call Close.
func NewTTSClient(ctx context.Context, encoding string, sampleRateHz int) (*TTSClient, error) {
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("new texttospeech client: %w", err)
	}
	enc, err := ParseAudioEncoding(encoding)
	if err != nil {
		return nil, err
	}
	return &TTSClient{
		client:       client,
		encoding:     enc,
		sampleRateHz: int32(sampleRateHz),
	}, nil
}

// Close releases the underlying gRPC client.
func (c *TTSClient) Close() error {
	if err := c.client.Close(); err != nil {
		return fmt.Errorf("close texttospeech client: %w", err)
	}
	return nil
}

// Synthesize renders the given text as an audio buffer.
//
// `voice` follows the Cloud TTS naming convention (e.g. "en-US-Neural2-C"),
// `lang` is the BCP-47 language code matching the voice (e.g. "en-US").
//
// Each protobuf message is allocated via new() and then populated via field
// setters so we don't have to enumerate every optional/generated field that
// exhaustruct would otherwise demand. The zero-valued fields (advanced voice
// options, prompt, pitch, speaking rate, etc.) tell Cloud TTS to use server
// defaults.
func (c *TTSClient) Synthesize(ctx context.Context, text, voice, lang string) ([]byte, error) {
	input := new(texttospeechpb.SynthesisInput)
	input.InputSource = &texttospeechpb.SynthesisInput_Text{Text: text}

	voiceParams := new(texttospeechpb.VoiceSelectionParams)
	voiceParams.LanguageCode = lang
	voiceParams.Name = voice

	audioCfg := new(texttospeechpb.AudioConfig)
	audioCfg.AudioEncoding = c.encoding.TTSEncoding()
	audioCfg.SampleRateHertz = c.sampleRateHz

	req := new(texttospeechpb.SynthesizeSpeechRequest)
	req.Input = input
	req.Voice = voiceParams
	req.AudioConfig = audioCfg

	resp, err := c.client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("synthesize speech: %w", err)
	}
	return resp.AudioContent, nil
}

