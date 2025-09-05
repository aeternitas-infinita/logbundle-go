package erri

import "github.com/gofiber/fiber/v2"

type ErriBuilder struct {
	err *Erri
}

func (b *ErriBuilder) Type(errorType ErriType) *ErriBuilder {
	b.err.Type = errorType
	return b
}

func (b *ErriBuilder) Message(message string) *ErriBuilder {
	b.err.Message = message
	return b
}

func (b *ErriBuilder) Details(details string) *ErriBuilder {
	b.err.Details = details
	return b
}

func (b *ErriBuilder) Property(property string) *ErriBuilder {
	b.err.Property = property
	return b
}

func (b *ErriBuilder) Value(value any) *ErriBuilder {
	b.err.Value = value
	return b
}

func (b *ErriBuilder) SystemError(systemError error) *ErriBuilder {
	b.err.SystemError = systemError
	return b
}

func (b *ErriBuilder) Build() *Erri {
	return b.err
}

func extractRequestInfo(c *fiber.Ctx) requestInfo {
	var params map[string]any
	if paramsValue := c.Locals("params"); paramsValue != nil {
		params = map[string]any{
			"params": paramsValue,
		}
	}

	queryParams := make(map[string]any)
	for key, value := range c.Context().QueryArgs().All() {
		queryParams[string(key)] = string(value)
	}

	return requestInfo{
		URL:         c.OriginalURL(),
		Method:      c.Method(),
		Params:      params,
		QueryParams: queryParams,
		Route:       c.Route().Path,
	}
}
