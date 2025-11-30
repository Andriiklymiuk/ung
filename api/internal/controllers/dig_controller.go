package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ung/api/internal/middleware"
	"ung/api/internal/models"
	"ung/api/internal/services"

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
	db := middleware.GetTenantDB(r)

	var req models.DigStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Idea == "" {
		RespondError(w, "Idea is required", http.StatusBadRequest)
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
		RespondError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Start async analysis
	go c.runAnalysis(db, &session)

	RespondJSON(w, session, http.StatusCreated)
}

// runAnalysis performs the multi-step analysis asynchronously with agentic early-exit logic
func (c *DigController) runAnalysis(db *gorm.DB, session *models.DigSession) {
	perspectives := c.digService.GetAnalysisPerspectives()
	var analyses []models.DigAnalysis
	stagesCompleted := []string{}

	// ========================================
	// Phase 1: Run core perspective analyses
	// ========================================
	for _, perspective := range perspectives {
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

		stagesJSON, _ := json.Marshal(stagesCompleted)
		session.StagesCompleted = string(stagesJSON)
		db.Save(session)
	}

	// Calculate initial score
	initialScore := c.digService.CalculateOverallScore(analyses)

	// ========================================
	// AGENTIC GATE: Viability Check for Low-Scoring Ideas
	// ========================================
	// If initial score is concerning (<45), run a viability check to determine:
	// - Is this fundamentally flawed with no path forward? → Stop early
	// - Is this salvageable with pivots? → Focus on alternatives
	// - Is there potential here? → Continue full analysis
	var shouldContinueHarsh = true
	var focusOnPivots = false

	if initialScore < 45 {
		session.CurrentStage = "viability_check"
		db.Save(session)

		viability, err := c.digService.EvaluateViability(session.RawIdea, analyses, initialScore)
		if err == nil {
			// Store viability check result
			session.ViabilityCheck = viability.FullAnalysisJSON
			session.FlawType = viability.FlawType

			if !viability.ShouldContinue {
				// EARLY EXIT: Idea is fundamentally flawed with no viable path
				session.EarlyExit = true
				session.EarlyExitReason = viability.Reasoning
				shouldContinueHarsh = false
				focusOnPivots = true // Still generate alternatives to help them find a better direction

				// Set recommendation immediately
				session.Recommendation = models.DigRecommendAbandon
			} else if viability.FocusOnPivots {
				// PIVOT MODE: Continue but focus on finding better directions
				focusOnPivots = true
				session.PivotFocus = true
				// Run reduced harsh analysis - just devil's advocate to surface key issues
				shouldContinueHarsh = true // But we'll limit which perspectives we run
			}

			stagesCompleted = append(stagesCompleted, "viability_check")
			stagesJSON, _ := json.Marshal(stagesCompleted)
			session.StagesCompleted = string(stagesJSON)
			db.Save(session)
		}
	}

	// ========================================
	// Phase 2: Run harsh/critical perspectives
	// ========================================
	// Only run if idea passed viability gate or we're doing pivot-focused analysis
	if shouldContinueHarsh && !session.EarlyExit {
		harshPerspectives := c.digService.GetHarshPerspectives()

		// If focusing on pivots, only run devil's advocate and copycat analysis
		// to identify what's wrong and what similar ideas succeeded
		if focusOnPivots {
			limitedPerspectives := []services.AnalysisPerspective{}
			for _, p := range harshPerspectives {
				if p.Name == models.DigPerspectiveDevilsAdvocate || p.Name == models.DigPerspectiveCopycat {
					limitedPerspectives = append(limitedPerspectives, p)
				}
			}
			harshPerspectives = limitedPerspectives
		}

		for _, perspective := range harshPerspectives {
			session.CurrentStage = string(perspective.Name)
			db.Save(session)

			analysis, err := c.digService.AnalyzeIdea(session.RawIdea, perspective)
			if err != nil {
				continue
			}

			analysis.SessionID = session.ID
			if err := db.Create(analysis).Error; err != nil {
				continue
			}

			analyses = append(analyses, *analysis)
			stagesCompleted = append(stagesCompleted, string(perspective.Name))

			stagesJSON, _ := json.Marshal(stagesCompleted)
			session.StagesCompleted = string(stagesJSON)
			db.Save(session)
		}
	}

	// ========================================
	// Phase 3: Generate outputs (only for viable ideas)
	// ========================================
	// Skip execution/marketing/revenue if:
	// - Early exit (fundamentally flawed)
	// - Pivot focus (need to find better direction first)
	// - Very low score (<40)
	if initialScore >= 40 && !session.EarlyExit && !focusOnPivots {
		// Generate execution plan
		session.CurrentStage = "execution_plan"
		db.Save(session)

		executionPlan, err := c.digService.GenerateExecutionPlan(session.RawIdea, analyses)
		if err == nil {
			executionPlan.SessionID = session.ID
			db.Create(executionPlan)
		}
		stagesCompleted = append(stagesCompleted, "execution_plan")

		// Generate marketing materials
		session.CurrentStage = "marketing"
		db.Save(session)

		marketing, err := c.digService.GenerateMarketing(session.RawIdea, analyses)
		if err == nil {
			marketing.SessionID = session.ID
			db.Create(marketing)
		}
		stagesCompleted = append(stagesCompleted, "marketing")

		// Generate revenue projections
		session.CurrentStage = "revenue"
		db.Save(session)

		revenue, err := c.digService.GenerateRevenueProjections(session.RawIdea, analyses)
		if err == nil {
			revenue.SessionID = session.ID
			db.Create(revenue)
		}
		stagesCompleted = append(stagesCompleted, "revenue")
	}

	// ========================================
	// Always generate alternatives
	// ========================================
	// This is ESPECIALLY important for early-exit and pivot-focus ideas
	// to help founders find better directions
	session.CurrentStage = "alternatives"
	db.Save(session)

	alternatives, err := c.digService.GenerateAlternatives(session.RawIdea, analyses)
	if err == nil {
		for i := range alternatives {
			alternatives[i].SessionID = session.ID
			db.Create(&alternatives[i])
		}
	}
	stagesCompleted = append(stagesCompleted, "alternatives")

	// ========================================
	// Finalize session
	// ========================================
	stagesJSON, _ := json.Marshal(stagesCompleted)
	session.StagesCompleted = string(stagesJSON)

	// Calculate overall score and recommendation (if not already set by early exit)
	overallScore := c.digService.CalculateOverallScore(analyses)
	if session.Recommendation == "" {
		session.Recommendation = c.digService.DetermineRecommendation(overallScore, analyses)
	}

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
	session.CurrentStage = "completed"
	session.CompletedAt = &now
	db.Save(session)
}

// GetSession retrieves a specific session with all data
func (c *DigController) GetSession(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		RespondError(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var session models.DigSession
	if err := db.Preload("Analyses").
		Preload("ExecutionPlan").
		Preload("Marketing").
		Preload("RevenueProjection").
		Preload("Alternatives").
		First(&session, id).Error; err != nil {
		RespondError(w, "Session not found", http.StatusNotFound)
		return
	}

	RespondJSON(w, session, http.StatusOK)
}

// GetProgress returns the progress of an ongoing analysis
func (c *DigController) GetProgress(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		RespondError(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var session models.DigSession
	if err := db.First(&session, id).Error; err != nil {
		RespondError(w, "Session not found", http.StatusNotFound)
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
		"viability_check":  "Running viability gate check...",
		"devils_advocate":  "Playing devil's advocate...",
		"copycat":          "Analyzing copycat potential...",
		"user_psychology":  "Evaluating user psychology...",
		"scalability":      "Stress testing scalability...",
		"worst_case":       "Mapping worst case scenarios...",
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

	RespondJSON(w, response, http.StatusOK)
}

// ListSessions returns all sessions for the user
func (c *DigController) ListSessions(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	var sessions []models.DigSession
	if err := db.Order("created_at DESC").Find(&sessions).Error; err != nil {
		RespondError(w, "Failed to fetch sessions", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, sessions, http.StatusOK)
}

// DeleteSession deletes a session
func (c *DigController) DeleteSession(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		RespondError(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Delete related data first (cascade)
	db.Where("session_id = ?", id).Delete(&models.DigAnalysis{})
	db.Where("session_id = ?", id).Delete(&models.DigExecutionPlan{})
	db.Where("session_id = ?", id).Delete(&models.DigMarketing{})
	db.Where("session_id = ?", id).Delete(&models.DigRevenueProjection{})
	db.Where("session_id = ?", id).Delete(&models.DigAlternative{})

	if err := db.Delete(&models.DigSession{}, id).Error; err != nil {
		RespondError(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}

	RespondJSON(w, map[string]string{"message": "Session deleted"}, http.StatusOK)
}

// GenerateImages generates images for a session's marketing prompts
func (c *DigController) GenerateImages(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		RespondError(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	var marketing models.DigMarketing
	if err := db.Where("session_id = ?", id).First(&marketing).Error; err != nil {
		RespondError(w, "Marketing data not found", http.StatusNotFound)
		return
	}

	// Parse imagery prompts
	var prompts []string
	if err := json.Unmarshal([]byte(marketing.ImageryPrompts), &prompts); err != nil || len(prompts) == 0 {
		RespondError(w, "No image prompts available", http.StatusBadRequest)
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

	RespondJSON(w, generatedImages, http.StatusOK)
}

// ExportSession exports a session in various formats
func (c *DigController) ExportSession(w http.ResponseWriter, r *http.Request) {
	db := middleware.GetTenantDB(r)

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		RespondError(w, "Invalid session ID", http.StatusBadRequest)
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
		RespondError(w, "Session not found", http.StatusNotFound)
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
		RespondError(w, "Unsupported format. Use 'json' or 'markdown'", http.StatusBadRequest)
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
