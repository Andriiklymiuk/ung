package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AIService handles AI-powered features like PDF parsing and proposal generation
type AIService struct {
	client    *http.Client
	apiKey    string
	baseURL   string
	model     string
}

// NewAIService creates a new AI service
func NewAIService() *AIService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := os.Getenv("OPENAI_BASE_URL")
	model := os.Getenv("OPENAI_MODEL")

	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}

	return &AIService{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
	}
}

// IsConfigured returns true if AI service is properly configured
func (s *AIService) IsConfigured() bool {
	return s.apiKey != ""
}

// ProfileData represents extracted profile data from CV
type ProfileData struct {
	Name       string   `json:"name"`
	Title      string   `json:"title"`
	Bio        string   `json:"bio"`
	Skills     []string `json:"skills"`
	Experience int      `json:"experience"`
	Rate       float64  `json:"rate"`
	Location   string   `json:"location"`
	Languages  []string `json:"languages"`
	Education  []string `json:"education"`
	Projects   []string `json:"projects"`
	Links      struct {
		Github    string `json:"github"`
		LinkedIn  string `json:"linkedin"`
		Portfolio string `json:"portfolio"`
	} `json:"links"`
}

// ExtractProfileFromPDF extracts profile data from PDF text using AI
func (s *AIService) ExtractProfileFromPDF(pdfText string) (*ProfileData, error) {
	if !s.IsConfigured() {
		return s.extractProfileManually(pdfText), nil
	}

	prompt := `Extract the following information from this CV/resume and return it as JSON:
{
  "name": "Full name",
  "title": "Professional title (e.g., 'Senior Go Developer')",
  "bio": "2-3 sentence professional summary",
  "skills": ["array", "of", "technical", "skills"],
  "experience": years of experience as integer,
  "rate": suggested hourly rate in USD based on experience (0 if not determinable),
  "location": "City, Country or Remote",
  "languages": ["spoken", "languages"],
  "education": ["Degree, University", "..."],
  "projects": ["Notable project 1", "Notable project 2"],
  "links": {
    "github": "github url if found",
    "linkedin": "linkedin url if found",
    "portfolio": "portfolio url if found"
  }
}

CV/Resume text:
` + pdfText

	response, err := s.chat(prompt)
	if err != nil {
		// Fall back to manual extraction
		return s.extractProfileManually(pdfText), nil
	}

	// Parse JSON from response
	var profile ProfileData
	// Try to extract JSON from response (might be wrapped in markdown code blocks)
	jsonStr := response
	if idx := strings.Index(response, "{"); idx != -1 {
		jsonStr = response[idx:]
		if endIdx := strings.LastIndex(jsonStr, "}"); endIdx != -1 {
			jsonStr = jsonStr[:endIdx+1]
		}
	}

	if err := json.Unmarshal([]byte(jsonStr), &profile); err != nil {
		return s.extractProfileManually(pdfText), nil
	}

	return &profile, nil
}

// GenerateProposal generates a job proposal using AI
func (s *AIService) GenerateProposal(profile *ProfileData, jobTitle, jobDescription, company string) (string, error) {
	if !s.IsConfigured() {
		return s.generateProposalManually(profile, jobTitle, company), nil
	}

	prompt := fmt.Sprintf(`Generate a professional job application proposal for the following position.
Keep it concise (3-4 paragraphs), professional, and highlight relevant experience.

Profile:
- Name: %s
- Title: %s
- Skills: %s
- Experience: %d years
- Bio: %s

Job:
- Title: %s
- Company: %s
- Description: %s

Write a compelling proposal that:
1. Opens with enthusiasm for the specific role
2. Highlights 2-3 most relevant skills/experiences
3. Shows understanding of what the company needs
4. Closes with availability and next steps

Do not include subject line or signature - just the body text.`,
		profile.Name,
		profile.Title,
		strings.Join(profile.Skills, ", "),
		profile.Experience,
		profile.Bio,
		jobTitle,
		company,
		jobDescription,
	)

	return s.chat(prompt)
}

// chat sends a message to the AI and returns the response
func (s *AIService) chat(message string) (string, error) {
	requestBody := map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "user", "content": message},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI API error: %s - %s", resp.Status, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return response.Choices[0].Message.Content, nil
}

// extractProfileManually extracts profile data without AI
func (s *AIService) extractProfileManually(text string) *ProfileData {
	profile := &ProfileData{
		Experience: 0,
		Rate:       0,
	}

	// Extract email to guess name
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// First non-empty line is often the name
		if profile.Name == "" && len(line) < 50 && !strings.Contains(line, "@") {
			profile.Name = line
			continue
		}
		break
	}

	// Extract common skills
	commonSkills := []string{
		"Go", "Golang", "Python", "JavaScript", "TypeScript", "React", "Vue", "Angular",
		"Node.js", "Java", "Kotlin", "Swift", "iOS", "Android", "Rust", "C++",
		"Ruby", "Rails", "PHP", "Laravel", "Django", "Flask", "Spring", "Docker",
		"Kubernetes", "AWS", "GCP", "Azure", "PostgreSQL", "MySQL", "MongoDB",
		"Redis", "GraphQL", "REST", "API", "Microservices", "DevOps", "CI/CD",
	}

	textLower := strings.ToLower(text)
	for _, skill := range commonSkills {
		if strings.Contains(textLower, strings.ToLower(skill)) {
			profile.Skills = append(profile.Skills, skill)
		}
	}

	// Try to extract years of experience
	expPatterns := []string{
		"years of experience", "years experience", "years'", "year experience",
	}
	for _, pattern := range expPatterns {
		if idx := strings.Index(textLower, pattern); idx > 0 {
			// Look backwards for a number
			start := max(0, idx-10)
			substr := text[start:idx]
			for i := len(substr) - 1; i >= 0; i-- {
				if substr[i] >= '0' && substr[i] <= '9' {
					profile.Experience = int(substr[i] - '0')
					// Check for two digit number
					if i > 0 && substr[i-1] >= '0' && substr[i-1] <= '9' {
						profile.Experience = int(substr[i-1]-'0')*10 + profile.Experience
					}
					break
				}
			}
			break
		}
	}

	// Suggest rate based on experience
	if profile.Experience > 0 {
		profile.Rate = float64(50 + profile.Experience*10) // Base $50 + $10 per year
	}

	return profile
}

// generateProposalManually generates a basic proposal without AI
func (s *AIService) generateProposalManually(profile *ProfileData, jobTitle, company string) string {
	skills := "relevant skills"
	if len(profile.Skills) > 0 {
		skills = strings.Join(profile.Skills[:min(5, len(profile.Skills))], ", ")
	}

	return fmt.Sprintf(`Hi,

I'm excited to apply for the %s position at %s.

With %d years of experience in %s, I believe I would be a great fit for this role. %s

I'm available to start immediately and would love to discuss how I can contribute to your team.

Best regards,
%s`,
		jobTitle,
		company,
		profile.Experience,
		skills,
		profile.Bio,
		profile.Name,
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
