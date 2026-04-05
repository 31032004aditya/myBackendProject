package repository

import (
	"backend/internal/models"
	"errors"
	"sort"
	"sync"
	"time"
)

type RecordFilter struct {
	Type      string
	Category  string
	StartDate *time.Time
	EndDate   *time.Time
}

type RecordRepository interface {
	Create(record *models.Record) error
	FindByID(id uint) (*models.Record, error)
	Update(record *models.Record) error
	Delete(id uint) error
	List(filter RecordFilter) ([]models.Record, error)
	GetSummary() (*models.SummaryData, error)
	GetCategoryTotals() ([]models.CategoryTotal, error)
}

type memoryRecordRepo struct {
	mu      sync.RWMutex
	records map[uint]*models.Record
	nextID  uint
}

func NewRecordRepository() RecordRepository {
	return &memoryRecordRepo{
		records: make(map[uint]*models.Record),
		nextID:  1,
	}
}

func (r *memoryRecordRepo) Create(record *models.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	record.ID = r.nextID
	r.nextID++
	r.records[record.ID] = record
	return nil
}

func (r *memoryRecordRepo) FindByID(id uint) (*models.Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if record, exists := r.records[id]; exists {
		rec := *record
		return &rec, nil
	}
	return nil, errors.New("record not found")
}

func (r *memoryRecordRepo) Update(record *models.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.records[record.ID]; exists {
		r.records[record.ID] = record
		return nil
	}
	return errors.New("record not found")
}

func (r *memoryRecordRepo) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.records[id]; exists {
		delete(r.records, id)
		return nil
	}
	return errors.New("record not found")
}

func (r *memoryRecordRepo) List(filter RecordFilter) ([]models.Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []models.Record
	for _, rec := range r.records {
		if filter.Type != "" && rec.Type != filter.Type {
			continue
		}
		if filter.Category != "" && rec.Category != filter.Category {
			continue
		}
		if filter.StartDate != nil && rec.Date.Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && rec.Date.After(*filter.EndDate) {
			continue
		}
		result = append(result, *rec)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Date.After(result[j].Date)
	})

	return result, nil
}

func (r *memoryRecordRepo) GetSummary() (*models.SummaryData, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var income, expense float64
	for _, rec := range r.records {
		if rec.Type == "income" {
			income += rec.Amount
		} else if rec.Type == "expense" {
			expense += rec.Amount
		}
	}

	return &models.SummaryData{
		TotalIncome:  income,
		TotalExpense: expense,
		NetBalance:   income - expense,
	}, nil
}

func (r *memoryRecordRepo) GetCategoryTotals() ([]models.CategoryTotal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalsMap := make(map[string]float64)
	for _, rec := range r.records {
		totalsMap[rec.Category] += rec.Amount
	}

	var totals []models.CategoryTotal
	for cat, tot := range totalsMap {
		totals = append(totals, models.CategoryTotal{
			Category: cat,
			Total:    tot,
		})
	}
	return totals, nil
}
