package repository

import (
	"billing-service/model"
)

type PackageRepository struct {
	db *Database
}

func NewPackageRepository(db *Database) *PackageRepository {
	return &PackageRepository{db: db}
}

func (r *PackageRepository) GetByType(pkgType model.PackageType) (*model.PackageInfo, error) {
	query := `
		SELECT id, type, name, price_monthly, sms_quota, storage_quota
		FROM packages
		WHERE type = ?
	`

	var pkg model.PackageInfo
	var pkgTypeStr string
	err := r.db.QueryRow(query, pkgType).Scan(
		&pkg.ID,
		&pkgTypeStr,
		&pkg.Name,
		&pkg.PriceMonthly,
		&pkg.SmsQuota,
		&pkg.StorageQuota,
	)
	if err != nil {
		return nil, err
	}
	pkg.Type = model.PackageType(pkgTypeStr)

	return &pkg, nil
}

func (r *PackageRepository) GetAll() ([]*model.PackageInfo, error) {
	query := `
		SELECT id, type, name, price_monthly, sms_quota, storage_quota
		FROM packages
		ORDER BY price_monthly
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []*model.PackageInfo
	for rows.Next() {
		var pkg model.PackageInfo
		var pkgTypeStr string
		err := rows.Scan(
			&pkg.ID,
			&pkgTypeStr,
			&pkg.Name,
			&pkg.PriceMonthly,
			&pkg.SmsQuota,
			&pkg.StorageQuota,
		)
		if err != nil {
			return nil, err
		}
		pkg.Type = model.PackageType(pkgTypeStr)
		packages = append(packages, &pkg)
	}

	return packages, nil
}
