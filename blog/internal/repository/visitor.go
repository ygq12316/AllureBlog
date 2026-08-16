package repository

import (
	"blog/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type VisitorRepo struct{ db *gorm.DB }

func NewVisitorRepo(db *gorm.DB) *VisitorRepo { return &VisitorRepo{db} }

func (r *VisitorRepo) FindByUUID(uuid string) (*model.Visitor, error) {
	var v model.Visitor
	err := r.db.Where("uuid = ?", uuid).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VisitorRepo) FindByUsername(username string) (*model.Visitor, error) {
	var v model.Visitor
	err := r.db.Where("username = ?", username).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *VisitorRepo) Create(v *model.Visitor) error {
	return r.db.Create(v).Error
}

func (r *VisitorRepo) CreateOrUpdate(v *model.Visitor) error {
	existing, err := r.FindByUUID(v.UUID)
	if err != nil {
		return r.db.Create(v).Error
	}
	existing.Nickname = v.Nickname
	existing.AvatarStyle = v.AvatarStyle
	existing.AvatarURL = v.AvatarURL
	existing.Signature = v.Signature
	return r.db.Save(existing).Error
}

// Register 带密码注册
func (r *VisitorRepo) Register(v *model.Visitor, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	v.Password = string(hash)
	if v.Nickname == "" {
		v.Nickname = v.Username
	}
	if v.AvatarStyle == "" {
		v.AvatarStyle = "lorelei"
	}
	return r.db.Create(v).Error
}

// UpdatePassword 更新密码
func (r *VisitorRepo) UpdatePassword(uuid, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return r.db.Model(&model.Visitor{}).Where("uuid = ?", uuid).Update("password", string(hash)).Error
}

// Login 登录校验
func (r *VisitorRepo) Login(username, password string) (*model.Visitor, error) {
	v, err := r.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(v.Password), []byte(password)); err != nil {
		return nil, err
	}
	return v, nil
}

func (r *VisitorRepo) ListAll() ([]model.Visitor, error) {
	var visitors []model.Visitor
	err := r.db.Order("created_at DESC").Find(&visitors).Error
	return visitors, err
}

func (r *VisitorRepo) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Visitor{}).Count(&count).Error
	return count, err
}

func (r *VisitorRepo) Update(v *model.Visitor) error {
	return r.db.Save(v).Error
}

func (r *VisitorRepo) DeleteByUUID(uuid string) error {
	return r.db.Where("uuid = ?", uuid).Delete(&model.Visitor{}).Error
}
