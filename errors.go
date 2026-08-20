package blogadmin

import "errors"

// Common errors
var (
	// ErrStoreRequired is returned when Store is not provided
	ErrStoreRequired = errors.New("blog store is required")

	// ErrLoggerRequired is returned when Logger is not provided
	ErrLoggerRequired = errors.New("logger is required")

	// ErrAIEnabledMissingLlmFactory is returned when AIEnabled is true
	// but LlmFactory is nil. AI controllers need an LLM engine to run.
	ErrAIEnabledMissingLlmFactory = errors.New("AIEnabled is true but LlmFactory is nil — provide an LlmFactory or set AIEnabled to false")

	// ErrAIEnabledMissingCustomStore is returned when AIEnabled is true
	// but CustomStore is nil. AI controllers persist generated content
	// and approved titles through CustomStore.
	ErrAIEnabledMissingCustomStore = errors.New("AIEnabled is true but CustomStore is nil — provide a CustomStore or set AIEnabled to false")

	// ErrAIEnabledMissingSettingStore is returned when AIEnabled is true
	// but SettingStore is nil. The AI title generator reads its model
	// and prompt settings from SettingStore.
	ErrAIEnabledMissingSettingStore = errors.New("AIEnabled is true but SettingStore is nil — provide a SettingStore or set AIEnabled to false")
)
