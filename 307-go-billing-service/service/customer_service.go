package service

import (
	"billing-service/model"
	"billing-service/repository"
	"time"
)

type CustomerService struct {
	customerRepo *repository.CustomerRepository
	packageRepo  *repository.PackageRepository
}

func NewCustomerService(
	customerRepo *repository.CustomerRepository,
	packageRepo *repository.PackageRepository,
) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
		packageRepo:  packageRepo,
	}
}

func (s *CustomerService) CreateCustomer(name string, pkgType model.PackageType) (*model.Customer, error) {
	pkg, err := s.packageRepo.GetByType(pkgType)
	if err != nil {
		return nil, err
	}
	if pkg == nil {
		return nil, nil
	}

	customer := &model.Customer{
		Name:            name,
		CurrentPackage:  pkgType,
		PackageStartDate: time.Now(),
	}

	if err := s.customerRepo.Create(customer); err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *CustomerService) GetCustomerByID(id uint) (*model.Customer, error) {
	return s.customerRepo.GetByID(id)
}

func (s *CustomerService) GetAllCustomers() ([]*model.Customer, error) {
	return s.customerRepo.GetAll()
}

func (s *CustomerService) ChangePackage(customerID uint, newPackage model.PackageType, changeDate time.Time) error {
	customer, err := s.customerRepo.GetByID(customerID)
	if err != nil {
		return err
	}
	if customer == nil {
		return nil
	}

	newPkg, err := s.packageRepo.GetByType(newPackage)
	if err != nil {
		return err
	}
	if newPkg == nil {
		return nil
	}

	pkgChange := &model.PackageChange{
		CustomerID: customerID,
		OldPackage: customer.CurrentPackage,
		NewPackage: newPackage,
		ChangeDate: changeDate,
	}

	if err := s.customerRepo.AddPackageChange(pkgChange); err != nil {
		return err
	}

	if err := s.customerRepo.UpdatePackage(customerID, newPackage, changeDate); err != nil {
		return err
	}

	return nil
}

func (s *CustomerService) GetPackageChanges(customerID uint, year int, month int) ([]*model.PackageChange, error) {
	return s.customerRepo.GetPackageChanges(customerID, year, month)
}
