package service

import (
	"confman/pkg/model"
	"confman/pkg/repository"
	"confman/pkg/utils"
	"gorm.io/gorm"
)

type UserService struct {
	repo *repository.Repository
	db   *gorm.DB
}

func NewUserService(repo *repository.Repository, db *gorm.DB) *UserService {
	return &UserService{repo: repo, db: db}
}

func (s *UserService) GetAll() ([]model.User, error) {
	var users []model.User
	result := s.db.Find(&users)
	return users, result.Error
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	var user model.User
	result := s.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *UserService) GetByEmail(email string) (*model.User, error) {
	var user model.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (s *UserService) Create(user *model.User) error {
	user.UUID = utils.NewUUID()
	if user.Role == "" {
		user.Role = model.RoleAuthor
	}
	return s.db.Create(user).Error
}

func (s *UserService) GetReviewers() ([]model.User, error) {
	var users []model.User
	result := s.db.Where("role = ? OR role = ? OR role = ?", model.RoleReviewer, model.RoleChair, model.RoleAdmin).Find(&users)
	return users, result.Error
}
