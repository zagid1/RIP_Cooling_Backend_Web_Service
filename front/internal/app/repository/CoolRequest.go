package repository

import (
	"RIP/internal/app/ds"
	"errors"
)

func (r *Repository) GetDraftCoolRequest(userID uint) (*ds.CoolRequest, error) {
	var request ds.CoolRequest

	err := r.db.Where("creator_id = ? AND status = ?", userID, ds.StatusDraft).First(&request).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *Repository) CreateCoolRequest(request *ds.CoolRequest) error {
	return r.db.Create(request).Error
}

func (r *Repository) AddComponentToCoolRequest(coolRequestID, componentID uint) error {
	var count int64

	r.db.Model(&ds.ComponentToRequest{}).Where("coolrequest_id = ? AND component_id = ?", coolRequestID, componentID).Count(&count)
	if count > 0 {
		return errors.New("components already in request")
	}

	link := ds.ComponentToRequest{
		CoolRequestID: coolRequestID,
		ComponentID:   componentID,
	}
	return r.db.Create(&link).Error
}

func (r *Repository) GetCoolRequestWithComponents(CoolRequestID uint) (*ds.CoolRequest, error) {
	var CoolRequest ds.CoolRequest

	err := r.db.Preload("ComponentLink.Component").First(&CoolRequest, CoolRequestID).Error //???????
	if err != nil {
		return nil, err
	}

	if CoolRequest.Status == ds.StatusDeleted {
		return nil, errors.New("Sorry. Page not found or has been deleted")
	}

	return &CoolRequest, nil
}

func (r *Repository) LogicallyDeleteCoolRequest(CoolRequestID uint) error {
	result := r.db.Exec("UPDATE cool_requests SET status = ? WHERE id = ?", ds.StatusDeleted, CoolRequestID)
	return result.Error
}
