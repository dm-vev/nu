package card

import "github.com/a2aproject/a2a-go/a2a"

// Option configures an AgentCard builder.
type Option func(*Builder)

// WithVersion sets the agent version on the card.
func WithVersion(version string) Option {
	return func(b *Builder) {
		b.version = version
	}
}

// WithProviderInfo sets the provider organization info on the card.
func WithProviderInfo(org, url string) Option {
	return func(b *Builder) {
		b.providerOrg = org
		b.providerURL = url
	}
}

// WithDocumentationURL sets the documentation URL on the card.
func WithDocumentationURL(url string) Option {
	return func(b *Builder) {
		b.documentationURL = url
	}
}

// WithStreaming enables or disables streaming capability on the card.
func WithStreaming(enabled bool) Option {
	return func(b *Builder) {
		b.streaming = enabled
	}
}

// WithInputModes sets the default accepted input MIME types.
func WithInputModes(modes ...string) Option {
	return func(b *Builder) {
		b.inputModes = modes
	}
}

// WithOutputModes sets the default accepted output MIME types.
func WithOutputModes(modes ...string) Option {
	return func(b *Builder) {
		b.outputModes = modes
	}
}

// WithSecurityRequirements sets the card security requirements.
func WithSecurityRequirements(security []a2a.SecurityRequirements) Option {
	return func(b *Builder) {
		b.security = security
	}
}

// WithNamedSecuritySchemes sets the card security schemes.
func WithNamedSecuritySchemes(schemes a2a.NamedSecuritySchemes) Option {
	return func(b *Builder) {
		b.securitySchemes = schemes
	}
}
