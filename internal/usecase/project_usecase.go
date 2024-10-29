// usecase/project_usecase.go
package usecase

import (
	"boonkosang/internal/domain/models"
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type ProjectUsecase interface {
	Create(ctx context.Context, req requests.CreateProjectRequest) (*responses.ProjectResponse, error)
	Update(ctx context.Context, id uuid.UUID, req requests.UpdateProjectRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*responses.ProjectResponse, error)
	List(ctx context.Context) (*responses.ProjectListResponse, error)
	Cancel(ctx context.Context, id uuid.UUID) error

	UpdateProjectStatus(ctx context.Context, req requests.UpdateProjectStatusRequest) error
	GetProjectOverview(ctx context.Context, projectID uuid.UUID) (*responses.ProjectOverviewResponse, error)

	GetProjectSummary(ctx context.Context, projectID uuid.UUID) (*responses.ProjectSummaryResponse, error)
}

type projectUsecase struct {
	projectRepo repositories.ProjectRepository
	clientRepo  repositories.ClientRepository
}

func NewProjectUsecase(projectRepo repositories.ProjectRepository, clientRepo repositories.ClientRepository) ProjectUsecase {
	return &projectUsecase{
		projectRepo: projectRepo,
		clientRepo:  clientRepo,
	}
}

func (u *projectUsecase) Create(ctx context.Context, req requests.CreateProjectRequest) (*responses.ProjectResponse, error) {
	client, err := u.clientRepo.GetByID(ctx, req.ClientID)
	if err != nil {
		return nil, errors.New("client not found")
	}

	project, err := u.projectRepo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return &responses.ProjectResponse{
		ID:          project.ProjectID,
		Name:        project.Name,
		Description: project.Description,
		Address:     project.Address,
		Status:      project.Status,
		ClientID:    project.ClientID,
		Client: &responses.ClientResponse{
			ID:      client.ClientID,
			Name:    client.Name,
			Email:   client.Email,
			Tel:     client.Tel,
			Address: client.Address,
			TaxID:   client.TaxID,
		},
		CreatedAt: project.CreatedAt,
	}, nil
}

func (u *projectUsecase) Update(ctx context.Context, id uuid.UUID, req requests.UpdateProjectRequest) error {
	_, err := u.projectRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("project not found")
	}

	return u.projectRepo.Update(ctx, id, req)

}

func (u *projectUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.projectRepo.Delete(ctx, id)
}

func (u *projectUsecase) GetByID(ctx context.Context, id uuid.UUID) (*responses.ProjectResponse, error) {
	project, client, err := u.projectRepo.GetByIDWithClient(ctx, id)
	if err != nil {
		return nil, err
	}

	return &responses.ProjectResponse{
		ID:          project.ProjectID,
		Name:        project.Name,
		Description: project.Description,
		Address:     project.Address,
		Status:      project.Status,
		ClientID:    project.ClientID,
		Client: &responses.ClientResponse{
			ID:      client.ClientID,
			Name:    client.Name,
			Email:   client.Email,
			Tel:     client.Tel,
			Address: client.Address,
			TaxID:   client.TaxID,
		},
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt.Time,
	}, nil
}

func (u *projectUsecase) List(
	ctx context.Context,
) (*responses.ProjectListResponse, error) {

	projects, err := u.projectRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	projectResponses := make([]responses.ProjectResponse, len(projects))
	for i, project := range projects {
		client, err := u.clientRepo.GetByID(ctx, project.ClientID)
		if err != nil {
			return nil, err
		}

		projectResponses[i] = responses.ProjectResponse{
			ID:          project.ProjectID,
			Name:        project.Name,
			Description: project.Description,
			Address:     project.Address,
			Status:      project.Status,
			ClientID:    project.ClientID,
			Client: &responses.ClientResponse{
				ID:      client.ClientID,
				Name:    client.Name,
				Email:   client.Email,
				Tel:     client.Tel,
				Address: client.Address,
				TaxID:   client.TaxID,
			},
			CreatedAt: project.CreatedAt,
			UpdatedAt: project.UpdatedAt.Time,
		}
	}

	return &responses.ProjectListResponse{
		Projects: projectResponses,
	}, nil
}

func (u *projectUsecase) Cancel(ctx context.Context, id uuid.UUID) error {
	return u.projectRepo.Cancel(ctx, id)
}

func (u *projectUsecase) UpdateProjectStatus(ctx context.Context, req requests.UpdateProjectStatusRequest) error {
	return u.projectRepo.UpdateStatus(ctx, req.ProjectID, req.Status)
}

func (u *projectUsecase) GetProjectOverview(ctx context.Context, projectID uuid.UUID) (*responses.ProjectOverviewResponse, error) {
	overview, err := u.projectRepo.GetProjectOverview(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project overview: %w", err)
	}

	// Calculate derived values
	taxAmount := overview.TotalSellingPrice * (overview.TaxPercentage / 100)
	totalWithTax := overview.TotalSellingPrice + taxAmount
	estimatedProfit := overview.TotalSellingPrice - overview.TotalOverallCost
	estimatedMargin := (estimatedProfit / overview.TotalSellingPrice) * 100
	actualProfit := overview.TotalSellingPrice - overview.TotalActualCost
	actualMargin := (actualProfit / overview.TotalSellingPrice) * 100

	return &responses.ProjectOverviewResponse{
		QuotationID:       overview.QuotationID.String(),
		BOQID:             overview.BOQID.String(),
		TotalOverallCost:  overview.TotalOverallCost,
		TotalSellingPrice: overview.TotalSellingPrice,
		TotalActualCost:   overview.TotalActualCost,
		TaxAmount:         taxAmount,
		TotalWithTax:      totalWithTax,
		EstimatedProfit:   estimatedProfit,
		EstimatedMargin:   estimatedMargin,
		ActualProfit:      actualProfit,
		ActualMargin:      actualMargin,
	}, nil
}

// usecase/project_summary_usecase.go

// toOverviewResponse converts ProjectOverview to ProjectOverviewResponse
func toOverviewResponse(overview *models.ProjectOverview) responses.ProjectOverviewResponse {
	// Calculate tax and totals
	taxAmount := overview.TotalSellingPrice * (overview.TaxPercentage / 100)
	totalWithTax := overview.TotalSellingPrice + taxAmount

	// Calculate profits and margins
	estimatedProfit := overview.TotalSellingPrice - overview.TotalOverallCost
	estimatedMargin := 0.0
	if overview.TotalSellingPrice > 0 {
		estimatedMargin = (estimatedProfit / overview.TotalSellingPrice) * 100
	}

	actualProfit := overview.TotalSellingPrice - overview.TotalActualCost
	actualMargin := 0.0
	if overview.TotalSellingPrice > 0 {
		actualMargin = (actualProfit / overview.TotalSellingPrice) * 100
	}

	return responses.ProjectOverviewResponse{
		QuotationID:       overview.QuotationID.String(),
		BOQID:             overview.BOQID.String(),
		TotalOverallCost:  overview.TotalOverallCost,
		TotalSellingPrice: overview.TotalSellingPrice,
		TotalActualCost:   overview.TotalActualCost,
		TaxAmount:         taxAmount,
		TotalWithTax:      totalWithTax,
		EstimatedProfit:   estimatedProfit,
		EstimatedMargin:   estimatedMargin,
		ActualProfit:      actualProfit,
		ActualMargin:      actualMargin,
	}
}

// Updated GetProjectSummary method to use the helper function
func (u *projectUsecase) GetProjectSummary(ctx context.Context, projectID uuid.UUID) (*responses.ProjectSummaryResponse, error) {
	// Get project details
	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Get summary data
	summary, err := u.projectRepo.GetProjectSummary(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Process job summaries
	var jobResponses []responses.JobSummaryResponse
	var totalStats responses.TotalStatsResponse

	for _, job := range summary.Jobs {
		estimatedMargin := 0.0
		if job.SellingPrice > 0 {
			estimatedMargin = (job.EstimatedProfit / job.SellingPrice) * 100
		}

		actualMargin := 0.0
		if job.SellingPrice > 0 {
			actualMargin = (job.ActualProfit / job.SellingPrice) * 100
		}

		// Handle valid date conversion
		var validDateStr *string
		if job.ValidDate.Valid {
			dateStr := job.ValidDate.Time.Format("2006-01-02")
			validDateStr = &dateStr
		}

		jobResponse := responses.JobSummaryResponse{
			JobName:           job.JobName,
			Unit:              job.Unit,
			Quantity:          job.Quantity,
			ValidDate:         validDateStr,
			LaborCost:         job.LaborCost,
			MaterialCost:      job.MaterialCost,
			OverallCost:       job.OverallCost,
			SellingPrice:      job.SellingPrice,
			EstimatedProfit:   job.EstimatedProfit,
			EstimatedMargin:   estimatedMargin,
			ActualOverallCost: job.ActualOverallCost,
			ActualProfit:      job.ActualProfit,
			ActualMargin:      actualMargin,
			TotalProfit:       job.TotalProfit,
			QuotationStatus:   job.QuotationStatus,
			TaxPercentage:     job.TaxPercentage,
		}

		jobResponses = append(jobResponses, jobResponse)

		// Accumulate totals
		totalStats.TotalEstimatedCost += job.OverallCost
		totalStats.TotalActualCost += job.ActualOverallCost
		totalStats.TotalSellingPrice += job.SellingPrice
		totalStats.TotalEstimatedProfit += job.EstimatedProfit
		totalStats.TotalActualProfit += job.ActualProfit
	}

	// Calculate overall statistics
	if totalStats.TotalSellingPrice > 0 {
		totalStats.EstimatedMargin = (totalStats.TotalEstimatedProfit / totalStats.TotalSellingPrice) * 100
		totalStats.ActualMargin = (totalStats.TotalActualProfit / totalStats.TotalSellingPrice) * 100
	}

	totalStats.CostVariance = totalStats.TotalEstimatedCost - totalStats.TotalActualCost
	if totalStats.TotalEstimatedCost > 0 {
		totalStats.CostVariancePercent = (totalStats.CostVariance / totalStats.TotalEstimatedCost) * 100
	}

	// Format final response
	return &responses.ProjectSummaryResponse{
		ProjectID:   project.ProjectID.String(),
		ProjectName: project.Name,
		Overview:    toOverviewResponse(&summary.ProjectOverview),
		Jobs:        jobResponses,
		TotalStats:  totalStats,
	}, nil
}
