package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/andriiklymiuk/ung/api/internal/models"
)

// DigService handles idea analysis and incubation
type DigService struct {
	aiService *AIService
}

// NewDigService creates a new dig service
func NewDigService(aiService *AIService) *DigService {
	return &DigService{
		aiService: aiService,
	}
}

// AnalysisPerspective defines a perspective for analysis
type AnalysisPerspective struct {
	Name        models.DigPerspective
	Description string
	Prompt      string
}

// GetAnalysisPerspectives returns all analysis perspectives
func (s *DigService) GetAnalysisPerspectives() []AnalysisPerspective {
	return []AnalysisPerspective{
		{
			Name:        models.DigPerspectiveFirstPrinciples,
			Description: "Elon Musk-style first principles thinking",
			Prompt: `You are analyzing an idea using FIRST PRINCIPLES thinking (Elon Musk style).

Break down the idea to its fundamental truths and reason up from there.
Ask: What are we absolutely sure is true? What are the physics/economics/human behavior fundamentals?

For the idea: "%s"

Provide analysis in this JSON format:
{
  "summary": "2-3 sentence first principles analysis",
  "fundamental_truths": ["list of undeniable truths this idea relies on"],
  "assumptions_challenged": ["common assumptions that might be wrong"],
  "strengths": ["what fundamentally works"],
  "weaknesses": ["what violates first principles"],
  "opportunities": ["what first principles reveal as possible"],
  "threats": ["fundamental risks"],
  "recommendations": ["specific actions based on first principles"],
  "score": 0-100,
  "verdict": "proceed/pivot/refine/abandon",
  "refined_idea": "improved version of the idea based on first principles"
}`,
		},
		{
			Name:        models.DigPerspectiveDesigner,
			Description: "UX/Product Designer perspective",
			Prompt: `You are analyzing an idea as a SENIOR PRODUCT DESIGNER.

Focus on: User experience, pain points, user journey, emotional design, accessibility, delight factors.

For the idea: "%s"

Provide analysis in this JSON format:
{
  "summary": "2-3 sentence design perspective",
  "target_users": ["specific user personas"],
  "pain_points_addressed": ["user problems this solves"],
  "user_journey": "brief description of ideal user flow",
  "ux_strengths": ["what will delight users"],
  "ux_weaknesses": ["friction points and concerns"],
  "design_opportunities": ["ways to improve user experience"],
  "accessibility_considerations": ["inclusivity factors"],
  "recommendations": ["specific design recommendations"],
  "score": 0-100,
  "key_screens": ["essential screens/features needed"]
}`,
		},
		{
			Name:        models.DigPerspectiveMarketing,
			Description: "Marketing team perspective",
			Prompt: `You are analyzing an idea as a CHIEF MARKETING OFFICER.

Focus on: Market positioning, target audience, messaging, competitive landscape, go-to-market strategy.

For the idea: "%s"

Provide analysis in this JSON format:
{
  "summary": "2-3 sentence marketing analysis",
  "target_market": "primary market segment",
  "market_size_estimate": "rough TAM/SAM/SOM",
  "positioning": "unique value proposition",
  "competitive_advantages": ["differentiators"],
  "competitive_threats": ["market challenges"],
  "messaging_angles": ["key marketing messages that would resonate"],
  "channels": ["recommended marketing channels"],
  "viral_potential": "low/medium/high with reasoning",
  "recommendations": ["go-to-market recommendations"],
  "score": 0-100,
  "tagline_options": ["3-5 potential taglines"]
}`,
		},
		{
			Name:        models.DigPerspectiveTechnical,
			Description: "Technical architect perspective",
			Prompt: `You are analyzing an idea as a SENIOR TECHNICAL ARCHITECT.

Focus on: Technical feasibility, architecture, scalability, security, build vs buy decisions.

For the idea: "%s"

Provide analysis in this JSON format:
{
  "summary": "2-3 sentence technical analysis",
  "technical_feasibility": "straightforward/moderate/complex/breakthrough",
  "core_components": ["main technical building blocks needed"],
  "recommended_stack": {
    "frontend": "recommended tech",
    "backend": "recommended tech",
    "database": "recommended tech",
    "infrastructure": "recommended approach"
  },
  "build_vs_buy": ["what to build custom vs use existing solutions"],
  "technical_risks": ["implementation challenges"],
  "scalability_approach": "how it would scale",
  "security_considerations": ["security requirements"],
  "mvp_timeline": "realistic MVP timeline",
  "recommendations": ["technical recommendations"],
  "score": 0-100,
  "integrations_needed": ["third-party services/APIs needed"]
}`,
		},
		{
			Name:        models.DigPerspectiveFinancial,
			Description: "Financial analyst perspective",
			Prompt: `You are analyzing an idea as a FINANCIAL ANALYST / VC.

Focus on: Revenue model, unit economics, funding requirements, ROI, market opportunity.

For the idea: "%s"

Provide analysis in this JSON format:
{
  "summary": "2-3 sentence financial analysis",
  "revenue_models": ["viable monetization strategies"],
  "recommended_pricing": "suggested pricing approach with reasoning",
  "unit_economics": {
    "estimated_cac": "customer acquisition cost estimate",
    "estimated_ltv": "lifetime value estimate",
    "ltv_cac_ratio": "estimated ratio"
  },
  "startup_costs": "estimated initial investment needed",
  "monthly_burn": "estimated monthly operating costs",
  "break_even": "when profitability is realistic",
  "funding_recommendation": "bootstrap/angel/seed/series with reasoning",
  "financial_risks": ["financial challenges"],
  "revenue_projections": {
    "year1": "conservative estimate",
    "year2": "growth estimate",
    "year3": "scale estimate"
  },
  "recommendations": ["financial recommendations"],
  "score": 0-100
}`,
		},
	}
}

// GetHarshPerspectives returns harsh/critical analysis perspectives for deeper analysis
func (s *DigService) GetHarshPerspectives() []AnalysisPerspective {
	return []AnalysisPerspective{
		{
			Name:        models.DigPerspectiveDevilsAdvocate,
			Description: "Devil's Advocate - Harsh critic",
			Prompt: `You are a BRUTAL DEVIL'S ADVOCATE. Your job is to DESTROY this idea.

Be harsh. Be critical. Find EVERY flaw. Assume the worst. Play the skeptic.
Your goal: If this idea can survive your criticism, it might actually work.

For the idea: "%s"

Attack from every angle:
- Why will this FAIL?
- What are they NOT seeing?
- Why won't customers care?
- Why is the timing wrong?
- What will kill this before it starts?

Provide analysis in this JSON format:
{
  "summary": "2-3 sentence brutal assessment",
  "fatal_flaws": ["reasons this will definitely fail"],
  "blind_spots": ["things the founder isn't seeing"],
  "customer_objections": ["why customers won't buy"],
  "timing_problems": ["why the timing is wrong"],
  "competition_will_crush": ["how competitors will destroy this"],
  "execution_nightmares": ["implementation problems that will kill it"],
  "weaknesses": ["consolidated list of all weaknesses"],
  "survival_requirements": ["what MUST be true for this to work"],
  "recommendations": ["harsh but actionable advice"],
  "score": 0-100,
  "verdict": "This idea will fail because..."
}`,
		},
		{
			Name:        models.DigPerspectiveCopycat,
			Description: "Copycat Analysis - Can copycats succeed?",
			Prompt: `You are analyzing this idea from a COPYCAT perspective.

Key question: If someone copies this idea tomorrow, can THEY still succeed?
This reveals how defensible the idea is and if there's room in the market.

For the idea: "%s"

Consider:
- How easy is it to copy?
- Can a well-funded competitor crush the original?
- Is this a "winner takes all" market?
- What would a copycat need to succeed?
- Are there successful copycats in similar markets?

Provide analysis in this JSON format:
{
  "summary": "2-3 sentence copycat viability assessment",
  "copy_difficulty": "trivial/easy/moderate/hard/very_hard",
  "time_to_copy": "how long for a competitor to replicate",
  "moat_analysis": {
    "network_effects": "none/weak/moderate/strong",
    "switching_costs": "none/low/medium/high",
    "brand_value": "none/low/medium/high",
    "data_advantage": "none/low/medium/high",
    "economies_of_scale": "none/low/medium/high"
  },
  "copycat_success_chance": "0-100% chance a copycat succeeds",
  "successful_copycat_examples": ["examples of copycats that won"],
  "market_room": "is there room for multiple players?",
  "first_mover_advantage": "low/medium/high with reasoning",
  "weaknesses": ["defensibility weaknesses"],
  "strengths": ["what makes this hard to copy"],
  "recommendations": ["how to become uncopyable"],
  "score": 0-100
}`,
		},
		{
			Name:        models.DigPerspectiveUserPsychology,
			Description: "User Psychology - Behavioral analysis",
			Prompt: `You are a BEHAVIORAL PSYCHOLOGIST analyzing this idea.

Focus on: Human behavior, habits, motivation, cognitive biases, behavior change.
Ask: Will people ACTUALLY do this? Not "would they say yes in a survey" but WILL THEY?

For the idea: "%s"

Analyze:
- What behavior change is required?
- What habits need to form or break?
- What's the motivation (pain vs gain)?
- What friction will stop adoption?
- Does this align with how humans actually behave?

Provide analysis in this JSON format:
{
  "summary": "2-3 sentence psychological analysis",
  "behavior_change_required": "none/minor/moderate/major/extreme",
  "habit_formation": {
    "trigger": "what triggers the behavior",
    "action": "what the user must do",
    "reward": "what reward reinforces it",
    "difficulty": "how hard to form this habit"
  },
  "motivation_type": "pain_avoidance/pleasure_seeking/social/aspirational",
  "psychological_barriers": ["mental obstacles to adoption"],
  "cognitive_biases_helping": ["biases that help adoption"],
  "cognitive_biases_hurting": ["biases that hurt adoption"],
  "addiction_potential": "will users come back naturally?",
  "social_proof_dependency": "how much does this need social proof?",
  "weaknesses": ["psychological weaknesses"],
  "strengths": ["psychological strengths"],
  "recommendations": ["how to align with human psychology"],
  "score": 0-100,
  "will_users_actually": "honest assessment of real behavior"
}`,
		},
		{
			Name:        models.DigPerspectiveScalability,
			Description: "Scalability Stress Test",
			Prompt: `You are running a SCALABILITY STRESS TEST on this idea.

Assume it works. Assume people love it. Now: CAN IT SCALE?
10x users. 100x users. Global expansion. What breaks?

For the idea: "%s"

Stress test:
- What happens at 10x scale?
- What happens at 100x scale?
- What's the unit economics at scale?
- What operational nightmares emerge?
- What regulatory/legal issues appear at scale?

Provide analysis in this JSON format:
{
  "summary": "2-3 sentence scalability assessment",
  "current_model_max": "maximum scale with current approach",
  "scale_10x": {
    "breaks": ["what breaks at 10x"],
    "cost_change": "how costs change",
    "quality_impact": "impact on quality"
  },
  "scale_100x": {
    "breaks": ["what breaks at 100x"],
    "cost_change": "how costs change",
    "requires": ["what's needed to reach this"]
  },
  "global_expansion_issues": ["international scaling problems"],
  "operational_bottlenecks": ["what operations can't scale"],
  "regulatory_at_scale": ["legal issues that appear at scale"],
  "unit_economics_at_scale": "do margins improve or collapse?",
  "weaknesses": ["scalability weaknesses"],
  "strengths": ["scalability strengths"],
  "recommendations": ["how to build for scale from day 1"],
  "score": 0-100,
  "scale_ceiling": "realistic maximum scale"
}`,
		},
		{
			Name:        models.DigPerspectiveWorstCase,
			Description: "Worst Case Scenario",
			Prompt: `You are a WORST CASE SCENARIO analyst.

Murphy's Law: Everything that can go wrong, will go wrong.
Your job: Map out exactly how this fails in the worst possible way.

For the idea: "%s"

Consider:
- Technical disasters (security breaches, outages, data loss)
- Market disasters (competitor launches, market crashes, regulation)
- Operational disasters (key person leaves, supplier fails, PR crisis)
- Financial disasters (funding dries up, costs explode, no revenue)
- Legal disasters (lawsuits, regulation, IP issues)

Provide analysis in this JSON format:
{
  "summary": "2-3 sentence worst case overview",
  "nightmare_scenarios": [
    {
      "scenario": "what happens",
      "probability": "low/medium/high",
      "impact": "minor/major/catastrophic",
      "recovery": "impossible/hard/possible/easy"
    }
  ],
  "single_points_of_failure": ["things that could kill everything"],
  "black_swan_risks": ["unlikely but devastating events"],
  "cascading_failures": ["how one problem triggers others"],
  "weaknesses": ["vulnerability points"],
  "survival_playbook": ["how to survive worst cases"],
  "insurance_options": ["ways to protect against disaster"],
  "recommendations": ["how to de-risk"],
  "score": 0-100,
  "existential_risk": "low/medium/high - chance of total failure"
}`,
		},
	}
}

// AnalyzeIdea performs a single perspective analysis
func (s *DigService) AnalyzeIdea(idea string, perspective AnalysisPerspective) (*models.DigAnalysis, error) {
	prompt := fmt.Sprintf(perspective.Prompt, idea)

	response, err := s.aiService.chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze with %s perspective: %w", perspective.Name, err)
	}

	// Parse JSON response
	jsonStr := extractJSON(response)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// Return partial analysis if JSON parsing fails
		return &models.DigAnalysis{
			Perspective: perspective.Name,
			Summary:     response,
		}, nil
	}

	analysis := &models.DigAnalysis{
		Perspective: perspective.Name,
	}

	// Extract fields
	if summary, ok := result["summary"].(string); ok {
		analysis.Summary = summary
	}
	if score, ok := result["score"].(float64); ok {
		analysis.Score = &score
	}

	// Convert arrays to JSON strings
	if strengths, ok := result["strengths"].([]interface{}); ok {
		analysis.Strengths = toJSONString(strengths)
	} else if strengths, ok := result["ux_strengths"].([]interface{}); ok {
		analysis.Strengths = toJSONString(strengths)
	} else if strengths, ok := result["competitive_advantages"].([]interface{}); ok {
		analysis.Strengths = toJSONString(strengths)
	}

	if weaknesses, ok := result["weaknesses"].([]interface{}); ok {
		analysis.Weaknesses = toJSONString(weaknesses)
	} else if weaknesses, ok := result["ux_weaknesses"].([]interface{}); ok {
		analysis.Weaknesses = toJSONString(weaknesses)
	} else if weaknesses, ok := result["technical_risks"].([]interface{}); ok {
		analysis.Weaknesses = toJSONString(weaknesses)
	} else if weaknesses, ok := result["financial_risks"].([]interface{}); ok {
		analysis.Weaknesses = toJSONString(weaknesses)
	}

	if opportunities, ok := result["opportunities"].([]interface{}); ok {
		analysis.Opportunities = toJSONString(opportunities)
	} else if opportunities, ok := result["design_opportunities"].([]interface{}); ok {
		analysis.Opportunities = toJSONString(opportunities)
	}

	if threats, ok := result["threats"].([]interface{}); ok {
		analysis.Threats = toJSONString(threats)
	} else if threats, ok := result["competitive_threats"].([]interface{}); ok {
		analysis.Threats = toJSONString(threats)
	}

	if recommendations, ok := result["recommendations"].([]interface{}); ok {
		analysis.Recommendations = toJSONString(recommendations)
	}

	// Store full response as detailed analysis
	analysis.DetailedAnalysis = jsonStr

	return analysis, nil
}

// GenerateExecutionPlan creates an implementation roadmap
func (s *DigService) GenerateExecutionPlan(idea string, analyses []models.DigAnalysis) (*models.DigExecutionPlan, error) {
	// Compile insights from analyses
	var insights strings.Builder
	for _, a := range analyses {
		insights.WriteString(fmt.Sprintf("\n%s Analysis:\n%s\n", a.Perspective, a.Summary))
	}

	prompt := fmt.Sprintf(`Based on the following idea and multi-perspective analysis, create a detailed execution plan.

IDEA: %s

ANALYSIS INSIGHTS: %s

Create an execution plan in this JSON format:
{
  "summary": "Executive summary of the plan",
  "mvp_scope": "What to build for MVP - be specific",
  "full_scope": "Complete product vision",
  "architecture": {
    "type": "monolith/microservices/serverless",
    "components": ["list of main components"],
    "data_flow": "how data moves through the system"
  },
  "tech_stack": {
    "frontend": "specific recommendation",
    "backend": "specific recommendation",
    "database": "specific recommendation",
    "hosting": "specific recommendation",
    "other": ["other tools needed"]
  },
  "integrations": ["third-party services needed"],
  "phases": [
    {
      "name": "Phase 1: MVP",
      "duration": "X weeks",
      "deliverables": ["what gets built"],
      "resources": "team needed"
    }
  ],
  "milestones": [
    {"name": "Milestone 1", "target": "Week X", "criteria": "success criteria"}
  ],
  "team_requirements": {
    "roles": ["roles needed"],
    "skills": ["skills required"],
    "min_team_size": 1
  },
  "estimated_cost": {
    "mvp": "$X-Y",
    "full_product": "$X-Y",
    "monthly_running": "$X-Y"
  },
  "llm_prompt": "A detailed prompt that someone could give to an LLM to help build this - include specific requirements, tech decisions, and context"
}`, idea, insights.String())

	response, err := s.aiService.chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate execution plan: %w", err)
	}

	jsonStr := extractJSON(response)
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &models.DigExecutionPlan{
			Summary: response,
		}, nil
	}

	plan := &models.DigExecutionPlan{}

	if summary, ok := result["summary"].(string); ok {
		plan.Summary = summary
	}
	if mvp, ok := result["mvp_scope"].(string); ok {
		plan.MVPScope = mvp
	}
	if full, ok := result["full_scope"].(string); ok {
		plan.FullScope = full
	}
	if llm, ok := result["llm_prompt"].(string); ok {
		plan.LLMPrompt = llm
	}

	// Store complex objects as JSON
	if arch, ok := result["architecture"]; ok {
		plan.Architecture = toJSONString(arch)
	}
	if stack, ok := result["tech_stack"]; ok {
		plan.TechStack = toJSONString(stack)
	}
	if integrations, ok := result["integrations"]; ok {
		plan.Integrations = toJSONString(integrations)
	}
	if phases, ok := result["phases"]; ok {
		plan.Phases = toJSONString(phases)
	}
	if milestones, ok := result["milestones"]; ok {
		plan.Milestones = toJSONString(milestones)
	}
	if team, ok := result["team_requirements"]; ok {
		plan.TeamRequirements = toJSONString(team)
	}
	if cost, ok := result["estimated_cost"]; ok {
		plan.EstimatedCost = toJSONString(cost)
	}

	return plan, nil
}

// GenerateMarketing creates marketing materials
func (s *DigService) GenerateMarketing(idea string, analyses []models.DigAnalysis) (*models.DigMarketing, error) {
	prompt := fmt.Sprintf(`Create compelling marketing materials for this idea: "%s"

Generate marketing content that would make people excited to use/buy this.
Think like Steve Jobs presenting + Elon Musk tweeting + top copywriters.

Return JSON:
{
  "value_proposition": "The ONE thing that makes this irresistible",
  "target_audience": [
    {"segment": "Primary users", "description": "who they are", "pain_point": "what hurts"}
  ],
  "positioning_statement": "For [target] who [need], [product] is a [category] that [benefit]. Unlike [competitors], we [differentiator].",
  "taglines": ["5 short punchy taglines - think Nike 'Just Do It' level"],
  "elevator_pitch": "30-second pitch that creates urgency",
  "headlines": ["5 attention-grabbing headlines for landing page"],
  "descriptions": [
    {"type": "tweet", "text": "280 char viral tweet"},
    {"type": "short", "text": "One paragraph description"},
    {"type": "long", "text": "Full marketing description"}
  ],
  "color_suggestions": {
    "primary": "#hex with reasoning",
    "secondary": "#hex with reasoning",
    "accent": "#hex with reasoning",
    "mood": "the feeling these colors evoke"
  },
  "imagery_prompts": [
    "DALL-E prompt for hero image",
    "DALL-E prompt for feature illustration",
    "DALL-E prompt for social media"
  ],
  "channel_strategy": {
    "primary_channels": ["where to focus"],
    "tactics": ["specific tactics for each channel"]
  },
  "launch_strategy": "How to launch for maximum impact"
}`, idea)

	response, err := s.aiService.chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate marketing: %w", err)
	}

	jsonStr := extractJSON(response)
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &models.DigMarketing{
			ValueProposition: response,
		}, nil
	}

	marketing := &models.DigMarketing{}

	if vp, ok := result["value_proposition"].(string); ok {
		marketing.ValueProposition = vp
	}
	if pos, ok := result["positioning_statement"].(string); ok {
		marketing.PositioningStatement = pos
	}
	if pitch, ok := result["elevator_pitch"].(string); ok {
		marketing.ElevatorPitch = pitch
	}
	if launch, ok := result["launch_strategy"].(string); ok {
		marketing.LaunchStrategy = launch
	}

	// Arrays and objects as JSON
	if audience, ok := result["target_audience"]; ok {
		marketing.TargetAudience = toJSONString(audience)
	}
	if taglines, ok := result["taglines"]; ok {
		marketing.Taglines = toJSONString(taglines)
	}
	if headlines, ok := result["headlines"]; ok {
		marketing.Headlines = toJSONString(headlines)
	}
	if descriptions, ok := result["descriptions"]; ok {
		marketing.Descriptions = toJSONString(descriptions)
	}
	if colors, ok := result["color_suggestions"]; ok {
		marketing.ColorSuggestions = toJSONString(colors)
	}
	if prompts, ok := result["imagery_prompts"]; ok {
		marketing.ImageryPrompts = toJSONString(prompts)
	}
	if channels, ok := result["channel_strategy"]; ok {
		marketing.ChannelStrategy = toJSONString(channels)
	}

	return marketing, nil
}

// GenerateRevenueProjections creates financial projections
func (s *DigService) GenerateRevenueProjections(idea string, analyses []models.DigAnalysis) (*models.DigRevenueProjection, error) {
	prompt := fmt.Sprintf(`Create realistic revenue projections for this idea: "%s"

Be specific with numbers. Use reasonable assumptions based on similar products.

Return JSON:
{
  "market_size": {
    "tam": "Total Addressable Market with calculation",
    "sam": "Serviceable Addressable Market with calculation",
    "som": "Serviceable Obtainable Market with calculation"
  },
  "market_growth": "Annual growth rate with source/reasoning",
  "competitors": [
    {"name": "Competitor", "revenue": "estimated", "market_share": "X%%", "pricing": "$X"}
  ],
  "pricing_models": [
    {"model": "Freemium/Subscription/etc", "price_points": ["$X/mo", "$Y/mo"], "pros": [], "cons": []}
  ],
  "recommended_price": "$X/month or $Y one-time",
  "pricing_rationale": "Why this pricing makes sense",
  "year1_revenue": {
    "assumptions": "customer growth assumptions",
    "monthly": [0, 100, 500, ...12 months],
    "total": "$X"
  },
  "year2_revenue": {
    "assumptions": "growth assumptions",
    "monthly": [...12 months],
    "total": "$X"
  },
  "year3_revenue": {
    "assumptions": "scale assumptions",
    "total": "$X"
  },
  "key_metrics": {
    "mrr_target": "Monthly recurring revenue goal",
    "churn_target": "Acceptable churn rate",
    "conversion_rate": "Expected free-to-paid",
    "cac": "Target customer acquisition cost",
    "ltv": "Target lifetime value"
  },
  "break_even_analysis": "When and how you'll break even",
  "assumptions": ["Key assumptions made in projections"],
  "risks": ["Financial risks to watch"]
}`, idea)

	response, err := s.aiService.chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate revenue projections: %w", err)
	}

	jsonStr := extractJSON(response)
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &models.DigRevenueProjection{
			MarketSize: response,
		}, nil
	}

	projection := &models.DigRevenueProjection{}

	if growth, ok := result["market_growth"].(string); ok {
		projection.MarketGrowth = growth
	}
	if price, ok := result["recommended_price"].(string); ok {
		projection.RecommendedPrice = price
	}
	if rationale, ok := result["pricing_rationale"].(string); ok {
		projection.PricingRationale = rationale
	}
	if breakEven, ok := result["break_even_analysis"].(string); ok {
		projection.BreakEvenAnalysis = breakEven
	}

	// Store complex objects as JSON
	if market, ok := result["market_size"]; ok {
		projection.MarketSize = toJSONString(market)
	}
	if competitors, ok := result["competitors"]; ok {
		projection.Competitors = toJSONString(competitors)
	}
	if pricing, ok := result["pricing_models"]; ok {
		projection.PricingModels = toJSONString(pricing)
	}
	if y1, ok := result["year1_revenue"]; ok {
		projection.Year1Revenue = toJSONString(y1)
	}
	if y2, ok := result["year2_revenue"]; ok {
		projection.Year2Revenue = toJSONString(y2)
	}
	if y3, ok := result["year3_revenue"]; ok {
		projection.Year3Revenue = toJSONString(y3)
	}
	if metrics, ok := result["key_metrics"]; ok {
		projection.KeyMetrics = toJSONString(metrics)
	}
	if assumptions, ok := result["assumptions"]; ok {
		projection.Assumptions = toJSONString(assumptions)
	}
	if risks, ok := result["risks"]; ok {
		projection.Risks = toJSONString(risks)
	}

	return projection, nil
}

// GenerateAlternatives suggests pivots and refinements
func (s *DigService) GenerateAlternatives(idea string, analyses []models.DigAnalysis) ([]models.DigAlternative, error) {
	// Compile weaknesses from analyses
	var weaknesses strings.Builder
	for _, a := range analyses {
		if a.Weaknesses != "" {
			weaknesses.WriteString(a.Weaknesses + "\n")
		}
	}

	prompt := fmt.Sprintf(`Based on this idea and its identified weaknesses, suggest 3-5 alternative directions or pivots that might be MORE successful.

ORIGINAL IDEA: %s

IDENTIFIED WEAKNESSES: %s

For each alternative, provide JSON array:
[
  {
    "alternative_idea": "Specific alternative or pivot",
    "rationale": "Why this might work better",
    "comparison": "How it addresses original weaknesses",
    "viability_score": 0-100,
    "effort_level": "low/medium/high",
    "potential": "low/medium/high/very_high"
  }
]

Think creatively - adjacent markets, different business models, simpler versions, bigger visions.`, idea, weaknesses.String())

	response, err := s.aiService.chat(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate alternatives: %w", err)
	}

	jsonStr := extractJSON(response)
	var alternatives []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &alternatives); err != nil {
		return nil, nil // Return empty if parsing fails
	}

	var result []models.DigAlternative
	for _, alt := range alternatives {
		alternative := models.DigAlternative{}

		if idea, ok := alt["alternative_idea"].(string); ok {
			alternative.AlternativeIdea = idea
		}
		if rationale, ok := alt["rationale"].(string); ok {
			alternative.Rationale = rationale
		}
		if comparison, ok := alt["comparison"].(string); ok {
			alternative.Comparison = comparison
		}
		if score, ok := alt["viability_score"].(float64); ok {
			alternative.ViabilityScore = &score
		}
		if effort, ok := alt["effort_level"].(string); ok {
			alternative.EffortLevel = effort
		}
		if potential, ok := alt["potential"].(string); ok {
			alternative.Potential = potential
		}

		result = append(result, alternative)
	}

	return result, nil
}

// GenerateTitle creates a title from the idea
func (s *DigService) GenerateTitle(idea string) string {
	// Simple title extraction - first 50 chars or first sentence
	title := idea
	if len(title) > 50 {
		// Find first sentence end
		for i, c := range title {
			if c == '.' || c == '!' || c == '?' {
				if i < 100 {
					title = title[:i]
					break
				}
			}
		}
		if len(title) > 60 {
			title = title[:57] + "..."
		}
	}
	return title
}

// CalculateOverallScore calculates weighted average of all perspective scores
func (s *DigService) CalculateOverallScore(analyses []models.DigAnalysis) float64 {
	weights := map[models.DigPerspective]float64{
		models.DigPerspectiveFirstPrinciples: 0.25,
		models.DigPerspectiveDesigner:        0.20,
		models.DigPerspectiveMarketing:       0.20,
		models.DigPerspectiveTechnical:       0.20,
		models.DigPerspectiveFinancial:       0.15,
	}

	var totalScore float64
	var totalWeight float64

	for _, a := range analyses {
		if a.Score != nil {
			weight := weights[a.Perspective]
			totalScore += *a.Score * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return totalScore / totalWeight
}

// DetermineRecommendation determines the final recommendation based on score
func (s *DigService) DetermineRecommendation(score float64, analyses []models.DigAnalysis) models.DigRecommendation {
	// Check for any critical failures
	for _, a := range analyses {
		if a.Score != nil && *a.Score < 30 {
			// If any perspective scores very low, recommend refinement or pivot
			if score < 50 {
				return models.DigRecommendAbandon
			}
			return models.DigRecommendPivot
		}
	}

	// Based on overall score
	switch {
	case score >= 75:
		return models.DigRecommendProceed
	case score >= 55:
		return models.DigRecommendRefine
	case score >= 40:
		return models.DigRecommendPivot
	default:
		return models.DigRecommendAbandon
	}
}

// GenerateImage generates an image using DALL-E
func (s *DigService) GenerateImage(prompt string) (string, error) {
	if !s.aiService.IsConfigured() {
		return "", fmt.Errorf("AI service not configured")
	}

	requestBody := map[string]interface{}{
		"model":  "dall-e-3",
		"prompt": prompt,
		"n":      1,
		"size":   "1024x1024",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.aiService.baseURL+"/images/generations", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.aiService.apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("DALL-E API error: %s - %s", resp.Status, string(body))
	}

	var response struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Data) == 0 {
		return "", fmt.Errorf("no image generated")
	}

	return response.Data[0].URL, nil
}

// Helper functions

func extractJSON(response string) string {
	// Try to extract JSON from response (might be wrapped in markdown code blocks)
	jsonStr := response

	// Remove markdown code blocks if present
	if idx := strings.Index(response, "```json"); idx != -1 {
		jsonStr = response[idx+7:]
		if endIdx := strings.Index(jsonStr, "```"); endIdx != -1 {
			jsonStr = jsonStr[:endIdx]
		}
	} else if idx := strings.Index(response, "```"); idx != -1 {
		jsonStr = response[idx+3:]
		if endIdx := strings.Index(jsonStr, "```"); endIdx != -1 {
			jsonStr = jsonStr[:endIdx]
		}
	}

	// Find JSON boundaries
	if idx := strings.Index(jsonStr, "{"); idx != -1 {
		jsonStr = jsonStr[idx:]
		// Find matching closing brace
		depth := 0
		for i, c := range jsonStr {
			if c == '{' {
				depth++
			} else if c == '}' {
				depth--
				if depth == 0 {
					jsonStr = jsonStr[:i+1]
					break
				}
			}
		}
	} else if idx := strings.Index(jsonStr, "["); idx != -1 {
		jsonStr = jsonStr[idx:]
		depth := 0
		for i, c := range jsonStr {
			if c == '[' {
				depth++
			} else if c == ']' {
				depth--
				if depth == 0 {
					jsonStr = jsonStr[:i+1]
					break
				}
			}
		}
	}

	return strings.TrimSpace(jsonStr)
}

func toJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
