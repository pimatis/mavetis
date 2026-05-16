package rule

import "github.com/Pimatis/mavetis/src/model"

func gospecific() []model.Rule {
	return []model.Rule{
		{
			ID:          "go.unsafe.usage",
			Title:       "unsafe package usage introduced",
			Message:     "The diff imports or uses the unsafe package.",
			Remediation: "Avoid unsafe unless absolutely required and document every usage with a security review.",
			Category:    "go",
			Severity:    "medium",
			Confidence:  "high",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(import\s+"unsafe"|unsafe\.Pointer|unsafe\.Sizeof|unsafe\.Alignof|unsafe\.Offsetof)`},
			Standards:   standard("OWASP-ASVS-V6.2", "OWASP-Go"),
		},
		{
			ID:          "go.pprof.exposed",
			Title:       "net/http/pprof exposed without protection",
			Message:     "The diff imports net/http/pprof without authentication or network restrictions.",
			Remediation: "Protect pprof endpoints behind authentication and restrict them to internal networks.",
			Category:    "go",
			Severity:    "high",
			Confidence:  "high",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(import\s+_\s+"net/http/pprof")`},
			Absent:      []string{`(?i)(BasicAuth|middleware|Auth|requireAuth|internal)`},
			Standards:   standard("OWASP-ASVS-V1.14", "OWASP-Go"),
		},
		{
			ID:          "go.server.timeouts",
			Title:       "HTTP server timeouts missing",
			Message:     "The diff creates an http.Server without read, write, or idle timeouts.",
			Remediation: "Set ReadTimeout, WriteTimeout, and IdleTimeout on every http.Server to prevent slowloris and resource exhaustion.",
			Category:    "go",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(&http\.Server\{|http\.Server\{)`},
			Absent:      []string{`(?i)(ReadTimeout|WriteTimeout|IdleTimeout)`},
			Standards:   standard("OWASP-ASVS-V9.1", "OWASP-Go"),
		},
		{
			ID:          "go.json.unmarshal.interface",
			Title:       "json.Unmarshal into interface{} from untrusted input",
			Message:     "The diff unmarshals JSON into an empty interface from request-controlled input.",
			Remediation: "Unmarshal into strict structs and validate every field instead of using interface{}.",
			Category:    "go",
			Severity:    "medium",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)json\.Unmarshal\(.*(interface\{\}|any\b)`},
			Near:        []string{`(?i)(req\.Body|request\.Body|body|payload)`},
			Standards:   standard("OWASP-ASVS-V1.5", "OWASP-Go"),
		},
	}
}
