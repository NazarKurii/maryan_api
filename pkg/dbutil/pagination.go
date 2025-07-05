package dbutil

import (
	"context"
	"maryan_api/pkg/pagination"
	rfc7807 "maryan_api/pkg/problem"
	"math"

	"gorm.io/gorm"
)

func Pagination[T any](ctx context.Context, db *gorm.DB, cfg pagination.Cfg, preload ...string) (entities []T, totalInDB int, err error) {
	request := buildRequest(ctx, db, cfg, preload...)

	err = PossibleRawsAffectedError(request.Find(&entities), "non-existing-page")
	if err != nil {
		return nil, 0, err
	}

	totalPages, err := countTotaPages[T](db, cfg.Size)
	if err != nil {
		return nil, 0, err
	}

	return entities, totalPages, nil
}

func PaginationWithCondition[T any](ctx context.Context, db *gorm.DB, cfgCondition pagination.CfgCondition, preload ...string) (entities []T, totalInDB int, err error) {
	request := buildRequest(ctx, db, cfgCondition.Cfg, preload...)

	err = PossibleRawsAffectedError(request.Where(cfgCondition.Condition.Where, cfgCondition.Condition.Values).Find(&entities), "non-existing-page")
	if err != nil {
		return nil, 0, err
	}

	totalPages, err := countTotaPages[T](db, cfgCondition.Size)
	if err != nil {
		return nil, 0, err
	}

	return entities, totalPages, nil
}

func countTotaPages[T any](db *gorm.DB, size int) (int, error) {
	var model T
	var totalInDbINT64 int64
	err := db.Model(&model).Count(&totalInDbINT64).Error
	if err != nil || totalInDbINT64 == 0 {
		return 0, rfc7807.DB("Could not count buses.")
	}
	return int(math.Ceil(float64(totalInDbINT64) / float64(size))), nil
}

func buildRequest(ctx context.Context, db *gorm.DB, cfg pagination.Cfg, preload ...string) *gorm.DB {
	cfg.Page--

	var request = db.WithContext(ctx).Limit(cfg.Size).Offset(cfg.Page * cfg.Size).Order(cfg.Order)

	if len(preload) != 0 {
		request.Preload(preload[0])
	}

	return request
}
