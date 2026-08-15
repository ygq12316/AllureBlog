package service

import (
	"blog/internal/model"
	"blog/internal/repository"
)

type VisitorService struct {
	visitorRepo  *repository.VisitorRepo
	commentRepo  *repository.CommentRepo
	danmakuRepo  *repository.DanmakuRepo
}

func NewVisitorService(vr *repository.VisitorRepo, cr *repository.CommentRepo, dr *repository.DanmakuRepo) *VisitorService {
	return &VisitorService{visitorRepo: vr, commentRepo: cr, danmakuRepo: dr}
}

func (s *VisitorService) Register(v *model.Visitor) error {
	return s.visitorRepo.CreateOrUpdate(v)
}

func (s *VisitorService) RegisterWithPassword(v *model.Visitor, password string) error {
	return s.visitorRepo.Register(v, password)
}

func (s *VisitorService) Login(username, password string) (*model.Visitor, error) {
	return s.visitorRepo.Login(username, password)
}

func (s *VisitorService) GetByUUID(uuid string) (*model.Visitor, error) {
	return s.visitorRepo.FindByUUID(uuid)
}

func (s *VisitorService) GetByUsername(username string) (*model.Visitor, error) {
	return s.visitorRepo.FindByUsername(username)
}

func (s *VisitorService) ListComments(noteID uint, limit int) ([]model.CommentWithVisitor, int64, error) {
	comments, err := s.commentRepo.ListByNote(noteID, limit)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.commentRepo.CountByNote(noteID)
	return comments, count, err
}

func (s *VisitorService) CreateComment(noteID uint, visitorUUID, content string) (*model.Comment, error) {
	c := &model.Comment{NoteID: noteID, VisitorUUID: visitorUUID, Content: content}
	err := s.commentRepo.Create(c)
	return c, err
}

// GetCommentWithVisitor 取单条评论含访客资料，供实时广播使用
func (s *VisitorService) GetCommentWithVisitor(id uint) (*model.CommentWithVisitor, error) {
	return s.commentRepo.FindByIDWithVisitor(id)
}

// VisitorExists 校验访客 UUID 是否已注册
func (s *VisitorService) VisitorExists(uuid string) bool {
	_, err := s.visitorRepo.FindByUUID(uuid)
	return err == nil
}

func (s *VisitorService) ListDanmaku(limit int) ([]model.DanmakuWithVisitor, error) {
	return s.danmakuRepo.ListRecent(limit)
}

func (s *VisitorService) CreateDanmaku(visitorUUID, content, color string) (*model.Danmaku, error) {
	d := &model.Danmaku{VisitorUUID: visitorUUID, Content: content, Color: color}
	err := s.danmakuRepo.Create(d)
	return d, err
}
