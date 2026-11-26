package animals

import (
	"bytes"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
)

const (
	maxPhotoSize   = 5 * 1024 * 1024
	maxPhotoPixels = 16_000_000
)

var (
	errPhotoTooLarge   = errors.New("photo exceeds byte limit")
	errInvalidPhoto    = errors.New("Выберите корректное изображение JPEG или PNG.")
	errPhotoDimensions = errors.New("Разрешение фотографии должно быть не больше 16 мегапикселей.")
	errMultiplePhotos  = errors.New("Можно загрузить одну фотографию.")
)

type uploadedPhoto struct {
	contentType string
	data        []byte
}

func readPhoto(form *multipart.Form) (*uploadedPhoto, error) {
	for _, value := range form.Value["photo"] {
		if value != "" {
			return nil, errInvalidPhoto
		}
	}
	files := form.File["photo"]
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > 1 {
		return nil, errMultiplePhotos
	}
	fileHeader := files[0]
	if fileHeader.Size > maxPhotoSize {
		return nil, errPhotoTooLarge
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	contentType, err := validatePhoto(data)
	if err != nil {
		return nil, err
	}
	return &uploadedPhoto{contentType: contentType, data: data}, nil
}

func validatePhoto(data []byte) (string, error) {
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") {
		return "", errInvalidPhoto
	}
	// Проверяем размеры по заголовку. Деление позволяет проверить предел без переполнения.
	if config.Width < 1 || config.Height < 1 || config.Width > maxPhotoPixels/config.Height {
		return "", errPhotoDimensions
	}
	if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
		return "", errInvalidPhoto
	}
	return "image/" + format, nil
}
