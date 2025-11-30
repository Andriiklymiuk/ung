package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/andriiklymiuk/ung/api/internal/models"
	"github.com/andriiklymiuk/ung/api/internal/services"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// DigController handles idea analysis endpoints
type DigController struct {
	digService *services.DigService
}

// NewDigController creates a new dig controller
func NewDigController(digService *services.DigService) *DigController {
	return &DigController{
		digService: digService,
	}
}

// StartSession starts a new idea analysis session
func (c *DigController) StartSession(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)

	var req models.DigStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Idea == "" {
		respondWithError(w, http.StatusBadRequest, "Idea is required")
		return
	}

	// Create session
	now := time.Now()
	session := models.DigSession{
		RawIdea:      req.Idea,
		Title:        c.digService.GenerateTitle(req.Idea),
		Status:       models.DigStatusAnalyzing,
		CurrentStage: string(models.DigPerspectiveFirstPrinciples),
		StartedAt:    &now,
	}

	if err := db.Create(&session).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Start async analysis
	go c.runAnalysis(db, &session)

	respondWithJSON(w, http.StatusCreated, models.StandardResponse{
		Success: true,
		Message: "Analysis started",
		Data:    session,
	})
}

// runAnalysis performs the multi-step analysis asynchronously
func (c *DigController) runAnalysis(db *gorm.DB, session *models.DigSession) {
	perspectives := c.digService.GetAnalysisPerspectives()
	var analyses []models.DigAnalysis
	stagesCompleted := []string{}

	// Run each perspective analysis
	for _, perspective := range perspectives {
		// Update current stage
		session.CurrentStage = string(perspective.Name)
		db.Save(session)

		analysis, err := c.digService.AnalyzeIdea(session.RawIdea, perspective)
		if err != nil {
			continue // Skip failed analyses
		}

		analysis.SessionID = session.ID
		if err := db.Create(analysis).Error; err != nil {
			continue
		}

		analyses = append(analyses, *analysis)
		stagesCompleted = append(stagesCompleted, string(perspective.Name))

		// Update stages completed
		stagesJSON, _ := json.Marshal(stagesCompleted)
		session.StagesCompleted = string(stagesJSON)
		db.Save(session)
	}

	// Generate execution plan
	session.CurrentStage = "execution_plan"
	db.Save(session)

	executionPlan, err := c.digService.GenerateExecutionPlan(session.RawIdea, analyses)
	if err == nil {
		executionPlan.SessionID = session.ID
		db.Create(executionPlan)
	}

	// Generate marketing materials
	session.CurrentStage = "marketing"
	db.Save(session)

	marketing, err := c.digService.GenerateMarketing(session.RawIdea, analyses)
	if err == nil {
		marketing.SessionID = session.ID
		db.Create(marketing)
	}

	// Generate revenue projections
	session.CurrentStage = "revenue"
	db.Save(session)

	revenue, err := c.digService.GenerateRevenueProjections(session.RawIdea, analyses)
	if err == nil {
		revenue.SessionID = session.ID
		db.Create(revenue)
	}

	// Generate alternatives
	session.CurrentStage = "alternatives"
	db.Save(session)

	alternatives, err := c.digService.GenerateAlternatives(session.RawIdea, analyses)
	if err == nil {
		for i := range alternatives {
			alternatives[i].SessionID = session.ID
			db.Create(&alternatives[i])
		}
	}

	// Calculate overall score and recommendation
	overallScore := c.digService.CalculateOverallScore(analyses)
	recommendation := c.digService.DetermineRecommendation(overallScore, analyses)

	// Get refined idea from first principles analysis
	for _, a := range analyses {
		if a.Perspective == models.DigPerspectiveFirstPrinciples && a.DetailedAnalysis != "" {
			var detailed map[string]interface{}
			if json.Unmarshal([]byte(a.DetailedAnalysis), &detailed) == nil {
				if refined, ok := detailed["refined_idea"].(string); ok {
					session.RefinedIdea = refined
				}
			}
		}
	}

	// Complete session
	now := time.Now()
	session.Status = models.DigStatusCompleted
	session.OverallScore = &overallScore
	session.Recommendation = recommendation
	session.CurrentStage = "completed"
	session.CompletedAt = &now
	db.Save(session)
}

// GetSession retrieves a specific session with all data
func (c *DigController) GetSession(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var session models.DigSession
	if err := db.Preload("Analyses").
		Preload("ExecutionPlan").
		Preload("Marketing").
		Preload("RevenueProjection").
		Preload("Alternatives").
		First(&session, id).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Session not found")
		return
	}

	respondWithJSON(w, http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    session,
	})
}

// GetProgress returns the progress of an ongoing analysis
func (c *DigController) GetProgress(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var session models.DigSession
	if err := db.First(&session, id).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Calculate progress percentage
	var stagesCompleted []string
	if session.StagesCompleted != "" {
		json.Unmarshal([]byte(session.StagesCompleted), &stagesCompleted)
	}

	totalStages := 9 // 5 perspectives + execution + marketing + revenue + alternatives
	progress := (len(stagesCompleted) * 100) / totalStages

	stageMessages := map[string]string{
		"first_principles": "Breaking down to first principles...",
		"designer":         "Analyzing user experience...",
		"marketing":        "Evaluating market potential...",
		"technical":        "Assessing technical feasibility...",
		"financial":        "Running financial analysis...",
		"execution_plan":   "Creating execution roadmap...",
		"revenue":          "Projecting revenue...",
		"alternatives":     "Generating alternative ideas...",
		"completed":        "Analysis complete!",
	}

	message := stageMessages[session.CurrentStage]
	if message == "" {
		message = "Analyzing..."
	}

	response := models.DigProgressResponse{
		SessionID:       session.ID,
		Status:          session.Status,
		CurrentStage:    session.CurrentStage,
		StagesCompleted: stagesCompleted,
		Progress:        progress,
		Message:         message,
	}

	respondWithJSON(w, http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    response,
	})
}

// ListSessions returns all sessions for the user
func (c *DigController) ListSessions(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)

	var sessions []models.DigSession
	if err := db.Order("created_at DESC").Find(&sessions).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch sessions")
		return
	}

	respondWithJSON(w, http.StatusOK, models.StandardResponse{
		Success: true,
		Data:    sessions,
	})
}

// DeleteSession deletes a session
func (c *DigController) DeleteSession(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Delete related data first (cascade)
	db.Where("session_id = ?", id).Delete(&models.DigAnalysis{})
	db.Where("session_id = ?", id).Delete(&models.DigExecutionPlan{})
	db.Where("session_id = ?", id).Delete(&models.DigMarketing{})
	db.Where("session_id = ?", id).Delete(&models.DigRevenueProjection{})
	db.Where("session_id = ?", id).Delete(&models.DigAlternative{})

	if err := db.Delete(&models.DigSession{}, id).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete session")
		return
	}

	respondWithJSON(w, http.StatusOK, models.StandardResponse{
		Success: true,
		Message: "Session deleted",
	})
}

// GenerateImages generates images for a session's marketing prompts
func (c *DigController) GenerateImages(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var marketing models.DigMarketing
	if err := db.Where("session_id = ?", id).First(&marketing).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Marketing data not found")
		return
	}

	// Parse imagery prompts
	var prompts []string
	if err := json.Unmarshal([]byte(marketing.ImageryPrompts), &prompts); err != nil || len(prompts) == 0 {
		respondWithError(w, http.StatusBadRequest, "No image prompts available")
		return
	}

	// Generate images for each prompt (max 3)
	var generatedImages []string
	for i, prompt := range prompts {
		if i >= 3 {
			break
		}

		imageURL, err := c.digService.GenerateImage(prompt)
		if err != nil {
			continue
		}
		generatedImages = append(generatedImages, imageURL)
	}

	// Save generated images
	imagesJSON, _ := json.Marshal(generatedImages)
	marketing.GeneratedImages = string(imagesJSON)
	db.Save(&marketing)

	respondWithJSON(w, http.StatusOK, models.StandardResponse{
		Success: true,
		Message: "Images generated",
		Data:    generatedImages,
	})
}

// ExportSession exports a session in various formats
func (c *DigController) ExportSession(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value("db").(*gorm.DB)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	var session models.DigSession
	if err := db.Preload("Analyses").
		Preload("ExecutionPlan").
		Preload("Marketing").
		Preload("RevenueProjection").
		Preload("Alternatives").
		First(&session, id).Error; err != nil {
		respondWithError(w, http.StatusNotFound, "Session not found")
		return
	}

	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=dig-analysis.json")
		json.NewEncoder(w).Encode(session)

	case "markdown":
		w.Header().Set("Content-Type", "text/markdown")
		w.Header().Set("Content-Disposition", "attachment; filename=dig-analysis.md")
		w.Write([]byte(generateMarkdownExport(session)))

	default:
		respondWithError(w, http.StatusBadRequest, "Unsupported format. Use 'json' or 'markdown'")
	}
}

// generateMarkdownExport creates a markdown report
func generateMarkdownExport(session models.DigSession) string {
	md := "# Dig Analysis Report\n\n"
	md += "## " + session.Title + "\n\n"
	md += "**Status:** " + string(session.Status) + "\n"
	if session.OverallScore != nil {
		md += "**Overall Score:** " + strconv.FormatFloat(*session.OverallScore, 'f', 1, 64) + "/100\n"
	}
	md += "**Recommendation:** " + string(session.Recommendation) + "\n\n"

	md += "### Original Idea\n" + session.RawIdea + "\n\n"

	if session.RefinedIdea != "" {
		md += "### Refined Idea\n" + session.RefinedIdea + "\n\n"
	}

	md += "---\n\n## Analysis by Perspective\n\n"

	for _, a := range session.Analyses {
		md += "### " + string(a.Perspective) + "\n"
		if a.Score != nil {
			md += "**Score:** " + strconv.FormatFloat(*a.Score, 'f', 1, 64) + "/100\n\n"
		}
		md += a.Summary + "\n\n"
	}

	if session.ExecutionPlan != nil {
		md += "---\n\n## Execution Plan\n\n"
		md += session.ExecutionPlan.Summary + "\n\n"
		md += "### MVP Scope\n" + session.ExecutionPlan.MVPScope + "\n\n"
		if session.ExecutionPlan.LLMPrompt != "" {
			md += "### LLM-Ready Prompt\n```\n" + session.ExecutionPlan.LLMPrompt + "\n```\n\n"
		}
	}

	if session.Marketing != nil {
		md += "---\n\n## Marketing\n\n"
		md += "**Value Proposition:** " + session.Marketing.ValueProposition + "\n\n"
		md += "**Elevator Pitch:** " + session.Marketing.ElevatorPitch + "\n\n"
		if session.Marketing.Taglines != "" {
			md += "**Taglines:**\n" + session.Marketing.Taglines + "\n\n"
		}
	}

	if session.RevenueProjection != nil {
		md += "---\n\n## Revenue Projections\n\n"
		md += "**Recommended Pricing:** " + session.RevenueProjection.RecommendedPrice + "\n\n"
		md += session.RevenueProjection.PricingRationale + "\n\n"
	}

	if len(session.Alternatives) > 0 {
		md += "---\n\n## Alternative Ideas\n\n"
		for i, alt := range session.Alternatives {
			md += strconv.Itoa(i+1) + ". **" + alt.AlternativeIdea + "**\n"
			md += "   - Potential: " + alt.Potential + "\n"
			md += "   - " + alt.Rationale + "\n\n"
		}
	}

	return md
}
