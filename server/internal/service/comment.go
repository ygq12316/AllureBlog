package service

import (
	"blog/internal/model"
	"blog/internal/repository"
)

// maxCommentLen 与 model.Comment.Content 的 gorm size:500 保持一致
const maxCommentLen = 500

type CommentService struct {
	commentRepo *repository.CommentRepo
	visitorRepo *repository.VisitorRepo
}

func NewCommentService(cr *repository.CommentRepo, vr *repository.VisitorRepo) *CommentService {
	return &CommentService{commentRepo: cr, visitorRepo: vr}
}

// ListByNote 随笔评论列表 + 总数（两次查询在 service 聚合，handler 不必分两步）
func (s *CommentService) ListByNote(noteID uint) ([]model.CommentWithVisitor, int64, error) {
	comments, err := s.commentRepo.ListByNote(noteID, 100)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.commentRepo.CountByNote(noteID)
	if err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

// Create 长度与访客身份校验后落库；返回带访客资料的评论供 handler 实时广播
func (s *CommentService) Create(noteID uint, visitorUUID, content string) (*model.CommentWithVisitor, error) {
	if len(content) > maxCommentLen {
		return nil, &ValidationError{Message: "评论不能超过500字"}
	}
	if _, err := s.visitorRepo.FindByUUID(visitorUUID); err != nil {
		return nil, ErrVisitorNotFound
	}
	comment := &model.Comment{NoteID: noteID, VisitorUUID: visitorUUID, Content: content}
	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}
	return s.commentRepo.FindByIDWithVisitor(comment.ID)
}

func (s *CommentService) Delete(id string) error {
	return s.commentRepo.Delete(id)
}

func (s *CommentService) CountAll() (int64, error) {
	return s.commentRepo.CountAll()
}
