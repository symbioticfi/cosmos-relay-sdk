package types

// Slashing module event types
const (
	EventTypeSlash    = "slash"
	EventTypeLiveness = "liveness"

	AttributeKeyAddress        = "address"
	AttributeKeyHeight         = "height"
	AttributeKeyPower          = "power"
	AttributeKeyReason         = "reason"
	AttributeKeyJailed         = "jailed"
	AttributeKeyMissedBlocks   = "missed_blocks"
	AttributeKeySlashRequestID = "slash_request_id"

	AttributeValueUnspecified      = "unspecified"
	AttributeValueDoubleSign       = "double_sign"
	AttributeValueMissingSignature = "missing_signature"
)
