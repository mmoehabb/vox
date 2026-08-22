package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/chasedputnam/go-kokoro-tts/pkg/api"
	"github.com/chasedputnam/go-kokoro-tts/pkg/audio"
	"github.com/chasedputnam/go-kokoro-tts/pkg/model"
	"github.com/chasedputnam/go-kokoro-tts/pkg/text"
	"github.com/chasedputnam/go-kokoro-tts/pkg/voice"
	"github.com/mmoehabb/vox/internal"
	"github.com/spf13/cobra"
)

var (
	pronounceVoiceName string
	pronounceSpeed     float64
)

var pronounceCmd = &cobra.Command{
	Use:   "pronounce <word>",
	Short: "Pronounce a word using its stored IPA via Kokoro TTS",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		word := strings.TrimSpace(args[0])
		entry, err := internal.GetCaseInsensitive(word)
		if err != nil {
			if err == internal.ErrNotFound {
				return fmt.Errorf("word not included in the dictionary: %s", word)
			}
			return err
		}

		fmt.Fprintf(os.Stderr, "Preparing TTS engine...\n")

		modelPath, err := internal.EnsureModel()
		if err != nil {
			return fmt.Errorf("model error: %w", err)
		}

		runtimeLibPath, err := internal.EnsureRuntime()
		if err != nil {
			return fmt.Errorf("runtime error: %w", err)
		}

		voiceDir, err := internal.EnsureVoices()
		if err != nil {
			return fmt.Errorf("voices error: %w", err)
		}

		voiceName := pronounceVoiceName
		if voiceName == "" {
			voiceName = "af_heart"
		}

		speed := pronounceSpeed
		if speed <= 0 {
			speed = 1.0
		}

		engineOpts := model.EngineOptions{}
		engine, err := model.NewONNXEngine(runtimeLibPath, modelPath, engineOpts)
		if err != nil {
			return fmt.Errorf("failed to initialize ONNX engine: %w", err)
		}
		defer engine.Close()

		pipeline := api.NewPipeline(
			text.NewBasicEnglishNormalizer(),
			text.NewESpeakPhonemizer(),
			text.NewTokenizer(nil),
			engine,
		)

		vLoader := voice.NewBinaryVoiceLoader(voiceDir)
		profile, err := vLoader.Load(voiceName)
		if err != nil {
			return fmt.Errorf("failed to load voice profile '%s': %w", voiceName, err)
		}

		samples, err := pipeline.Synthesize(entry.Word, "en-us", profile, float32(speed))
		if err != nil {
			return fmt.Errorf("synthesis failed: %w", err)
		}

		tmpFile, err := os.CreateTemp("", "vox-pronounce-*.wav")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		if err := audio.WriteWAV(tmpFile.Name(), samples, 24000); err != nil {
			return fmt.Errorf("failed to write WAV: %w", err)
		}

		playCmd := exec.Command("aplay", tmpFile.Name())
		playCmd.Stdout = cmd.OutOrStdout()
		playCmd.Stderr = cmd.ErrOrStderr()
		return playCmd.Run()
	},
}

func init() {
	pronounceCmd.Flags().StringVar(&pronounceVoiceName, "voice", "af_heart", "Voice profile name (default: af_heart)")
	pronounceCmd.Flags().Float64Var(&pronounceSpeed, "speed", 1.0, "Speech speed (0.5-2.0)")
}
