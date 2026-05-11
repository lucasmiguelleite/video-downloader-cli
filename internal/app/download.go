package app

import (
	"fmt"
	"path/filepath"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/fs"
)

type DownloadUseCase struct {
	service downloader.Service
	dir     string
}

func NewDownloadUseCase(s downloader.Service) *DownloadUseCase {
	return &DownloadUseCase{service: s}
}

func NewDownloadUseCaseWithDir(s downloader.Service, dir string) *DownloadUseCase {
	return &DownloadUseCase{service: s, dir: dir}
}

func (uc *DownloadUseCase) Execute(url, quality string) (string, error) {
	video, err := uc.service.GetVideo(url)
	if err != nil {
		return "", err
	}

	data, err := uc.service.Download(video, quality)
	if err != nil {
		return "", err
	}

	dir := uc.dir
	if dir == "" {
		dir, err = fs.DefaultDownloadDir()
		if err != nil {
			return "", err
		}
	}

	filename := fmt.Sprintf("%s.mp4", video.Title)
	if err := fs.Save(dir, filename, data); err != nil {
		return "", err
	}

	return filepath.Join(dir, filename), nil
}
