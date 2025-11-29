package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"ung/api/internal/middleware"
	"ung/api/internal/models"
	"ung/api/internal/services"
)

// HunterController handles job hunting endpoints
type HunterController struct {
	scraperService *services.ScraperService
	aiService      *services.AIService
	pdfService     *services.PDFService
}

// NewHunterController creates a new hunter controller
func NewHunterController(uploadDir string) *HunterController {
	return &HunterController{
		scraperService: services.NewScraperService(),
		aiService:      services.NewAIService(),
		pdfService:     services.NewPDFService(uploadDir),
	}
}

// ==================== Profile Endpoints ====================

// ImportProfile handles POST /api/v1/hunter/profile/import
// Accepts PDF file upload, extracts text, and uses AI to parse profile data
func (c *HunterController) ImportProfile(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		RespondError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("cv")
	if err != nil {
		RespondError(w, "No CV file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		RespondError(w, "Only PDF files are supported", http.StatusBadRequest)
		return
	}

	// Save uploaded file
	pdfPath, err := c.pdfService.SaveUploadedFile(file, header.Filename)
	if err != nil {
		RespondError(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Extract text from PDF
	pdfText, err := c.pdfService.ExtractText(pdfPath)
	if err != nil {
		RespondError(w, "Failed to extract text from PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Use AI to extract profile data
	profileData, err := c.aiService.ExtractProfileFromPDF(pdfText)
	if err != nil {
		RespondError(w, "Failed to parse CV: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to JSON for storage
	skillsJSON, _ := json.Marshal(profileData.Skills)
	languagesJSON, _ := json.Marshal(profileData.Languages)
	educationJSON, _ := json.Marshal(profileData.Education)
	projectsJSON, _ := json.Marshal(profileData.Projects)
	linksJSON, _ := json.Marshal(profileData.Links)

	// Check if profile already exists
	var existingProfile models.Profile
	result := db.First(&existingProfile)

	profile := models.Profile{
		Name:       profileData.Name,
		Title:      profileData.Title,
		Bio:        profileData.Bio,
		Skills:     string(skillsJSON),
		Experience: profileData.Experience,
		Rate:       profileData.Rate,
		Currency:   "USD",
		Location:   profileData.Location,
		Remote:     true,
		Languages:  string(languagesJSON),
		Education:  string(educationJSON),
		Projects:   string(projectsJSON),
		Links:      string(linksJSON),
		PDFPath:    pdfPath,
		PDFContent: pdfText,
	}

	if result.Error == nil {
		// Update existing profile
		profile.ID = existingProfile.ID
		profile.CreatedAt = existingProfile.CreatedAt
		if err := db.Save(&profile).Error; err != nil {
			RespondError(w, "Failed to update profile: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Create new profile
		if err := db.Create(&profile).Error; err != nil {
			RespondError(w, "Failed to create profile: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	RespondJSON(w, map[string]interface{}{
		"message": "Profile imported successfully",
		"profile": profile,
		"extracted": profileData,
	}, http.StatusOK)
}

// GetProfile handles GET /api/v1/hunter/profile
func (c *HunterController) GetProfile(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var profile models.Profile
	if err := db.First(&profile).Error; err != nil {
		RespondError(w, "No profile found. Please import your CV first.", http.StatusNotFound)
		return
	}

	RespondJSON(w, profile, http.StatusOK)
}

// UpdateProfile handles PUT /api/v1/hunter/profile
func (c *HunterController) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var profile models.Profile
	if err := db.First(&profile).Error; err != nil {
		RespondError(w, "No profile found", http.StatusNotFound)
		return
	}

	var req struct {
		Name       string   `json:"name"`
		Title      string   `json:"title"`
		Bio        string   `json:"bio"`
		Skills     []string `json:"skills"`
		Experience int      `json:"experience"`
		Rate       float64  `json:"rate"`
		Currency   string   `json:"currency"`
		Location   string   `json:"location"`
		Remote     bool     `json:"remote"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Name != "" {
		profile.Name = req.Name
	}
	if req.Title != "" {
		profile.Title = req.Title
	}
	if req.Bio != "" {
		profile.Bio = req.Bio
	}
	if len(req.Skills) > 0 {
		skillsJSON, _ := json.Marshal(req.Skills)
		profile.Skills = string(skillsJSON)
	}
	if req.Experience > 0 {
		profile.Experience = req.Experience
	}
	if req.Rate > 0 {
		profile.Rate = req.Rate
	}
	if req.Currency != "" {
		profile.Currency = req.Currency
	}
	if req.Location != "" {
		profile.Location = req.Location
	}
	profile.Remote = req.Remote

	if err := db.Save(&profile).Error; err != nil {
		RespondError(w, "Failed to update profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, profile, http.StatusOK)
}

// ==================== Job Hunting Endpoints ====================

// Hunt handles POST /api/v1/hunter/hunt
// Scrapes jobs from configured sources and matches them to the user's profile
func (c *HunterController) Hunt(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Get user's profile
	var profile models.Profile
	if err := db.First(&profile).Error; err != nil {
		RespondError(w, "No profile found. Please import your CV first.", http.StatusNotFound)
		return
	}

	// Parse skills from profile
	var skills []string
	if profile.Skills != "" {
		json.Unmarshal([]byte(profile.Skills), &skills)
	}

	// Optional: get sources from request
	var req struct {
		Sources []string `json:"sources"` // hackernews, remoteok, etc.
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Scrape jobs
	jobs, err := c.scraperService.ScrapeJobs(skills)
	if err != nil {
		RespondError(w, "Failed to scrape jobs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save jobs to database
	for i := range jobs {
		// Check if job already exists
		var existing models.Job
		if err := db.Where("source = ? AND source_id = ?", jobs[i].Source, jobs[i].SourceID).First(&existing).Error; err == nil {
			// Update existing job
			jobs[i].ID = existing.ID
			jobs[i].CreatedAt = existing.CreatedAt
			db.Save(&jobs[i])
		} else {
			// Create new job
			db.Create(&jobs[i])
		}
	}

	// Sort by match score (highest first)
	sortJobsByMatchScore(jobs)

	RespondJSON(w, map[string]interface{}{
		"message":    "Jobs scraped successfully",
		"total":      len(jobs),
		"jobs":       jobs,
		"profile_skills": skills,
	}, http.StatusOK)
}

// ListJobs handles GET /api/v1/hunter/jobs
func (c *HunterController) ListJobs(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	// Parse query params
	source := r.URL.Query().Get("source")
	minScore := r.URL.Query().Get("min_score")

	query := db.Order("match_score DESC, created_at DESC")

	if source != "" {
		query = query.Where("source = ?", source)
	}
	if minScore != "" {
		if score, err := strconv.ParseFloat(minScore, 64); err == nil {
			query = query.Where("match_score >= ?", score)
		}
	}

	var jobs []models.Job
	if err := query.Find(&jobs).Error; err != nil {
		RespondError(w, "Failed to fetch jobs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, jobs, http.StatusOK)
}

// GetJob handles GET /api/v1/hunter/jobs/:id
func (c *HunterController) GetJob(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var job models.Job
	if err := db.First(&job, id).Error; err != nil {
		RespondError(w, "Job not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, job, http.StatusOK)
}

// DeleteJob handles DELETE /api/v1/hunter/jobs/:id
func (c *HunterController) DeleteJob(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	if err := db.Delete(&models.Job{}, id).Error; err != nil {
		RespondError(w, "Failed to delete job: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Job deleted successfully"}, http.StatusOK)
}

// ==================== Application Endpoints ====================

// CreateApplication handles POST /api/v1/hunter/applications
func (c *HunterController) CreateApplication(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var req struct {
		JobID       uint   `json:"job_id"`
		GenerateAI  bool   `json:"generate_ai"` // Use AI to generate proposal
		CoverLetter string `json:"cover_letter"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get job
	var job models.Job
	if err := db.First(&job, req.JobID).Error; err != nil {
		RespondError(w, "Job not found", http.StatusNotFound)
		return
	}

	// Get profile
	var profile models.Profile
	if err := db.First(&profile).Error; err != nil {
		RespondError(w, "No profile found", http.StatusNotFound)
		return
	}

	// Generate proposal if requested
	proposal := req.CoverLetter
	if req.GenerateAI || proposal == "" {
		var skills []string
		json.Unmarshal([]byte(profile.Skills), &skills)

		profileData := &services.ProfileData{
			Name:       profile.Name,
			Title:      profile.Title,
			Bio:        profile.Bio,
			Skills:     skills,
			Experience: profile.Experience,
		}

		generatedProposal, err := c.aiService.GenerateProposal(profileData, job.Title, job.Description, job.Company)
		if err == nil {
			proposal = generatedProposal
		}
	}

	// Create application
	application := models.Application{
		JobID:       req.JobID,
		ProfileID:   profile.ID,
		Proposal:    proposal,
		CoverLetter: req.CoverLetter,
		Status:      models.AppStatusDraft,
	}

	if err := db.Create(&application).Error; err != nil {
		RespondError(w, "Failed to create application: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Load job for response
	db.First(&application.Job, application.JobID)

	RespondJSON(w, application, http.StatusCreated)
}

// ListApplications handles GET /api/v1/hunter/applications
func (c *HunterController) ListApplications(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	status := r.URL.Query().Get("status")

	query := db.Preload("Job").Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var applications []models.Application
	if err := query.Find(&applications).Error; err != nil {
		RespondError(w, "Failed to fetch applications: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, applications, http.StatusOK)
}

// GetApplication handles GET /api/v1/hunter/applications/:id
func (c *HunterController) GetApplication(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var application models.Application
	if err := db.Preload("Job").First(&application, id).Error; err != nil {
		RespondError(w, "Application not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, application, http.StatusOK)
}

// UpdateApplication handles PUT /api/v1/hunter/applications/:id
func (c *HunterController) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	var application models.Application
	if err := db.First(&application, id).Error; err != nil {
		RespondError(w, "Application not found", http.StatusNotFound)
		return
	}

	var req struct {
		Status      string `json:"status"`
		Proposal    string `json:"proposal"`
		CoverLetter string `json:"cover_letter"`
		Notes       string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status != "" {
		application.Status = models.ApplicationStatus(req.Status)
		if req.Status == string(models.AppStatusApplied) {
			now := time.Now()
			application.AppliedAt = &now
		}
	}
	if req.Proposal != "" {
		application.Proposal = req.Proposal
	}
	if req.CoverLetter != "" {
		application.CoverLetter = req.CoverLetter
	}
	if req.Notes != "" {
		application.Notes = req.Notes
	}

	if err := db.Save(&application).Error; err != nil {
		RespondError(w, "Failed to update application: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, application, http.StatusOK)
}

// DeleteApplication handles DELETE /api/v1/hunter/applications/:id
func (c *HunterController) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	if err := db.Delete(&models.Application{}, id).Error; err != nil {
		RespondError(w, "Failed to delete application: "+err.Error(), http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Application deleted successfully"}, http.StatusOK)
}

// ==================== Stats Endpoint ====================

// GetStats handles GET /api/v1/hunter/stats
func (c *HunterController) GetStats(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var stats struct {
		TotalJobs        int64 `json:"total_jobs"`
		TotalApplications int64 `json:"total_applications"`
		Applied          int64 `json:"applied"`
		Responses        int64 `json:"responses"`
		Interviews       int64 `json:"interviews"`
		Offers           int64 `json:"offers"`
	}

	db.Model(&models.Job{}).Count(&stats.TotalJobs)
	db.Model(&models.Application{}).Count(&stats.TotalApplications)
	db.Model(&models.Application{}).Where("status = ?", models.AppStatusApplied).Count(&stats.Applied)
	db.Model(&models.Application{}).Where("status = ?", models.AppStatusResponse).Count(&stats.Responses)
	db.Model(&models.Application{}).Where("status = ?", models.AppStatusInterview).Count(&stats.Interviews)
	db.Model(&models.Application{}).Where("status = ?", models.AppStatusOffer).Count(&stats.Offers)

	RespondJSON(w, stats, http.StatusOK)
}

// Helper function to sort jobs by match score
func sortJobsByMatchScore(jobs []models.Job) {
	for i := 0; i < len(jobs)-1; i++ {
		for j := i + 1; j < len(jobs); j++ {
			if jobs[j].MatchScore > jobs[i].MatchScore {
				jobs[i], jobs[j] = jobs[j], jobs[i]
			}
		}
	}
}
