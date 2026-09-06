package service

import (
	"blog/internal/model"
	"blog/internal/repository"
	"errors"
	"strings"
)

type VisitorService struct {
	visitorRepo *repository.VisitorRepo
}

func NewVisitorService(vr *repository.VisitorRepo) *VisitorService {
	return &VisitorService{visitorRepo: vr}
}

// RegisterAnonymous 匿名访客身份建档/更新；头像风格缺省为 lorelei
func (s *VisitorService) RegisterAnonymous(uuid, nickname, avatarStyle, avatarURL, signature string) (*model.Visitor, error) {
	if avatarStyle == "" {
		avatarStyle = "lorelei"
	}
	v := &model.Visitor{
		UUID:        uuid,
		Nickname:    nickname,
		AvatarStyle: avatarStyle,
		AvatarURL:   avatarURL,
		Signature:   signature,
	}
	if err := s.visitorRepo.CreateOrUpdate(v); err != nil {
		return nil, err
	}
	return v, nil
}

// RegisterWithAccount 账号注册：长度规则与重名判定都在此，
// 重名直接依赖数据库唯一索引，避免 check-then-act 竞态
func (s *VisitorService) RegisterWithAccount(uuid, username, password string) (*model.Visitor, error) {
	if len(username) < 2 || len(username) > 20 {
		return nil, &ValidationError{Message: "用户名2-20个字符"}
	}
	if len(password) < 4 {
		return nil, &ValidationError{Message: "密码至少4位"}
	}
	v := &model.Visitor{
		UUID:        uuid,
		Username:    username,
		Nickname:    username,
		AvatarStyle: "lorelei",
	}
	if err := s.visitorRepo.Register(v, password); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}
	return v, nil
}

// Login 账号登录；任何失败对调用方都是「用户名或密码错误」，不泄露细节
func (s *VisitorService) Login(username, password string) (*model.Visitor, error) {
	return s.visitorRepo.Login(username, password)
}

// GetByUUID 查访客；未找到归一为 ErrNotFound（屏蔽存储层细节）
func (s *VisitorService) GetByUUID(uuid string) (*model.Visitor, error) {
	v, err := s.visitorRepo.FindByUUID(uuid)
	if err != nil {
		return nil, ErrNotFound
	}
	return v, nil
}

// UpdateProfile 管理后台改昵称/签名：先查后改的两步操作封装在此
func (s *VisitorService) UpdateProfile(uuid, nickname, signature string) (*model.Visitor, error) {
	v, err := s.visitorRepo.FindByUUID(uuid)
	if err != nil {
		return nil, ErrNotFound
	}
	v.Nickname = nickname
	v.Signature = signature
	if err := s.visitorRepo.Update(v); err != nil {
		return nil, err
	}
	return v, nil
}

// Delete 删除访客；admin_ 前缀的账号受保护（管理员登录时会重建）
func (s *VisitorService) Delete(uuid string) error {
	if strings.HasPrefix(uuid, "admin_") {
		return ErrProtectedVisitor
	}
	return s.visitorRepo.DeleteByUUID(uuid)
}

// SyncAdmin 管理员账号种子同步（启动时以环境凭据调用）：
// 管理员与其他用户同表同登录，仅 Role 不同。失败不影响启动
func (s *VisitorService) SyncAdmin(username, password string) {
	uuid := "admin_" + username
	v, err := s.visitorRepo.FindByUUID(uuid)
	if err != nil {
		nv := &model.Visitor{
			UUID:        uuid,
			Username:    username,
			Nickname:    username,
			AvatarStyle: "lorelei",
			Role:        "admin",
		}
		s.visitorRepo.Register(nv, password)
		return
	}
	// 既有行补角色（历史数据可能为空/默认 user）
	if v.Role != "admin" {
		s.visitorRepo.UpdateRole(v.UUID, "admin")
	}
	s.visitorRepo.UpdatePassword(v.UUID, password)
}

func (s *VisitorService) ListAll() ([]model.Visitor, error) {
	return s.visitorRepo.ListAll()
}

func (s *VisitorService) CountAll() (int64, error) {
	return s.visitorRepo.CountAll()
}
