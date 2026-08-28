package service

import (
	"blog/internal/model"
	"blog/internal/repository"
	"strings"
)

type NoteService struct {
	noteRepo *repository.NoteRepo
}

func NewNoteService(nr *repository.NoteRepo) *NoteService {
	return &NoteService{noteRepo: nr}
}

// buildNote 组装随笔的展示字段：换行转 <br>，图片去空白且最多保留 9 张
func buildNote(content, images string) (html, imgs string) {
	html = strings.ReplaceAll(content, "\n", "<br>")
	imgList := strings.Split(images, ",")
	var clean []string
	for i, img := range imgList {
		img = strings.TrimSpace(img)
		if img != "" && i < 9 {
			clean = append(clean, img)
		}
	}
	return html, strings.Join(clean, ",")
}

func (s *NoteService) Create(content string, images string, isPublished bool) (*model.Note, error) {
	html, imgs := buildNote(content, images)
	n := &model.Note{
		Content:     content,
		HTML:        html,
		Images:      imgs,
		IsPublished: isPublished,
	}
	if err := s.noteRepo.Create(n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *NoteService) GetByID(id uint) (*model.Note, error) {
	return s.noteRepo.FindByID(id)
}

func (s *NoteService) Update(id uint, content string, images string, isPublished bool) (*model.Note, error) {
	n, err := s.noteRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	html, imgs := buildNote(content, images)
	n.Content = content
	n.HTML = html
	n.Images = imgs
	n.IsPublished = isPublished
	if err := s.noteRepo.Update(n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *NoteService) Delete(id uint) error {
	return s.noteRepo.Delete(id)
}

func (s *NoteService) ListPublished(page, pageSize int) ([]model.Note, int64, error) {
	return s.noteRepo.ListPublished(page, pageSize)
}

func (s *NoteService) ListAll(page, pageSize int) ([]model.Note, int64, error) {
	return s.noteRepo.ListAll(page, pageSize)
}

func (s *NoteService) CountAll() (int64, error) {
	return s.noteRepo.CountAll()
}
