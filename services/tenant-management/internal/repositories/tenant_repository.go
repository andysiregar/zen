package repositories

import (
	"github.com/zen/shared/pkg/models"
	"gorm.io/gorm"
)

type TenantRepository interface {
	Create(tenant *models.Tenant) error
	GetByID(id string) (*models.Tenant, error)
	GetBySlug(slug string) (*models.Tenant, error)
	Update(tenant *models.Tenant) error
	Delete(id string) error
	List(limit, offset int) ([]*models.Tenant, error)
	Count() (int64, error)
}

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(tenant *models.Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *tenantRepository) GetByID(id string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.Where("id = ?", id).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) GetBySlug(slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := r.db.Where("slug = ?", slug).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) Update(tenant *models.Tenant) error {
	return r.db.Save(tenant).Error
}

func (r *tenantRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Tenant{}).Error
}

func (r *tenantRepository) List(limit, offset int) ([]*models.Tenant, error) {
	var tenants []*models.Tenant
	err := r.db.Limit(limit).Offset(offset).Find(&tenants).Error
	return tenants, err
}

func (r *tenantRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Tenant{}).Count(&count).Error
	return count, err
}