package service

import (
	"blog/internal/model"
	"blog/internal/repository"
)

// 弹幕业务常量：长度上限与 model.Danmaku 字段约束一致，颜色缺省取主题金
const (
	maxDanmakuLen       = 100
	defaultDanmakuColor = "#b8944c"
)

type DanmakuService struct {
	danmakuRepo *repository.DanmakuRepo
	visitorRepo *repository.VisitorRepo
}

func NewDanmakuService(dr *repository.DanmakuRepo, vr *repository.VisitorRepo) *DanmakuService {
	return &DanmakuService{danmakuRepo: dr, visitorRepo: vr}
}

// ListRecent 首页弹幕墙只取最近 50 条（策略收在 service，视图不关心数字）
func (s *DanmakuService) ListRecent() ([]model.DanmakuWithVisitor, error) {
	return s.danmakuRepo.ListRecent(50)
}

// Create 长度、缺省颜色、登录身份校验后落库
func (s *DanmakuService) Create(visitorUUID, content, color string) (*model.Danmaku, error) {
	if len(content) > maxDanmakuLen {
		return nil, &ValidationError{Message: "弹幕不能超过100字"}
	}
	if color == "" {
		color = defaultDanmakuColor
	}
	visitor, err := s.visitorRepo.FindByUUID(visitorUUID)
	if err != nil {
		return nil, ErrVisitorNotFound
	}
	// 仅登录账号可发弹幕（访客模式已移除）
	if visitor.Username == "" {
		return nil, &ValidationError{Message: "登录后才能发弹幕"}
	}
	d := &model.Danmaku{VisitorUUID: visitorUUID, Content: content, Color: color}
	if err := s.danmakuRepo.Create(d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DanmakuService) Delete(id string) error {
	return s.danmakuRepo.Delete(id)
}
