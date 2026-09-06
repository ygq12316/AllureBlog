package service

import (
	"blog/internal/model"
	"blog/internal/repository"
	"strconv"
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

// Create 校验后落库；返回带访客资料的评论与被回复人昵称（供实时广播渲染「回复@」）。
// 回复语义：parent 本身是回复时归根到其根评论，层级恒为两层。
func (s *CommentService) Create(noteID uint, visitorUUID, content string, parentID *uint) (*model.CommentWithVisitor, string, error) {
	if len(content) > maxCommentLen {
		return nil, "", &ValidationError{Message: "评论不能超过500字"}
	}
	visitor, err := s.visitorRepo.FindByUUID(visitorUUID)
	if err != nil {
		return nil, "", ErrVisitorNotFound
	}
	// 仅登录账号可评论（访客模式已移除，uuid 必为账号身份）
	if visitor.Username == "" {
		return nil, "", &ValidationError{Message: "登录后才能参与评论"}
	}

	replyTo := ""
	if parentID != nil {
		parent, err := s.commentRepo.FindByIDWithVisitor(*parentID)
		if err != nil {
			return nil, "", &ValidationError{Message: "回复的评论不存在"}
		}
		if parent.NoteID != noteID {
			return nil, "", &ValidationError{Message: "不能回复其他随笔下的评论"}
		}
		// 归根：父本身是回复时，挂到父的根上
		root := *parentID
		if parent.ParentID != nil {
			root = *parent.ParentID
		}
		parentID = &root
		replyTo = parent.Nickname
	}

	comment := &model.Comment{NoteID: noteID, VisitorUUID: visitorUUID, Content: content, ParentID: parentID}
	if err := s.commentRepo.Create(comment); err != nil {
		return nil, "", err
	}
	fresh, err := s.commentRepo.FindByIDWithVisitor(comment.ID)
	return fresh, replyTo, err
}

func (s *CommentService) Delete(id string) error {
	commentID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return &ValidationError{Message: "无效的评论 ID"}
	}
	return s.commentRepo.DeleteCascade(uint(commentID))
}

func (s *CommentService) CountAll() (int64, error) {
	return s.commentRepo.CountAll()
}
