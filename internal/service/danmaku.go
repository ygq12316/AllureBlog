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

// Create 长度、缺省颜色、访客身份校验后落库
func (s *DanmakuService) Create(visitorUUID, content, color string) (*model.Danmaku, error) {
	if len(content) > maxDanmakuLen {
		return nil, &ValidationError{Message: "弹幕不能超过100字"}
	}
	if color == "" {
		color = defaultDanmakuColor
	}
	if _, err := s.visitorRepo.FindByUUID(visitorUUID); err != nil {
		return nil, ErrVisitorNotFound
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
