package rule

import "github.com/Pimatis/mavetis/src/model"

func graphql() []model.Rule {
	return []model.Rule{
		{
			ID:          "inject.graphql.introspection",
			Title:       "GraphQL introspection enabled",
			Message:     "The diff enables GraphQL introspection, which can leak schema details.",
			Remediation: "Disable introspection in production and restrict it to authenticated internal tooling.",
			Category:    "injection",
			Severity:    "medium",
			Confidence:  "high",
			Target:      "added",
			Paths:       codeAndConfigFiles(),
			Require:     []string{`(?i)(introspection\s*[:=]\s*true|__schema|GraphiQL|graphiql)`},
			Standards:   standard("OWASP-ASVS-V5.1", "OWASP-GraphQL"),
		},
		{
			ID:          "inject.graphql.depth",
			Title:       "GraphQL depth limit missing",
			Message:     "The diff introduces a GraphQL endpoint without a query depth limit.",
			Remediation: "Enforce a maximum query depth to prevent deeply nested denial-of-service queries.",
			Category:    "injection",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(graphql|apollo|gql|express-graphql|fastify-gql)`},
			Absent:      []string{`(?i)(maxDepth|depthLimit|depthLimit|queryDepth|maxQueryDepth)`},
			Standards:   standard("OWASP-ASVS-V5.1", "OWASP-GraphQL"),
		},
		{
			ID:          "api.graphql.complexity",
			Title:       "GraphQL complexity limit missing",
			Message:     "The diff introduces a GraphQL endpoint without a complexity limit.",
			Remediation: "Enforce a maximum query complexity to prevent expensive denial-of-service queries.",
			Category:    "logic",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(graphql|apollo|gql|express-graphql|fastify-gql)`},
			Absent:      []string{`(?i)(complexityLimit|maxComplexity|queryComplexity|costAnalysis|costLimit)`},
			Standards:   standard("OWASP-ASVS-V5.1", "OWASP-GraphQL"),
		},
	}
}
