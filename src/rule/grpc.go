package rule

import "github.com/Pimatis/mavetis/src/model"

func grpc() []model.Rule {
	return []model.Rule{
		{
			ID:          "grpc.tls.missing",
			Title:       "gRPC server without TLS",
			Message:     "The diff starts a gRPC server without TLS or mTLS.",
			Remediation: "Require TLS for all gRPC connections and enforce mutual TLS where feasible.",
			Category:    "transport",
			Severity:    "critical",
			Confidence:  "high",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(grpc\.NewServer\(|grpc\.NewServerWithOptions)`},
			Absent:      []string{`(?i)(credentials\.NewTLS|credentials\.NewServerTLSFromFile|mTLS|tlsConfig|TLSConfig)`},
			Standards:   standard("OWASP-ASVS-V9.1", "OWASP-Transport"),
		},
		{
			ID:          "grpc.reflection.enabled",
			Title:       "gRPC reflection enabled in production",
			Message:     "The diff enables gRPC reflection, which can leak service definitions.",
			Remediation: "Disable gRPC reflection in production or restrict it to authenticated internal access.",
			Category:    "transport",
			Severity:    "medium",
			Confidence:  "high",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(reflection\.Register\(|grpc\.reflection)`},
			Standards:   standard("OWASP-ASVS-V9.1", "OWASP-Transport"),
		},
	}
}
