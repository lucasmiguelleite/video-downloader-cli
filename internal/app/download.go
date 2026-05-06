// Package app contains the application logic for the YouTube downloader. It defines use cases that orchestrate the interactions between the downloader service and the file system.
package app

import (
	"fmt"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/fs"
)

type DownloadUseCase struct {
	service downloader.Service
}

func NewDownloadUseCase(s downloader.Service) *DownloadUseCase {
	return &DownloadUseCase{s}
}

func (uc *DownloadUseCase) Execute(url, quality string) error {
	video, err := uc.service.GetVideo(url)
	if err != nil {
		return err
	}

	data, err := uc.service.Download(video, quality)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.mp4", video.Title)
	return fs.Save(filename, data)
}
