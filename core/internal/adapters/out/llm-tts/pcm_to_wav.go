package llm_tts

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
)

// PCMToWAV конвертирует PCM 16-bit данные в WAV формат
func pcmToWAV(pcmData []byte, sampleRate int32, numChannels int16) ([]byte, error) {

	// WAV заголовок
	var wavData bytes.Buffer

	// RIFF chunk
	wavData.WriteString("RIFF")
	fileSize := int32(36 + len(pcmData))
	_ = binary.Write(&wavData, binary.LittleEndian, fileSize)
	wavData.WriteString("WAVE")

	// fmt subchunk
	wavData.WriteString("fmt ")
	_ = binary.Write(&wavData, binary.LittleEndian, int32(16))   // Subchunk1Size
	_ = binary.Write(&wavData, binary.LittleEndian, int16(1))    // AudioFormat (1 = PCM)
	_ = binary.Write(&wavData, binary.LittleEndian, numChannels) // NumChannels
	_ = binary.Write(&wavData, binary.LittleEndian, sampleRate)  // SampleRate

	byteRate := sampleRate * int32(numChannels) * 2 // SampleRate * NumChannels * BytesPerSample
	_ = binary.Write(&wavData, binary.LittleEndian, byteRate)

	blockAlign := numChannels * 2 // NumChannels * BytesPerSample
	_ = binary.Write(&wavData, binary.LittleEndian, blockAlign)

	_ = binary.Write(&wavData, binary.LittleEndian, int16(16)) // BitsPerSample

	// data subchunk
	wavData.WriteString("data")
	_ = binary.Write(&wavData, binary.LittleEndian, int32(len(pcmData)))
	wavData.Write(pcmData)

	return wavData.Bytes(), nil
}

// parseDataURI извлекает base64 данные из data URI
func parseDataURI(dataURI string) ([]byte, int32, error) {
	// Формат: data:audio/L16;codec=pcm;rate=24000;base64,XXXXX
	parts := strings.Split(dataURI, ",")
	if len(parts) != 2 { // Первая часть это описание формата (заголовок), вторая данные
		return nil, 0, fmt.Errorf("invalid data URI format")
	}

	// Извлекаем sample rate из header
	header := parts[0]
	var sampleRate int32 = 24000 // default

	// Вырезаем rate
	if strings.Contains(header, "rate=") {
		rateStart := strings.Index(header, "rate=") + 5
		rateEnd := strings.Index(header[rateStart:], ";")
		if rateEnd == -1 {
			rateEnd = len(header[rateStart:])
		}
		_, _ = fmt.Sscanf(header[rateStart:rateStart+rateEnd], "%d", &sampleRate)
	}

	// Декодируем base64 данные
	base64Data := parts[1]
	pcmData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, 0, err
	}

	return pcmData, sampleRate, nil
}
