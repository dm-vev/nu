package azureopenai

// Name implements contracts.LLM.Name
func (c *Client) Name() string { return "azure-openai" }

// SupportsStreaming implements contracts.LLM.SupportsStreaming
func (c *Client) SupportsStreaming() bool { return true }

// GetModel returns the model name being used
func (c *Client) GetModel() string { return c.Model }

// GetDeployment returns the deployment name being used
func (c *Client) GetDeployment() string { return c.deployment }

// GetRegion returns the Azure region being used
func (c *Client) GetRegion() string { return c.region }

// GetResourceName returns the Azure resource name being used
func (c *Client) GetResourceName() string { return c.resourceName }

// GetBaseURL returns the base URL being used
func (c *Client) GetBaseURL() string { return c.baseURL }
