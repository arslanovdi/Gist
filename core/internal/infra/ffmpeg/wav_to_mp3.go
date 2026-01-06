package ffmpeg

import "os/exec"

func ConvertWavToMp3(inputPath, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-q:a", "0", // лучшее качество
		"-map", "a",
		outputPath,
	)

	return cmd.Run()
}
