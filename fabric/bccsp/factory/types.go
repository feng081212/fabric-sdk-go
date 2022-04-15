package factory

// BccspOpts holds configuration information used to initialize factory implementations
type BccspOpts struct {
	Default string  `json:"default" yaml:"Default"`
	SW      *SwOpts `json:"SW,omitempty" yaml:"SW,omitempty"`
}

// SwOpts contains options for the SWFactory
type SwOpts struct {
	// Default algorithms when not specified (Deprecated?)
	SecLevel      int                `mapstructure:"security" json:"security" yaml:"Security"`
	HashFamily    string             `mapstructure:"hash" json:"hash" yaml:"Hash"`
	ByteKeyStore  *ByteKeyStore      `mapstructure:"bytekeystore,omitempty" json:"bytekeystore,omitempty" yaml:"ByteKeyStore,omitempty"`
	FileKeystore  *FileKeystoreOpts  `mapstructure:"filekeystore,omitempty" json:"filekeystore,omitempty" yaml:"FileKeyStore,omitempty"`
	InMemKeystore *InMemKeystoreOpts `mapstructure:"inmemkeystore,omitempty" json:"inmemkeystore,omitempty"`
}

// FileKeystoreOpts Pluggable Keystores, could add JKS, P12, etc.
type FileKeystoreOpts struct {
	KeyStorePath string `mapstructure:"keystore" yaml:"KeyStore" json:"KeyStore"`
	Password     string `mapstructure:"password" yaml:"password" json:"password"`
}

type ByteKeyStore struct {
	Value    string `mapstructure:"value" yaml:"value" json:"value"`
	Password string `mapstructure:"password" yaml:"password" json:"password"`
}

// InMemKeystoreOpts - empty, as there is no config for the in-memory keystore
type InMemKeystoreOpts struct{}
