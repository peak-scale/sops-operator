package api

// SopsMetadata defines the encryption details
// +kubebuilder:object:generate=true
type SopsMetadata struct {
	// Aws KMS configuration
	//+optional
	AwsKms []KmsDataItem `json:"kms,omitempty"`

	// PGP configuration
	//+optional
	Pgp []PgpDataItem `json:"pgp,omitempty"`

	// Azure KMS configuration
	//+optional
	AzureKms []AzureKmsItem `json:"azure_kv,omitempty"`

	// Hashicorp Vault KMS configurarion
	//+optional
	HcVault []HcVaultItem `json:"hc_vault,omitempty"`

	// Gcp KMS configuration
	//+optional
	GcpKms []GcpKmsDataItem `json:"gcp_kms,omitempty"`

	// Age configuration
	//+optional
	Age []AgeItem `json:"age,omitempty"`

	// Mac - sops setting
	//+optional
	Mac string `json:"mac,omitempty"`

	// LastModified date when SopsSecret was last modified
	//+optional
	LastModified string `json:"lastmodified,omitempty"`

	// Version of the sops tool used to encrypt SopsSecret
	//+optional
	Version string `json:"version,omitempty"`

	// Suffix used to encrypt SopsSecret resource
	//+optional
	EncryptedSuffix string `json:"encrypted_suffix,omitempty"`

	// Regex used to encrypt SopsSecret resource
	// This opstion should be used with more care, as it can make resource unapplicable to the cluster.
	//+optional
	EncryptedRegex string `json:"encrypted_regex,omitempty"`
}

// KmsDataItem defines AWS KMS specific encryption details
type KmsDataItem struct {
	// Arn - KMS key ARN to use
	//+optional
	Arn string `json:"arn,omitempty"`
	// AWS Iam Role
	//+optional
	Role string `json:"role,omitempty"`

	//+optional
	EncryptedKey string `json:"enc,omitempty"`
	// Object creation date
	//+optional
	CreationDate string `json:"created_at,omitempty"`
	//+optional
	AwsProfile string `json:"aws_profile,omitempty"`
}

// PgpDataItem defines PGP specific encryption details
type PgpDataItem struct {
	//+optional
	EncryptedKey string `json:"enc,omitempty"`

	// Object creation date
	//+optional
	CreationDate string `json:"created_at,omitempty"`
	// PGP FingerPrint of the key which can be used for decryption
	//+optional
	FingerPrint string `json:"fp,omitempty"`
}

// AzureKmsItem defines Azure Keyvault Key specific encryption details
type AzureKmsItem struct {
	// Azure KMS vault URL
	//+optional
	VaultURL string `json:"vault_url,omitempty"`
	//+optional
	KeyName string `json:"name,omitempty"`
	//+optional
	Version string `json:"version,omitempty"`
	//+optional
	EncryptedKey string `json:"enc,omitempty"`
	// Object creation date
	//+optional
	CreationDate string `json:"created_at,omitempty"`
}

// AgeItem defines FiloSottile/age specific encryption details
type AgeItem struct {
	// Recipient which private key can be used for decription
	//+optional
	Recipient string `json:"recipient,omitempty"`
	//+optional
	EncryptedKey string `json:"enc,omitempty"`
}

// HcVaultItem defines Hashicorp Vault Key specific encryption details
type HcVaultItem struct {
	//+optional
	VaultAddress string `json:"vault_address,omitempty"`
	//+optional
	EnginePath string `json:"engine_path,omitempty"`
	//+optional
	KeyName string `json:"key_name,omitempty"`
	//+optional
	CreationDate string `json:"created_at,omitempty"`
	//+optional
	EncryptedKey string `json:"enc,omitempty"`
}

// GcpKmsDataItem defines GCP KMS Key specific encryption details
type GcpKmsDataItem struct {
	//+optional
	VaultURL string `json:"resource_id,omitempty"`
	//+optional
	EncryptedKey string `json:"enc,omitempty"`
	// Object creation date
	//+optional
	CreationDate string `json:"created_at,omitempty"`
}
